package dota2gc

import (
	"context"
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
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
	StatusRateLimited
)

type Client struct {
	steamClient *steam.Client
	dotaClient  *dota2.Dota2

	username string
	password string
	authCode string

	status          ConnectionStatus
	statusMutex     sync.RWMutex
	statusSetTime   time.Time
	statusTimeMutex sync.RWMutex

	stopChan             chan struct{}
	eventLoopWg          sync.WaitGroup // WaitGroup to track event loop completion
	connectMutex         sync.Mutex     // Mutex to prevent concurrent Connect() calls
	disableAutoReconnect bool           // Flag to disable auto-reconnect (e.g., after auth failures)
	autoReconnectMutex   sync.Mutex     // Mutex for auto-reconnect flag
	sentryPath           string

	lastLogonResult      steamlang.EResult // Track last error to determine auth code type
	lastReconnectAttempt time.Time         // Track last reconnection attempt to prevent rapid reconnects
	reconnectMutex       sync.Mutex        // Mutex for lastReconnectAttempt
	lastConnectionFailed bool              // Track if last connection attempt failed
	lastErrorMessage     string            // Track last error message for user display
	errorMessageMutex    sync.RWMutex      // Mutex for lastErrorMessage
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
	c.status = s
	c.statusMutex.Unlock()

	c.statusTimeMutex.Lock()
	c.statusSetTime = time.Now()
	c.statusTimeMutex.Unlock()
}

func (c *Client) GetLastErrorMessage() string {
	c.errorMessageMutex.RLock()
	defer c.errorMessageMutex.RUnlock()
	return c.lastErrorMessage
}

func (c *Client) ClearErrorMessage() {
	c.errorMessageMutex.Lock()
	defer c.errorMessageMutex.Unlock()
	c.lastErrorMessage = ""
}

func (c *Client) GetStatus() ConnectionStatus {
	c.statusMutex.RLock()
	defer c.statusMutex.RUnlock()
	return c.status
}

func (c *Client) IsStatusStuck() bool {
	c.statusMutex.RLock()
	status := c.status
	c.statusMutex.RUnlock()

	if status != StatusConnecting {
		return false
	}

	c.statusTimeMutex.RLock()
	statusTime := c.statusSetTime
	c.statusTimeMutex.RUnlock()

	return time.Since(statusTime) > 20*time.Second
}

func (c *Client) Connect() error {
	c.connectMutex.Lock()
	defer c.connectMutex.Unlock()

	// Check reconnection cooldown, but allow bypass if:
	// 1. Last connection attempt failed (immediate retry needed)
	// 2. We're in StatusConnecting (initial connection attempt, allow retry)
	currentStatus := c.GetStatus()
	c.reconnectMutex.Lock()
	timeSinceLastReconnect := time.Since(c.lastReconnectAttempt)
	lastFailed := c.lastConnectionFailed
	c.reconnectMutex.Unlock()

	reconnectCooldown := 3 * time.Second
	allowBypass := lastFailed || currentStatus == StatusConnecting

	if !allowBypass && timeSinceLastReconnect < reconnectCooldown {
		log.Printf("Steam: Skipping connection attempt, cooldown period active (last attempt %v ago, need %v)", timeSinceLastReconnect, reconnectCooldown)
		return fmt.Errorf("connection cooldown active")
	}

	c.reconnectMutex.Lock()
	c.lastReconnectAttempt = time.Now()
	c.lastConnectionFailed = false
	c.reconnectMutex.Unlock()

	// Check if already connected or connecting (re-check status after cooldown check)
	currentStatus = c.GetStatus()
	if c.steamClient != nil {
		if currentStatus == StatusConnected || currentStatus == StatusGCReady {
			log.Printf("Steam client already connected (status: %d), skipping", currentStatus)
			return nil
		}
		if currentStatus == StatusConnecting {
			// Check if we've been connecting for too long (stuck)
			if c.IsStatusStuck() {
				log.Printf("Steam client stuck in StatusConnecting, resetting...")
			} else {
				log.Printf("Steam client already connecting (status: %d), skipping duplicate connection attempt", currentStatus)
				return nil
			}
		}
		// If we have a client but it's disconnected, stop the old event loop first
		log.Printf("Stopping old event loop before reconnecting...")
		select {
		case <-c.stopChan:
		default:
			close(c.stopChan)
		}

		// Release lock while waiting for event loop to stop to prevent deadlock
		c.connectMutex.Unlock()

		done := make(chan struct{})
		go func() {
			c.eventLoopWg.Wait()
			close(done)
		}()

		select {
		case <-done:
		case <-time.After(2 * time.Second):
			log.Printf("Timeout waiting for old event loop to stop, proceeding with cleanup")
		}

		// Re-acquire lock for cleanup and new client creation
		c.connectMutex.Lock()

		cleanupDone := make(chan struct{})
		go func() {
			c.cleanup()
			close(cleanupDone)
		}()

		select {
		case <-cleanupDone:
		case <-time.After(3 * time.Second):
			log.Printf("Timeout waiting for cleanup, proceeding anyway")
		}

		c.eventLoopWg = sync.WaitGroup{}
	}

	log.Printf("Creating new Steam client connection...")
	c.steamClient = steam.NewClient()
	c.dotaClient = dota2.New(c.steamClient, logrus.New())

	// Create new stopChan for this connection
	c.stopChan = make(chan struct{})

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

func (c *Client) SubmitCode(code string) error {
	log.Printf("SubmitCode called with code (length: %d)", len(code))

	currentStatus := c.GetStatus()
	c.connectMutex.Lock()
	c.authCode = code
	steamClient := c.steamClient
	c.connectMutex.Unlock()

	didReconnect := false
	if steamClient == nil {
		log.Printf("Steam client not initialized, connecting...")
		if err := c.Connect(); err != nil {
			log.Printf("Failed to connect when submitting code: %v", err)
			return fmt.Errorf("failed to connect: %w", err)
		}
		didReconnect = true
		// Wait a moment for connection to establish
		time.Sleep(500 * time.Millisecond)
		c.connectMutex.Lock()
		steamClient = c.steamClient
		c.connectMutex.Unlock()
		if steamClient == nil {
			return fmt.Errorf("steam client still not initialized after connect attempt")
		}
	}

	if !steamClient.Connected() {
		if currentStatus == StatusNeedGuardCode {
			log.Printf("Steam client disconnected while waiting for guard code, reconnecting...")
			if err := c.Connect(); err != nil {
				log.Printf("Failed to reconnect when submitting code: %v", err)
				return fmt.Errorf("failed to reconnect: %w", err)
			}
			didReconnect = true
			time.Sleep(500 * time.Millisecond)
			c.connectMutex.Lock()
			steamClient = c.steamClient
			c.connectMutex.Unlock()
			if steamClient == nil || !steamClient.Connected() {
				log.Printf("Steam client still not connected after reconnect attempt")
				return fmt.Errorf("connection not established after reconnect")
			}
		} else if currentStatus == StatusConnecting {
			log.Printf("Steam client is connecting, waiting for connection to establish...")
			// Wait for connection to establish instead of reconnecting
			for i := 0; i < 20; i++ {
				time.Sleep(250 * time.Millisecond)
				if steamClient.Connected() {
					break
				}
				currentStatus = c.GetStatus()
				if currentStatus != StatusConnecting && currentStatus != StatusNeedGuardCode {
					break
				}
			}
			if !steamClient.Connected() {
				log.Printf("Steam client still not connected after waiting, will retry login when connected")
				return fmt.Errorf("connection not established yet, please wait")
			}
		} else {
			log.Printf("Steam client not connected, reconnecting...")
			if err := c.Connect(); err != nil {
				log.Printf("Failed to reconnect when submitting code: %v", err)
				return fmt.Errorf("failed to reconnect: %w", err)
			}
			didReconnect = true
			time.Sleep(500 * time.Millisecond)
			c.connectMutex.Lock()
			steamClient = c.steamClient
			c.connectMutex.Unlock()
			if steamClient == nil || !steamClient.Connected() {
				log.Printf("Steam client still not connected after reconnect attempt")
				return fmt.Errorf("connection not established after reconnect")
			}
		}
	}

	if !didReconnect {
		log.Printf("Steam client connected, retrying login with code...")
		c.logOn()
	} else {
		log.Printf("Steam client reconnected, waiting for Connected event to handle login...")
	}
	return nil
}

func (c *Client) logOn() {
	loginDetails := &steam.LogOnDetails{
		Username: c.username,
		Password: c.password,
	}

	if sentryFileContent, err := ioutil.ReadFile(c.sentryPath); err == nil && len(sentryFileContent) > 0 {
		// SentryFileHash should be the SHA1 hash of the sentry file content
		hash := sha1.Sum(sentryFileContent)
		loginDetails.SentryFileHash = steam.SentryHash(hash[:])
		log.Printf("Steam: Using Sentry File (Hash: %x)", hash[:])
	}

	if c.authCode != "" {
		// Determine which code field to populate based on the last error
		// If we don't know (0), default to both or try to guess (usually TwoFactor for mobile, AuthCode for email)
		// But sending both can cause issues, so let's try to be specific if possible.

		useTwoFactor := false
		if c.lastLogonResult == steamlang.EResult_AccountLoginDeniedNeedTwoFactor ||
			c.lastLogonResult == steamlang.EResult_AccountLogonDeniedNeedTwoFactorCode ||
			c.lastLogonResult == steamlang.EResult_TwoFactorCodeMismatch ||
			int32(c.lastLogonResult) == 85 {
			useTwoFactor = true
		}

		if useTwoFactor {
			log.Printf("Steam: Using TwoFactorCode (based on error %v)", c.lastLogonResult)
			loginDetails.TwoFactorCode = c.authCode
		} else {
			log.Printf("Steam: Using AuthCode (based on error %v)", c.lastLogonResult)
			loginDetails.AuthCode = c.authCode
		}
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
				currentStatus := c.GetStatus()
				c.connectMutex.Lock()
				hasAuthCode := c.authCode != ""
				c.connectMutex.Unlock()

				if currentStatus == StatusNeedGuardCode && !hasAuthCode {
					log.Println("Steam: Connected but waiting for guard code, not auto-logging in")
					continue
				}
				c.logOn()

			case *steam.LoggedOnEvent:
				log.Println("Steam: Logged On")
				c.connectMutex.Lock()
				c.authCode = ""
				c.connectMutex.Unlock()
				c.SetStatus(StatusConnected)
				// Mark connection as successful
				c.reconnectMutex.Lock()
				c.lastConnectionFailed = false
				c.reconnectMutex.Unlock()

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
				c.lastLogonResult = e.Result

				// Check for Steam Guard code requirement (both deprecated and current enum names map to same value 85)
				if e.Result == steamlang.EResult_AccountLogonDenied ||
					e.Result == steamlang.EResult_AccountLoginDeniedNeedTwoFactor ||
					e.Result == steamlang.EResult_AccountLogonDeniedNeedTwoFactorCode ||
					int32(e.Result) == 85 { // Numeric check as fallback
					c.connectMutex.Lock()
					hasAuthCode := c.authCode != ""
					if hasAuthCode {
						log.Printf("Steam: Guard code required but code was already provided - clearing expired/wrong code")
						c.authCode = ""
					}
					c.connectMutex.Unlock()
					c.SetStatus(StatusNeedGuardCode)
				} else {
					c.connectMutex.Lock()
					hasAuthCode := c.authCode != ""
					c.connectMutex.Unlock()

					// Check if this is a wrong code error - reset to StatusNeedGuardCode to allow retry
					if e.Result == steamlang.EResult_TwoFactorCodeMismatch ||
						e.Result == steamlang.EResult_InvalidLoginAuthCode {
						log.Printf("Steam: Wrong guard code, clearing and allowing retry")
						c.connectMutex.Lock()
						c.authCode = ""
						c.connectMutex.Unlock()
						c.SetStatus(StatusNeedGuardCode)
						c.autoReconnectMutex.Lock()
						c.disableAutoReconnect = true
						c.autoReconnectMutex.Unlock()
						c.errorMessageMutex.Lock()
						c.lastErrorMessage = "Wrong Steam Guard code. Please try again."
						c.errorMessageMutex.Unlock()
					} else if e.Result == steamlang.EResult_InvalidPassword {
						if hasAuthCode {
							log.Printf("Steam: InvalidPassword error but auth code was provided - treating as wrong code, allowing retry")
							c.connectMutex.Lock()
							c.authCode = ""
							c.connectMutex.Unlock()
							c.SetStatus(StatusNeedGuardCode)
							c.autoReconnectMutex.Lock()
							c.disableAutoReconnect = true
							c.autoReconnectMutex.Unlock()
						} else {
							log.Printf("Steam: InvalidPassword error (no auth code provided) - Steam may require 2FA, allowing code entry")
							c.SetStatus(StatusNeedGuardCode)
							c.autoReconnectMutex.Lock()
							c.disableAutoReconnect = true
							c.autoReconnectMutex.Unlock()
						}
					} else if e.Result == steamlang.EResult_RateLimitExceeded {
						log.Printf("Steam: Rate limit exceeded (E84) - Too many login attempts. Please wait at least 24 hours before retrying.")
						c.SetStatus(StatusRateLimited)
						c.autoReconnectMutex.Lock()
						c.disableAutoReconnect = true
						c.autoReconnectMutex.Unlock()
					} else if e.Result == steamlang.EResult_AccountLoginDeniedThrottle {
						log.Printf("Steam: Account login denied due to throttling (E87) - Too many failed login attempts. Please wait at least 24 hours before retrying.")
						c.SetStatus(StatusRateLimited)
						c.autoReconnectMutex.Lock()
						c.disableAutoReconnect = true
						c.autoReconnectMutex.Unlock()
					} else {
						c.SetStatus(StatusDisconnected)
						c.autoReconnectMutex.Lock()
						c.disableAutoReconnect = true
						c.autoReconnectMutex.Unlock()
					}
				}

			case *steam.MachineAuthUpdateEvent:
				log.Printf("Steam: Machine Auth Update (Hash: %x, Bytes: %d)", e.Hash, len(e.Bytes))
				if len(e.Bytes) > 0 {
					if err := ioutil.WriteFile(c.sentryPath, e.Bytes, 0600); err != nil {
						log.Printf("Steam: Failed to save Sentry File: %v", err)
					} else {
						log.Printf("Steam: Sentry File saved successfully to %s", c.sentryPath)
					}
				} else {
					log.Printf("Steam: Warning - MachineAuthUpdateEvent received with empty Bytes!")
				}

			case *events.ClientStateChanged:
				if e.NewState.ConnectionStatus == protocol.GCConnectionStatus_GCConnectionStatus_HAVE_SESSION {
					log.Println("Dota 2 GC: Ready")
					c.SetStatus(StatusGCReady)
				}

			case *steam.DisconnectedEvent:
				log.Println("Steam: Disconnected")
				currentStatus := c.GetStatus()

				// If we are waiting for a guard code, don't reset status or auto-reconnect loop
				// The user needs to submit the code, which will trigger a reconnect if needed.
				if currentStatus == StatusNeedGuardCode {
					log.Println("Steam: Disconnected while waiting for Guard Code. Waiting for user input.")
					continue
				}

				// During initial connection, transient disconnections can occur - be more lenient
				if currentStatus == StatusConnecting {
					log.Printf("Steam: Disconnected during connection establishment (status: %d), this may be transient", currentStatus)
					// Don't immediately reset status - give it a moment to recover
					// The connection timeout watchdog will handle if it's truly stuck
					continue
				}

				// Only set to disconnected if we were actually connected
				if currentStatus == StatusConnected || currentStatus == StatusGCReady {
					c.SetStatus(StatusDisconnected)
				} else if currentStatus == StatusDisconnected {
					// Already disconnected, don't trigger another reconnect
					continue
				} else {
					c.SetStatus(StatusDisconnected)
				}

				// Check if auto-reconnect is disabled (e.g., after auth failures)
				c.autoReconnectMutex.Lock()
				shouldAutoReconnect := !c.disableAutoReconnect
				c.autoReconnectMutex.Unlock()

				if !shouldAutoReconnect {
					log.Println("Steam: Auto-reconnect disabled (likely due to authentication failure). Waiting for manual reconnect.")
					continue
				}

				// Check reconnection cooldown to prevent rapid reconnection loops
				c.reconnectMutex.Lock()
				timeSinceLastReconnect := time.Since(c.lastReconnectAttempt)
				c.reconnectMutex.Unlock()

				reconnectCooldown := 5 * time.Second
				if timeSinceLastReconnect < reconnectCooldown {
					log.Printf("Steam: Skipping auto-reconnect, cooldown period active (last attempt %v ago, need %v)", timeSinceLastReconnect, reconnectCooldown)
					continue
				}

				// Attempt to reconnect automatically (but only if not shutting down)
				select {
				case <-c.stopChan:
					// Shutting down, don't reconnect
					return
				default:
					c.reconnectMutex.Lock()
					c.lastReconnectAttempt = time.Now()
					c.reconnectMutex.Unlock()

					go func() {
						time.Sleep(5 * time.Second)
						// Check again if we should reconnect
						select {
						case <-c.stopChan:
							return
						default:
							// Double-check auto-reconnect flag and status
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
				currentStatus := c.GetStatus()
				errorStr := e.Error()
				isEOF := strings.Contains(errorStr, "EOF") || strings.Contains(errorStr, "connection reset")

				if currentStatus == StatusConnecting {
					// Mark connection as failed to allow immediate retry
					c.reconnectMutex.Lock()
					c.lastConnectionFailed = true
					c.reconnectMutex.Unlock()

					c.connectMutex.Lock()
					hasAuthCode := c.authCode != ""
					c.connectMutex.Unlock()

					if isEOF && hasAuthCode {
						// EOF during guard code submission - set error message and status
						c.errorMessageMutex.Lock()
						c.lastErrorMessage = "Connection error during code submission. Please try submitting the code again."
						c.errorMessageMutex.Unlock()
						log.Printf("Steam: EOF error during guard code submission, setting status to NeedGuardCode")
						c.SetStatus(StatusNeedGuardCode)
						// Don't auto-retry immediately for EOF during code submission - let user retry
					} else if hasAuthCode {
						log.Printf("Steam: Connection error while connecting with guard code, auto-retrying immediately")
						go func() {
							select {
							case <-c.stopChan:
								return
							default:
								time.Sleep(2 * time.Second)
								log.Printf("Steam: Auto-retrying connection with guard code after error")
								if err := c.Connect(); err != nil {
									log.Printf("Steam: Auto-retry connection failed: %v", err)
									c.SetStatus(StatusNeedGuardCode)
								}
							}
						}()
					} else {
						log.Printf("Steam: Connection error while connecting (status: %d), auto-retrying immediately", currentStatus)
						go func() {
							select {
							case <-c.stopChan:
								log.Printf("Steam: Stop signal received, skipping auto-retry after error")
								return
							default:
								time.Sleep(2 * time.Second)
								log.Printf("Steam: Auto-retrying connection after error")
								c.autoReconnectMutex.Lock()
								c.disableAutoReconnect = false
								c.autoReconnectMutex.Unlock()
								if err := c.Connect(); err != nil {
									log.Printf("Steam: Auto-retry connection failed: %v", err)
								}
							}
						}()
					}
				}

			case *steam.AccountInfoEvent:
			case *steam.PersonaStateEvent:
			case *steam.FriendsListEvent:
			case *steam.ClanStateEvent:
			case *events.GCConnectionStatusChanged:
			case *events.ClientWelcomed:
				log.Println("Dota 2 GC: Welcomed (forcing Ready state)")
				c.SetStatus(StatusGCReady)
			default:
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
		log.Printf("[GetReplayInfo] GC is Connected but not Ready, waiting for GC session...")
		if c.dotaClient != nil {
			go c.dotaClient.SayHello()
		}
		for i := 0; i < 10; i++ { // Wait up to 2.5 seconds
			time.Sleep(250 * time.Millisecond)
			status = c.GetStatus()
			if status == StatusGCReady {
				log.Printf("[GetReplayInfo] GC is now Ready")
				break
			}
		}
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

func (c *Client) GetPlayerConductScorecard() (*protocol.CMsgPlayerConductScorecard, error) {
	status := c.GetStatus()
	if status != StatusGCReady && status != StatusConnected {
		return nil, fmt.Errorf("GC not ready (Status: %d)", status)
	}
	if status == StatusConnected {
		log.Printf("[GetPlayerConductScorecard] GC is Connected but not Ready, waiting for GC session...")
		if c.dotaClient != nil {
			go c.dotaClient.SayHello()
		}
		for i := 0; i < 10; i++ {
			time.Sleep(250 * time.Millisecond)
			status = c.GetStatus()
			if status == StatusGCReady {
				log.Printf("[GetPlayerConductScorecard] GC is now Ready")
				break
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	res, err := c.dotaClient.RequestLatestConductScorecard(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to request conduct scorecard: %w", err)
	}

	return res, nil
}

type Match struct {
	ID        int64  `json:"id"`
	GameMode  uint32 `json:"gameMode"`
	LobbyType uint32 `json:"lobbyType"`
	StartTime uint32 `json:"startTime"`
}

func (c *Client) GetPlayerMatchHistory(steamID64 int64, limit int, turboOnly bool) ([]Match, error) {
	return c.GetPlayerMatchHistoryPaginated(steamID64, limit, turboOnly, 0)
}

func (c *Client) GetPlayerMatchHistoryPaginated(steamID64 int64, limit int, turboOnly bool, startAtMatchID uint64) ([]Match, error) {
	status := c.GetStatus()
	if status != StatusGCReady && status != StatusConnected {
		return nil, fmt.Errorf("GC not ready (Status: %d)", status)
	}

	if status == StatusConnected {
		log.Printf("[GetPlayerMatchHistory] GC is Connected but not Ready, waiting for GC session...")
		// Force SayHello to kickstart session if needed
		if c.dotaClient != nil {
			go c.dotaClient.SayHello()
		}

		for i := 0; i < 20; i++ { // Wait up to 5 seconds
			time.Sleep(250 * time.Millisecond)
			status = c.GetStatus()
			if status == StatusGCReady {
				log.Printf("[GetPlayerMatchHistory] GC is now Ready")
				break
			}
			// Every 1 second (4 iterations), try saying hello again if still not ready
			if i > 0 && i%4 == 0 && c.dotaClient != nil {
				go c.dotaClient.SayHello()
			}
		}
		if status != StatusGCReady {
			log.Printf("[GetPlayerMatchHistory] GC still not Ready (status: %d) after waiting, proceeding anyway", status)
		}
	}

	accountID := uint32(convertSteamID(uint64(steamID64), false))
	matchesRequested := uint32(limit)
	if matchesRequested > 20 {
		matchesRequested = 20
	}
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
	if startAtMatchID > 0 {
		req.StartAtMatchId = &startAtMatchID
	}

	timeout := 20 * time.Second
	if limit > 20 || startAtMatchID == 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	log.Printf("[GetPlayerMatchHistory] Requesting %d matches (max 20 per request) with timeout %v, startAtMatchID=%d", limit, timeout, startAtMatchID)
	resp, err := c.dotaClient.GetPlayerMatchHistory(ctx, req)
	if err != nil {
		log.Printf("[GetPlayerMatchHistory] ERROR: %v", err)
		if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "timeout") {
			log.Printf("[GetPlayerMatchHistory] Timeout detected, refreshing GC connection...")
			if c.dotaClient != nil {
				c.SetStatus(StatusConnected)
				c.dotaClient.SayHello()
			}
		}
		return nil, fmt.Errorf("failed to get player match history: %w", err)
	}
	log.Printf("[GetPlayerMatchHistory] Successfully got response with %d matches", len(resp.GetMatches()))

	var matches []Match
	for _, m := range resp.GetMatches() {
		if turboOnly {
			if m.GetGameMode() != 23 {
				continue
			}
		}
		matches = append(matches, Match{
			ID:        int64(m.GetMatchId()),
			GameMode:  m.GetGameMode(),
			LobbyType: m.GetLobbyType(),
			StartTime: m.GetStartTime(),
		})
	}

	return matches, nil
}

const (
	GameModeSingleDraft = 4
	LobbyTypeRanked     = 7
)

func checkIfEnoughMatches(matches []Match, maxDepth int) bool {
	singleDraftCount := 0
	rankedAfterSDCount := 0
	currentDepth := 0
	i := 0

	for currentDepth < maxDepth && i < len(matches) {
		foundSingleDraft := false
		singleDraftIndex := -1

		for i < len(matches) {
			if matches[i].GameMode == GameModeSingleDraft {
				foundSingleDraft = true
				singleDraftIndex = i
				singleDraftCount++
				break
			}
			i++
		}

		if !foundSingleDraft {
			return false
		}

		foundRanked := false
		for j := singleDraftIndex + 1; j < len(matches); j++ {
			if matches[j].LobbyType == LobbyTypeRanked {
				rankedAfterSDCount++
				foundRanked = true
				i = j + 1
				break
			}
		}

		if !foundRanked {
			return false
		}

		currentDepth++
	}

	return currentDepth >= maxDepth
}

type FatalMatchInfo struct {
	FatalMatchID       int64   `json:"fatalMatchId"`
	SingleDraftMatchID int64   `json:"singleDraftMatchId"`
	SingleDraftDate    uint32  `json:"singleDraftDate"`
	AdditionalMatchIDs []int64 `json:"additionalMatchIds"`
}

func (c *Client) FindFatalGames(steamID64 int64, maxDepth int, gamesPerFatal int) ([]FatalMatchInfo, error) {
	log.Printf("[FindFatalGames] Starting: steamID64=%d, maxDepth=%d, gamesPerFatal=%d", steamID64, maxDepth, gamesPerFatal)
	if maxDepth < 1 {
		return nil, fmt.Errorf("maxDepth must be at least 1")
	}

	status := c.GetStatus()
	log.Printf("[FindFatalGames] GC status: %d", status)
	if status != StatusGCReady && status != StatusConnected {
		return nil, fmt.Errorf("GC not ready (Status: %d)", status)
	}

	var fatalMatches []FatalMatchInfo
	batchSize := 20
	maxBatches := 30

	log.Printf("[FindFatalGames] Fetching match history incrementally: batchSize=%d, maxBatches=%d", batchSize, maxBatches)
	var allMatches []Match
	var startAtMatchID uint64 = 0
	batchesFetched := 0

	for batchesFetched < maxBatches {
		if batchesFetched > 0 {
			time.Sleep(500 * time.Millisecond)
			status := c.GetStatus()
			if status != StatusGCReady && status != StatusConnected {
				log.Printf("[FindFatalGames] GC status degraded to %d, refreshing connection...", status)
				if c.dotaClient != nil {
					c.dotaClient.SayHello()
				}
				time.Sleep(1 * time.Second)
				status = c.GetStatus()
				if status != StatusGCReady && status != StatusConnected {
					log.Printf("[FindFatalGames] GC still not ready after refresh (status: %d), but continuing", status)
				}
			} else if c.dotaClient != nil {
				go c.dotaClient.SayHello()
			}
		}

		log.Printf("[FindFatalGames] Fetching batch %d/%d: %d matches, startAtMatchID=%d", batchesFetched+1, maxBatches, batchSize, startAtMatchID)
		matches, err := c.GetPlayerMatchHistoryPaginated(steamID64, batchSize, false, startAtMatchID)
		if err != nil {
			log.Printf("[FindFatalGames] ERROR getting match history batch: %v", err)
			if strings.Contains(err.Error(), "context deadline exceeded") || strings.Contains(err.Error(), "timeout") {
				log.Printf("[FindFatalGames] Timeout detected, refreshing GC connection and retrying...")
				if c.dotaClient != nil {
					c.dotaClient.SayHello()
				}
				time.Sleep(2 * time.Second)
				matches, err = c.GetPlayerMatchHistoryPaginated(steamID64, batchSize, false, startAtMatchID)
				if err != nil {
					log.Printf("[FindFatalGames] Retry also failed: %v", err)
					if len(allMatches) == 0 {
						return nil, fmt.Errorf("failed to get match history: %w", err)
					}
					log.Printf("[FindFatalGames] Using %d matches from previous batches", len(allMatches))
					break
				}
				log.Printf("[FindFatalGames] Retry succeeded after connection refresh")
			} else {
				if len(allMatches) == 0 {
					return nil, fmt.Errorf("failed to get match history: %w", err)
				}
				log.Printf("[FindFatalGames] Using %d matches from previous batches", len(allMatches))
				break
			}
		}

		if len(matches) == 0 {
			log.Printf("[FindFatalGames] No more matches available")
			break
		}

		allMatches = append(allMatches, matches...)
		batchesFetched++
		log.Printf("[FindFatalGames] Batch %d complete: got %d matches, total so far: %d", batchesFetched, len(matches), len(allMatches))

		if len(matches) < batchSize {
			log.Printf("[FindFatalGames] Received fewer matches than requested, no more available")
			break
		}

		startAtMatchID = uint64(matches[len(matches)-1].ID)

		if len(allMatches) >= 30 {
			log.Printf("[FindFatalGames] Have %d matches, checking if we can find fatal games", len(allMatches))
			if foundEnough := checkIfEnoughMatches(allMatches, maxDepth); foundEnough {
				// Only stop early if we have enough matches to potentially find gamesPerFatal ranked games before singledraft
				// We need at least singledraft position + gamesPerFatal matches to ensure we can find enough ranked games
				// Since matches are in reverse chronological order, we need to have fetched enough to cover the range
				if len(allMatches) >= 50 || gamesPerFatal == 0 {
					log.Printf("[FindFatalGames] Found enough matches to complete search, stopping early")
					break
				}
				log.Printf("[FindFatalGames] Have enough for fatal search but may need more for gamesPerFatal=%d, continuing...", gamesPerFatal)
			}
		}
	}

	matches := allMatches
	log.Printf("[FindFatalGames] Got %d total matches from history (fetched in %d batches)", len(matches), batchesFetched)
	if len(matches) == 0 {
		log.Printf("[FindFatalGames] No matches found, returning empty")
		return fatalMatches, nil
	}

	currentDepth := 0
	i := 0

	log.Printf("[FindFatalGames] Starting search loop: maxDepth=%d, totalMatches=%d", maxDepth, len(matches))
	for currentDepth < maxDepth && i < len(matches) {
		log.Printf("[FindFatalGames] Depth %d/%d: starting from index %d", currentDepth+1, maxDepth, i)
		foundSingleDraft := false
		singleDraftIndex := -1

		for i < len(matches) {
			m := matches[i]
			if m.GameMode == GameModeSingleDraft {
				if m.LobbyType == LobbyTypeRanked {
					log.Printf("[FindFatalGames] ERROR: Invalid match data at index %d: match %d is single draft but ranked", i, m.ID)
					return nil, fmt.Errorf("invalid match data: single draft game cannot be ranked (match %d)", m.ID)
				}
				foundSingleDraft = true
				singleDraftIndex = i
				log.Printf("[FindFatalGames] Found single draft game at index %d: match %d (gameMode=%d, lobbyType=%d)", i, m.ID, m.GameMode, m.LobbyType)
				break
			}
			i++
		}

		if !foundSingleDraft {
			log.Printf("[FindFatalGames] No more single draft games found at depth %d", currentDepth+1)
			break
		}

		foundRanked := false
		rankedIndex := -1
		log.Printf("[FindFatalGames] Searching for ranked game after single draft (starting from index %d)", singleDraftIndex+1)
		for j := singleDraftIndex + 1; j < len(matches); j++ {
			m := matches[j]
			if m.LobbyType == LobbyTypeRanked {
				log.Printf("[FindFatalGames] Found ranked game at index %d: match %d (gameMode=%d, lobbyType=%d)", j, m.ID, m.GameMode, m.LobbyType)
				singleDraftMatch := matches[singleDraftIndex]

				// Find additional ranked games BEFORE singledraft (matches are in reverse chronological order, newest first)
				// So we need to look at indices BEFORE singleDraftIndex (which are older games)
				var additionalIDs []int64
				if gamesPerFatal > 0 {
					// Start from the match right before singledraft and go backwards (to older games)
					for k := singleDraftIndex - 1; k >= 0 && len(additionalIDs) < gamesPerFatal; k-- {
						if matches[k].LobbyType == LobbyTypeRanked {
							additionalIDs = append(additionalIDs, matches[k].ID)
						}
					}
					log.Printf("[FindFatalGames] Found %d ranked games before singledraft (need %d)", len(additionalIDs), gamesPerFatal)
				}

				fatalMatches = append(fatalMatches, FatalMatchInfo{
					FatalMatchID:       m.ID,
					SingleDraftMatchID: singleDraftMatch.ID,
					SingleDraftDate:    singleDraftMatch.StartTime,
					AdditionalMatchIDs: additionalIDs,
				})
				foundRanked = true
				rankedIndex = j
				break
			}
		}

		if !foundRanked {
			log.Printf("[FindFatalGames] No ranked game found after single draft at depth %d", currentDepth+1)
			break
		}

		currentDepth++
		i = rankedIndex + 1
		log.Printf("[FindFatalGames] Completed depth %d: found fatal match %d (singleDraft: %d), continuing from index %d", currentDepth, fatalMatches[len(fatalMatches)-1].FatalMatchID, fatalMatches[len(fatalMatches)-1].SingleDraftMatchID, i)
	}

	log.Printf("[FindFatalGames] Search complete: found %d fatal matches", len(fatalMatches))
	return fatalMatches, nil
}

func (c *Client) connectionTimeoutWatchdog() {
	timeout := 30 * time.Second
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	for {
		select {
		case <-timer.C:
			currentStatus := c.GetStatus()
			if currentStatus != StatusConnecting {
				log.Printf("Steam: Connection timeout watchdog fired but status is no longer Connecting (status: %d)", currentStatus)
				return
			}

			c.connectMutex.Lock()
			steamClient := c.steamClient
			hasAuthCode := c.authCode != ""
			c.connectMutex.Unlock()

			if steamClient != nil && steamClient.Connected() {
				log.Printf("Steam: Connection timeout watchdog fired but TCP connection is established, extending timeout")
				timer.Reset(30 * time.Second)
				continue
			}

			log.Printf("Steam: Connection timeout watchdog fired after %v (status: %d, TCP connected: %v)", timeout, currentStatus, steamClient != nil && steamClient.Connected())

			// Mark connection as failed to allow immediate retry
			c.reconnectMutex.Lock()
			c.lastConnectionFailed = true
			timeSinceLastReconnect := time.Since(c.lastReconnectAttempt)
			c.reconnectMutex.Unlock()

			reconnectCooldown := 5 * time.Second
			if timeSinceLastReconnect < reconnectCooldown {
				log.Printf("Steam: Skipping timeout retry, cooldown period active (last attempt %v ago)", timeSinceLastReconnect)
				timer.Reset(10 * time.Second)
				continue
			}

			if hasAuthCode {
				log.Printf("Steam: Connection timeout after %v while waiting for guard code, auto-retrying", timeout)
				go func() {
					select {
					case <-c.stopChan:
						return
					default:
						log.Printf("Steam: Auto-retrying connection with guard code after timeout")
						c.reconnectMutex.Lock()
						c.lastReconnectAttempt = time.Now()
						c.reconnectMutex.Unlock()
						if err := c.Connect(); err != nil {
							log.Printf("Steam: Auto-retry connection failed: %v", err)
							c.SetStatus(StatusNeedGuardCode)
						}
					}
				}()
			} else {
				log.Printf("Steam: Connection timeout after %v, auto-retrying", timeout)
				go func() {
					select {
					case <-c.stopChan:
						log.Printf("Steam: Stop signal received, skipping auto-retry after timeout")
						return
					default:
						log.Printf("Steam: Auto-retrying connection after timeout")
						c.reconnectMutex.Lock()
						c.lastReconnectAttempt = time.Now()
						c.reconnectMutex.Unlock()
						c.autoReconnectMutex.Lock()
						c.disableAutoReconnect = false
						c.autoReconnectMutex.Unlock()
						if err := c.Connect(); err != nil {
							log.Printf("Steam: Auto-retry connection failed: %v", err)
						}
					}
				}()
			}
			return
		case <-ticker.C:
			currentStatus := c.GetStatus()
			if currentStatus != StatusConnecting {
				log.Printf("Steam: Connection timeout watchdog cancelled early (status changed to %d)", currentStatus)
				return
			}
		case <-c.stopChan:
			log.Printf("Steam: Connection timeout watchdog stopped (stopChan closed)")
			return
		}
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
