package dota2gc

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/paralin/go-dota2"
	"github.com/paralin/go-dota2/events"
	"github.com/paralin/go-dota2/protocol"
	"github.com/paralin/go-steam"
	"github.com/paralin/go-steam/protocol/steamlang"
	"github.com/sirupsen/logrus"
)

type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusNeedGuardCode
	StatusConnected
	StatusGCReady
)

type Client struct {
	steamClient *steam.Client
	dotaClient  *dota2.Dota2

	username string
	password string
	authCode string

	status      ConnectionStatus
	statusMutex sync.RWMutex

	stopChan             chan struct{}
	eventLoopWg          sync.WaitGroup // WaitGroup to track event loop completion
	connectMutex         sync.Mutex     // Mutex to prevent concurrent Connect() calls
	disableAutoReconnect bool           // Flag to disable auto-reconnect (e.g., after auth failures)
	autoReconnectMutex   sync.Mutex     // Mutex for auto-reconnect flag
	sentryPath           string
}

func NewClient(username, password string) *Client {
	logrus.SetLevel(logrus.DebugLevel)

	home, _ := os.UserHomeDir()
	sentryPath := filepath.Join(home, ".dota-report-timestamps", "sentry.bin")
	os.MkdirAll(filepath.Dir(sentryPath), 0755)

	return &Client{
		username:   username,
		password:   password,
		status:     StatusDisconnected,
		stopChan:   make(chan struct{}),
		sentryPath: sentryPath,
	}
}

func (c *Client) SetStatus(s ConnectionStatus) {
	c.statusMutex.Lock()
	defer c.statusMutex.Unlock()
	c.status = s
}

func (c *Client) GetStatus() ConnectionStatus {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.status
}

func (c *Client) Connect() error {
	c.connectMutex.Lock()
	defer c.connectMutex.Unlock()

	// Check if already connected or connecting
	currentStatus := c.GetStatus()
	if c.steamClient != nil {
		if currentStatus == StatusConnecting || currentStatus == StatusConnected || currentStatus == StatusGCReady {
			log.Printf("Steam client already connecting/connected (status: %d), skipping", currentStatus)
			return nil
		}
		// If we have a client but it's disconnected, clean it up first
		log.Printf("Cleaning up disconnected client before reconnecting...")
		c.cleanup()
	}

	log.Printf("Creating new Steam client connection...")
	c.steamClient = steam.NewClient()
	c.dotaClient = dota2.New(c.steamClient, logrus.New())

	// Create new stopChan for this connection
	c.stopChan = make(chan struct{})

	// Clear auth code for fresh connection attempt
	c.authCode = ""

	// Re-enable auto-reconnect when manually connecting
	c.autoReconnectMutex.Lock()
	c.disableAutoReconnect = false
	c.autoReconnectMutex.Unlock()

	c.SetStatus(StatusConnecting)

	// Start event loop
	c.eventLoopWg.Add(1)
	go c.eventLoop()

	// Start connection timeout watchdog
	go c.connectionTimeoutWatchdog()

	// Start connection (non-blocking)
	c.steamClient.Connect()

	return nil
}

func (c *Client) SubmitCode(code string) {
	log.Printf("SubmitCode called with code (length: %d)", len(code))

	c.connectMutex.Lock()
	c.authCode = code
	steamClient := c.steamClient
	needsConnect := (steamClient == nil)
	c.connectMutex.Unlock()

	if needsConnect {
		// Client not initialized, need to connect first
		log.Printf("Steam client not initialized, connecting...")
		if err := c.Connect(); err != nil {
			log.Printf("Failed to connect when submitting code: %v", err)
			return
		}
		// The Connect() call will trigger ConnectedEvent which calls logOn() automatically
		// So we don't need to call logOn() here - it will be called by the event loop
		return
	}

	// Re-check connection status after lock release
	c.connectMutex.Lock()
	steamClient = c.steamClient
	c.connectMutex.Unlock()

	if steamClient == nil {
		log.Printf("Steam client became nil, connecting...")
		if err := c.Connect(); err != nil {
			log.Printf("Failed to connect when submitting code: %v", err)
			return
		}
		return
	}

	// If not connected, try to reconnect (this will trigger ConnectedEvent -> logOn automatically)
	if !steamClient.Connected() {
		log.Printf("Steam client not connected, reconnecting...")
		c.SetStatus(StatusConnecting)
		steamClient.Connect()
		// The Connect() will trigger ConnectedEvent which calls logOn() automatically
		return
	}

	// Client is connected, retry login with the code
	log.Printf("Steam client connected, retrying login with code...")
	c.logOn()
}

func (c *Client) logOn() {
	loginDetails := &steam.LogOnDetails{
		Username: c.username,
		Password: c.password,
	}

	if sentry, err := ioutil.ReadFile(c.sentryPath); err == nil && len(sentry) > 0 {
		loginDetails.SentryFileHash = steam.SentryHash(sentry)
	}

	if c.authCode != "" {
		loginDetails.AuthCode = c.authCode
		loginDetails.TwoFactorCode = c.authCode
	}

	c.steamClient.Auth.LogOn(loginDetails)
}

func (c *Client) eventLoop() {
	defer c.eventLoopWg.Done()

	for {
		select {
		case event := <-c.steamClient.Events():
			if event == nil {
				// Channel closed, exit
				log.Println("Steam: Event channel closed, exiting event loop")
				return
			}

			switch e := event.(type) {
			case *steam.ConnectedEvent:
				log.Println("Steam: Connected")
				c.logOn()

			case *steam.LoggedOnEvent:
				log.Println("Steam: Logged On")
				c.SetStatus(StatusConnected)

				// Launch in goroutine to not block event loop
				go func() {
					c.steamClient.Social.SetPersonaState(steamlang.EPersonaState_Online)
					time.Sleep(2 * time.Second)
					c.dotaClient.SetPlaying(true)
					time.Sleep(1 * time.Second)
					c.dotaClient.SayHello()
				}()

			case *steam.LogOnFailedEvent:
				log.Printf("Steam: LogOn Failed: %v", e.Result)
				// Check for Steam Guard code requirement (both deprecated and current enum names map to same value 85)
				if e.Result == steamlang.EResult_AccountLogonDenied ||
					e.Result == steamlang.EResult_AccountLoginDeniedNeedTwoFactor ||
					e.Result == steamlang.EResult_AccountLogonDeniedNeedTwoFactorCode ||
					int32(e.Result) == 85 { // Numeric check as fallback
					c.SetStatus(StatusNeedGuardCode)
				} else {
					c.SetStatus(StatusDisconnected)
					// Disable auto-reconnect on authentication failures (wrong password, wrong code, etc.)
					// These require user intervention
					if e.Result == steamlang.EResult_InvalidPassword ||
						e.Result == steamlang.EResult_TwoFactorCodeMismatch ||
						e.Result == steamlang.EResult_InvalidLoginAuthCode {
						log.Printf("Steam: Authentication failure detected, disabling auto-reconnect")
						c.autoReconnectMutex.Lock()
						c.disableAutoReconnect = true
						c.autoReconnectMutex.Unlock()
					}
				}

			case *steam.MachineAuthUpdateEvent:
				log.Println("Steam: Machine Auth Update")
				ioutil.WriteFile(c.sentryPath, e.Hash, 0600)

			case *events.ClientStateChanged:
				if e.NewState.ConnectionStatus == protocol.GCConnectionStatus_GCConnectionStatus_HAVE_SESSION {
					log.Println("Dota 2 GC: Ready")
					c.SetStatus(StatusGCReady)
				}

			case *steam.DisconnectedEvent:
				log.Println("Steam: Disconnected")

				// If we are waiting for a guard code, don't reset status or auto-reconnect loop
				// The user needs to submit the code, which will trigger a reconnect if needed.
				if c.GetStatus() == StatusNeedGuardCode {
					log.Println("Steam: Disconnected while waiting for Guard Code. Waiting for user input.")
					continue
				}

				c.SetStatus(StatusDisconnected)

				// Check if auto-reconnect is disabled (e.g., after auth failures)
				c.autoReconnectMutex.Lock()
				shouldAutoReconnect := !c.disableAutoReconnect
				c.autoReconnectMutex.Unlock()

				if !shouldAutoReconnect {
					log.Println("Steam: Auto-reconnect disabled (likely due to authentication failure). Waiting for manual reconnect.")
					continue
				}

				// Attempt to reconnect automatically (but only if not shutting down)
				select {
				case <-c.stopChan:
					// Shutting down, don't reconnect
					return
				default:
					go func() {
						time.Sleep(10 * time.Second)
						// Check again if we should reconnect
						select {
						case <-c.stopChan:
							return
						default:
							// Double-check auto-reconnect flag
							c.autoReconnectMutex.Lock()
							shouldReconnect := !c.disableAutoReconnect
							c.autoReconnectMutex.Unlock()

							if shouldReconnect && c.GetStatus() == StatusDisconnected {
								log.Println("Steam: Auto-reconnecting...")
								if err := c.Connect(); err != nil {
									log.Printf("Steam: Auto-reconnect failed: %v", err)
								}
							}
						}
					}()
				}

			case error:
				log.Printf("Steam/Dota Error: %v", e)
				if c.GetStatus() == StatusConnecting {
					log.Printf("Steam: Connection error while connecting, resetting to disconnected")
					c.SetStatus(StatusDisconnected)
				}
			}

		case <-c.stopChan:
			log.Println("Steam: Stop signal received, exiting event loop")
			return
		}
	}
}

const SteamID64Identifier = 76561197960265728

func convertSteamID(input uint64, to64 bool) uint64 {
	if input == 0 {
		return 0
	}
	is64 := input >= SteamID64Identifier
	if to64 {
		if is64 {
			return input
		}
		return input + SteamID64Identifier
	} else {
		if !is64 {
			return input
		}
		return input - SteamID64Identifier
	}
}

func (c *Client) GetReplayInfo(matchID uint64) (uint32, uint64, error) {
	status := c.GetStatus()
	if status != StatusGCReady && status != StatusConnected {
		return 0, 0, fmt.Errorf("GC not ready (Status: %d)", status)
	}
	if status == StatusConnected {
		// Relaxed condition: Allow requests when Connected (but not GCReady yet)
		// log.Println("Warning: Requesting replay info while GC is CONNECTED but not READY. This might fail.")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// RequestMatchDetails returns (*protocol.CMsgGCMatchDetailsResponse, error)
	res, err := c.dotaClient.RequestMatchDetails(ctx, matchID)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to request match details: %w", err)
	}

	if res.GetResult() != uint32(steamlang.EResult_OK) {
		return 0, 0, fmt.Errorf("GC returned error result: %v", res.GetResult())
	}

	return res.GetMatch().GetCluster(), uint64(res.GetMatch().GetReplaySalt()), nil
}

type Match struct {
	ID int64 `json:"id"`
}

func (c *Client) GetPlayerMatchHistory(steamID64 int64, limit int, turboOnly bool) ([]Match, error) {
	status := c.GetStatus()
	if status != StatusGCReady && status != StatusConnected {
		return nil, fmt.Errorf("GC not ready (Status: %d)", status)
	}

	accountID := uint32(convertSteamID(uint64(steamID64), false))
	matchesRequested := uint32(limit)
	includePractice := false
	includeCustom := false
	includeEvent := false

	req := &protocol.CMsgDOTAGetPlayerMatchHistory{
		AccountId:              &accountID,
		MatchesRequested:       &matchesRequested,
		IncludePracticeMatches: &includePractice,
		IncludeCustomGames:     &includeCustom,
		IncludeEventGames:      &includeEvent,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := c.dotaClient.GetPlayerMatchHistory(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get player match history: %w", err)
	}

	var matches []Match
	for _, m := range resp.GetMatches() {
		if turboOnly {
			if m.GetGameMode() != 23 {
				continue
			}
		}
		matches = append(matches, Match{
			ID: int64(m.GetMatchId()),
		})
	}

	return matches, nil
}

func (c *Client) connectionTimeoutWatchdog() {
	timeout := 30 * time.Second
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case <-timer.C:
		if c.GetStatus() == StatusConnecting {
			log.Printf("Steam: Connection timeout after %v, resetting to disconnected", timeout)
			c.SetStatus(StatusDisconnected)
		}
	case <-c.stopChan:
		return
	}
}

// cleanup performs cleanup without waiting (internal use)
func (c *Client) cleanup() {
	if c.dotaClient != nil {
		c.dotaClient.Close()
		c.dotaClient = nil
	}
	if c.steamClient != nil {
		c.steamClient.Disconnect()
		c.steamClient = nil
	}
}

// Close properly shuts down the client and waits for cleanup
func (c *Client) Close() {
	c.connectMutex.Lock()
	defer c.connectMutex.Unlock()

	log.Println("Steam: Closing client...")
	c.SetStatus(StatusDisconnected)
	c.authCode = ""

	// Reset auto-reconnect flag
	c.autoReconnectMutex.Lock()
	c.disableAutoReconnect = false
	c.autoReconnectMutex.Unlock()

	// Signal event loop to stop
	select {
	case <-c.stopChan:
		// Already closed
		log.Println("Steam: Stop channel already closed")
	default:
		close(c.stopChan)
	}

	// Cleanup clients
	c.cleanup()

	// Wait for event loop to finish (with timeout)
	done := make(chan struct{})
	go func() {
		c.eventLoopWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("Steam: Event loop stopped")
	case <-time.After(5 * time.Second):
		log.Println("Steam: Timeout waiting for event loop to stop")
	}
}
