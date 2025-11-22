package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/d3nd3/dota-report-timestamps/pkg/botclient"
	"github.com/d3nd3/dota-report-timestamps/pkg/downloader"
	"github.com/d3nd3/dota-report-timestamps/pkg/parser"
	"github.com/d3nd3/dota-report-timestamps/pkg/steamapi"
	// "github.com/d3nd3/dota-report-timestamps/pkg/stratz" // DEPRECATED: Stratz API no longer used
)

var (
	validateReportCardLocks sync.Map // Map[uint64]*sync.Mutex to prevent concurrent validate report card requests for the same match ID
	validateReportCardInProgress sync.Map // Map[uint64]bool to track in-progress requests
)

// convertSteamID ensures we have the correct format.
// If input < 76561197960265728, assumes it's AccountID (32-bit) and converts to SteamID64 if to64 is true.
// If input >= 76561197960265728, assumes it's SteamID64 and converts to AccountID (32-bit) if to64 is false.
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
		// to 32-bit
		if !is64 {
			return input
		}
		return input - SteamID64Identifier
	}
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		json.NewEncoder(w).Encode(config)
	} else if r.Method == http.MethodPost {
		var newConfig Config
		if err := json.NewDecoder(r.Body).Decode(&newConfig); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if newConfig.ReplayDir != "" {
			config.ReplayDir = newConfig.ReplayDir
		}
		if newConfig.StratzAPIToken != "" {
			config.StratzAPIToken = strings.TrimSpace(newConfig.StratzAPIToken)
			log.Printf("Stratz API Token updated (length: %d)", len(config.StratzAPIToken))
		}
		if newConfig.SteamAPIKey != "" {
			config.SteamAPIKey = strings.TrimSpace(newConfig.SteamAPIKey)
			log.Printf("Steam API Key updated (length: %d)", len(config.SteamAPIKey))
		}
		if newConfig.SteamUser != "" {
			config.SteamUser = strings.TrimSpace(newConfig.SteamUser)
			log.Printf("Steam User updated: %s", config.SteamUser)
		}
		if newConfig.SteamPass != "" {
			config.SteamPass = strings.TrimSpace(newConfig.SteamPass)
			log.Printf("Steam Password updated (length: %d)", len(config.SteamPass))
		}
		json.NewEncoder(w).Encode(config)
	}
}

type SteamLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Code     string `json:"code"`
}

func handleSteamLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SteamLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.Username != "" {
		config.SteamUser = req.Username
	}
	if req.Password != "" {
		config.SteamPass = req.Password
	}

	// Always try to ensure bot is initialized if we have creds
	if config.SteamUser == "" || config.SteamPass == "" {
		http.Error(w, "Username and Password required", http.StatusBadRequest)
		return
	}

	// If we are disconnected or providing new creds, Init the bot
	// If we are just sending a code, we assume bot is running but waiting
	// BUT submitting a code requires calling SubmitCode, not Init.

	if req.Code != "" {
		log.Printf("Submitting Steam Guard code (length: %d)", len(req.Code))
		if err := gcClient.SubmitCode(req.Code); err != nil {
			log.Printf("Failed to submit code: %v", err)
			// Return error but also return current status so frontend can update UI
			status := gcClient.GetStatus()
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"status":  int(status),
				"error":   "Failed to submit code: " + err.Error(),
			})
			return
		}
		// Give it a moment to process
		time.Sleep(500 * time.Millisecond)
	} else {
		// Standard login request
		// Check current status first - if already connected, return current status
		// Allow resetting StatusConnecting to handle stuck connections
		// StatusNeedGuardCode is NOT included here - we allow re-init to start fresh connection
		currentStatus := gcClient.GetStatus()
		if currentStatus == botclient.StatusConnected || currentStatus == botclient.StatusGCReady {
			log.Printf("Steam client already in state %d, returning current status", currentStatus)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"status":  int(currentStatus),
				"message": "Already connected",
			})
			return
		}
		
		if currentStatus == botclient.StatusConnecting {
			log.Printf("Steam client in StatusConnecting, allowing reset to handle stuck connection")
		}

		// Init if disconnected, StatusNeedGuardCode, or StatusConnecting (allows reset)
		if err := gcClient.Init(config.SteamUser, config.SteamPass); err != nil {
			log.Printf("Failed to initialize Steam client: %v", err)
			http.Error(w, "Failed to connect: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Give it a moment to start connecting
		time.Sleep(500 * time.Millisecond)
	}

	// Poll status a few times to get the most up-to-date state
	status := gcClient.GetStatus()
	for i := 0; i < 3 && (status == botclient.StatusDisconnected || status == botclient.StatusConnecting); i++ {
		time.Sleep(200 * time.Millisecond)
		status = gcClient.GetStatus()
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"status":  int(status),
	})
}

func handleSteamDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if gcClient == nil {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Already disconnected",
		})
		return
	}

	log.Printf("Disconnecting Steam client...")
	if err := gcClient.Disconnect(); err != nil {
		log.Printf("Failed to disconnect: %v", err)
		http.Error(w, "Failed to disconnect: "+err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Steam client disconnected successfully")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Disconnected successfully",
	})
}

func handleSteamStatus(w http.ResponseWriter, r *http.Request) {
	status := botclient.StatusDisconnected
	var errorMessage string
	if gcClient != nil {
		status, errorMessage = gcClient.GetStatusWithError()
	}

	statusText := "Disconnected"
	switch status {
	case botclient.StatusConnecting:
		statusText = "Connecting"
	case botclient.StatusNeedGuardCode:
		statusText = "Need Steam Guard Code"
	case botclient.StatusConnected:
		statusText = "Connected to Steam"
	case botclient.StatusGCReady:
		statusText = "Dota 2 GC Ready"
	case botclient.StatusRateLimited:
		statusText = "Rate Limited (Wait 24h)"
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":      int(status),
		"statusText":  statusText,
		"errorMessage": errorMessage,
	})
}

func handleConductScorecard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if gcClient == nil {
		http.Error(w, "GC client not available", http.StatusInternalServerError)
		return
	}

	status := gcClient.GetStatus()
	if status != botclient.StatusGCReady && status != botclient.StatusConnected {
		http.Error(w, "GC not ready. Please connect to Steam first.", http.StatusBadRequest)
		return
	}

	scorecard, err := gcClient.GetPlayerConductScorecard()
	if err != nil {
		log.Printf("Error fetching conduct scorecard: %v", err)
		http.Error(w, fmt.Sprintf("Error fetching conduct scorecard: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(scorecard)
}

type ValidateReportCardRequest struct {
	MatchID   uint64 `json:"matchId"`
	SteamID64 int64  `json:"steamId64,omitempty"`
	AccountID uint32 `json:"accountId,omitempty"`
}

func handleValidateReportCard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidateReportCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.MatchID == 0 {
		http.Error(w, "Invalid match ID", http.StatusBadRequest)
		return
	}

	if gcClient == nil {
		http.Error(w, "GC client not available", http.StatusInternalServerError)
		return
	}

	status := gcClient.GetStatus()
	if status != botclient.StatusGCReady && status != botclient.StatusConnected {
		http.Error(w, "GC not ready. Please connect to Steam first.", http.StatusBadRequest)
		return
	}

	var steamID64 int64
	if req.SteamID64 > 0 {
		steamID64 = req.SteamID64
	} else if req.AccountID > 0 {
		steamID64 = int64(convertSteamID(uint64(req.AccountID), true))
	} else {
		http.Error(w, "Steam ID or Account ID required", http.StatusBadRequest)
		return
	}

	lockInterface, _ := validateReportCardLocks.LoadOrStore(req.MatchID, &sync.Mutex{})
	lock := lockInterface.(*sync.Mutex)
	lock.Lock()
	
	if _, inProgress := validateReportCardInProgress.Load(req.MatchID); inProgress {
		lock.Unlock()
		log.Printf("[ValidateReportCard] Request for match %d already in progress, returning early", req.MatchID)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"downloaded": []int64{},
			"skipped":   []int64{},
			"errors":    []string{},
			"total":     0,
			"message":   "Request already in progress for this match ID",
		})
		return
	}
	
	validateReportCardInProgress.Store(req.MatchID, true)
	defer func() {
		validateReportCardInProgress.Delete(req.MatchID)
		lock.Unlock()
	}()

	log.Printf("[ValidateReportCard] Starting validation for match %d, steamID %d", req.MatchID, steamID64)

	reportCardsDir := filepath.Join(config.ReplayDir, "reportcards", fmt.Sprintf("%d", req.MatchID))
	if err := os.MkdirAll(reportCardsDir, os.ModePerm); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create reportcards directory: %v", err), http.StatusInternalServerError)
		return
	}

	const (
		LobbyTypeRanked     = 7
		GameModeSingleDraft = 4
		GameModeTurbo       = 23
		targetGames         = 15
	)

	var rankedMatches []int64
	batchSize := 50
	startAtMatchID := uint64(0)
	batchesFetched := 0
	maxBatches := 5
	foundStartMatch := false

	for len(rankedMatches) < targetGames && batchesFetched < maxBatches {
		if batchesFetched > 0 {
			time.Sleep(500 * time.Millisecond)
		}

		log.Printf("[ValidateReportCard] Fetching batch %d: startAtMatchID=%d, looking for match %d and older games", batchesFetched+1, startAtMatchID, req.MatchID)
		matches, err := gcClient.GetPlayerMatchHistoryPaginated(steamID64, batchSize, false, startAtMatchID)
		if err != nil {
			log.Printf("[ValidateReportCard] Error fetching match history: %v", err)
			http.Error(w, fmt.Sprintf("Error fetching match history: %v", err), http.StatusInternalServerError)
			return
		}

		if len(matches) == 0 {
			log.Printf("[ValidateReportCard] No more matches available")
			break
		}

		for i, m := range matches {
			if m.ID == int64(req.MatchID) {
				foundStartMatch = true
				if m.LobbyType == LobbyTypeRanked && m.GameMode != GameModeSingleDraft && m.GameMode != GameModeTurbo {
					rankedMatches = append(rankedMatches, int64(m.ID))
					log.Printf("[ValidateReportCard] Added start match %d (ranked, non-singledraft, non-turbo)", m.ID)
				}

				for j := i + 1; j < len(matches) && len(rankedMatches) < targetGames; j++ {
					m2 := matches[j]
					if m2.LobbyType == LobbyTypeRanked && m2.GameMode != GameModeSingleDraft && m2.GameMode != GameModeTurbo {
						rankedMatches = append(rankedMatches, int64(m2.ID))
						log.Printf("[ValidateReportCard] Added match %d (ranked, non-singledraft, non-turbo)", m2.ID)
					}
				}
				break
			}
		}

		if foundStartMatch {
			if len(rankedMatches) >= targetGames {
				break
			}
			if len(matches) < batchSize {
				log.Printf("[ValidateReportCard] No more matches available (batch size: %d)", len(matches))
				break
			}
			startAtMatchID = uint64(matches[len(matches)-1].ID)
			batchesFetched++
			continue
		}

		if len(matches) < batchSize {
			log.Printf("[ValidateReportCard] No more matches available (batch size: %d), start match not found", len(matches))
			break
		}

		startAtMatchID = uint64(matches[len(matches)-1].ID)
		batchesFetched++
	}

	if !foundStartMatch {
		log.Printf("[ValidateReportCard] Start match %d not found in match history", req.MatchID)
		http.Error(w, fmt.Sprintf("Start match %d not found in match history", req.MatchID), http.StatusBadRequest)
		return
	}

	if len(rankedMatches) == 0 {
		http.Error(w, "No ranked matches found starting from the specified match ID", http.StatusBadRequest)
		return
	}

	log.Printf("[ValidateReportCard] Found %d ranked matches to download", len(rankedMatches))

	var downloaded []int64
	var skipped []int64
	var errors []string

	for _, matchID := range rankedMatches {
		filePath := filepath.Join(reportCardsDir, fmt.Sprintf("%d.dem", matchID))
		if _, err := os.Stat(filePath); err == nil {
			log.Printf("[ValidateReportCard] Match %d already exists, skipping", matchID)
			skipped = append(skipped, matchID)
			continue
		}

		lockInterface, _ := downloadLocks.LoadOrStore(matchID, &sync.Mutex{})
		lock := lockInterface.(*sync.Mutex)
		lock.Lock()

		if _, err := os.Stat(filePath); err == nil {
			log.Printf("[ValidateReportCard] Match %d was downloaded by another request, skipping", matchID)
			skipped = append(skipped, matchID)
			lock.Unlock()
			continue
		}

		log.Printf("[ValidateReportCard] Downloading match %d", matchID)
		if err := downloader.DownloadReplay(matchID, reportCardsDir, config.StratzAPIToken, config.SteamAPIKey, gcClient); err != nil {
			lock.Unlock()
			if strings.Contains(err.Error(), "queued for parsing") {
				log.Printf("[ValidateReportCard] Match %d queued for parsing", matchID)
				errors = append(errors, fmt.Sprintf("Match %d queued for parsing", matchID))
			} else {
				log.Printf("[ValidateReportCard] Error downloading match %d: %v", matchID, err)
				errors = append(errors, fmt.Sprintf("Match %d: %v", matchID, err))
			}
			continue
		}

		lock.Unlock()

		if _, err := os.Stat(filePath); err == nil {
			downloaded = append(downloaded, matchID)
			log.Printf("[ValidateReportCard] Successfully downloaded match %d", matchID)
		} else {
			log.Printf("[ValidateReportCard] Match %d download reported success but file not found, may be queued", matchID)
			errors = append(errors, fmt.Sprintf("Match %d queued for parsing", matchID))
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"downloaded": downloaded,
		"skipped":   skipped,
		"errors":    errors,
		"total":     len(rankedMatches),
		"directory": reportCardsDir,
	})
}

func handleValidateReportCardCurrent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ValidateReportCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.MatchID == 0 {
		http.Error(w, "Invalid match ID", http.StatusBadRequest)
		return
	}

	if gcClient == nil {
		http.Error(w, "GC client not available", http.StatusInternalServerError)
		return
	}

	status := gcClient.GetStatus()
	if status != botclient.StatusGCReady && status != botclient.StatusConnected {
		http.Error(w, "GC not ready. Please connect to Steam first.", http.StatusBadRequest)
		return
	}

	var steamID64 int64
	if req.SteamID64 > 0 {
		steamID64 = req.SteamID64
	} else if req.AccountID > 0 {
		steamID64 = int64(convertSteamID(uint64(req.AccountID), true))
	} else {
		http.Error(w, "Steam ID or Account ID required", http.StatusBadRequest)
		return
	}

	currentKey := req.MatchID + 10000000000
	lockInterface, _ := validateReportCardLocks.LoadOrStore(currentKey, &sync.Mutex{})
	lock := lockInterface.(*sync.Mutex)
	lock.Lock()
	
	if _, inProgress := validateReportCardInProgress.Load(currentKey); inProgress {
		lock.Unlock()
		log.Printf("[ValidateReportCardCurrent] Request for match %d (current) already in progress, returning early", req.MatchID)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"downloaded": []int64{},
			"skipped":   []int64{},
			"errors":    []string{},
			"total":     0,
			"message":   "Request already in progress for this match ID",
		})
		return
	}
	
	validateReportCardInProgress.Store(currentKey, true)
	defer func() {
		validateReportCardInProgress.Delete(currentKey)
		lock.Unlock()
	}()

	log.Printf("[ValidateReportCardCurrent] Starting validation for games after match %d, steamID %d", req.MatchID, steamID64)

	reportCardsCurrentDir := filepath.Join(config.ReplayDir, "reportcards", "current")
	if err := os.MkdirAll(reportCardsCurrentDir, os.ModePerm); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create reportcards/current directory: %v", err), http.StatusInternalServerError)
		return
	}

	const (
		LobbyTypeRanked     = 7
		GameModeSingleDraft = 4
		GameModeTurbo       = 23
		maxGames            = 15
	)

	var rankedMatches []int64
	batchSize := 50
	startAtMatchID := uint64(0)
	batchesFetched := 0
	maxBatches := 3
	foundStartMatch := false

	for len(rankedMatches) < maxGames && batchesFetched < maxBatches {
		if batchesFetched > 0 {
			time.Sleep(500 * time.Millisecond)
		}

		log.Printf("[ValidateReportCardCurrent] Fetching batch %d: startAtMatchID=%d, looking for games after match %d", batchesFetched+1, startAtMatchID, req.MatchID)
		matches, err := gcClient.GetPlayerMatchHistoryPaginated(steamID64, batchSize, false, startAtMatchID)
		if err != nil {
			log.Printf("[ValidateReportCardCurrent] Error fetching match history: %v", err)
			http.Error(w, fmt.Sprintf("Error fetching match history: %v", err), http.StatusInternalServerError)
			return
		}

		if len(matches) == 0 {
			log.Printf("[ValidateReportCardCurrent] No more matches available")
			break
		}

		for i, m := range matches {
			if m.ID == int64(req.MatchID) {
				foundStartMatch = true
				for j := i - 1; j >= 0 && len(rankedMatches) < maxGames; j-- {
					m2 := matches[j]
					if m2.LobbyType == LobbyTypeRanked && m2.GameMode != GameModeSingleDraft && m2.GameMode != GameModeTurbo {
						rankedMatches = append(rankedMatches, int64(m2.ID))
						log.Printf("[ValidateReportCardCurrent] Added match %d (ranked, non-singledraft, non-turbo)", m2.ID)
					}
				}
				break
			}
		}

		if foundStartMatch {
			break
		}

		if len(matches) < batchSize {
			log.Printf("[ValidateReportCardCurrent] No more matches available (batch size: %d)", len(matches))
			break
		}

		startAtMatchID = uint64(matches[len(matches)-1].ID)
		batchesFetched++
	}

	if !foundStartMatch {
		log.Printf("[ValidateReportCardCurrent] Start match %d not found in match history", req.MatchID)
		http.Error(w, fmt.Sprintf("Start match %d not found in match history", req.MatchID), http.StatusBadRequest)
		return
	}

	if len(rankedMatches) == 0 {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"downloaded": []int64{},
			"skipped":   []int64{},
			"errors":    []string{},
			"total":     0,
			"message":   "No ranked games found after the specified match ID",
			"directory": reportCardsCurrentDir,
		})
		return
	}

	if len(rankedMatches) >= maxGames {
		log.Printf("[ValidateReportCardCurrent] Found %d games (>= %d), not downloading", len(rankedMatches), maxGames)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   true,
			"downloaded": []int64{},
			"skipped":   []int64{},
			"errors":    []string{},
			"total":     len(rankedMatches),
			"message":   fmt.Sprintf("Found %d games (>= %d), not downloading", len(rankedMatches), maxGames),
			"directory": reportCardsCurrentDir,
		})
		return
	}

	log.Printf("[ValidateReportCardCurrent] Found %d ranked matches to download (less than %d)", len(rankedMatches), maxGames)

	var downloaded []int64
	var skipped []int64
	var errors []string

	for _, matchID := range rankedMatches {
		filePath := filepath.Join(reportCardsCurrentDir, fmt.Sprintf("%d.dem", matchID))
		if _, err := os.Stat(filePath); err == nil {
			log.Printf("[ValidateReportCardCurrent] Match %d already exists, skipping", matchID)
			skipped = append(skipped, matchID)
			continue
		}

		lockInterface, _ := downloadLocks.LoadOrStore(matchID, &sync.Mutex{})
		lock := lockInterface.(*sync.Mutex)
		lock.Lock()

		if _, err := os.Stat(filePath); err == nil {
			log.Printf("[ValidateReportCardCurrent] Match %d was downloaded by another request, skipping", matchID)
			skipped = append(skipped, matchID)
			lock.Unlock()
			continue
		}

		log.Printf("[ValidateReportCardCurrent] Downloading match %d", matchID)
		if err := downloader.DownloadReplay(matchID, reportCardsCurrentDir, config.StratzAPIToken, config.SteamAPIKey, gcClient); err != nil {
			lock.Unlock()
			if strings.Contains(err.Error(), "queued for parsing") {
				log.Printf("[ValidateReportCardCurrent] Match %d queued for parsing", matchID)
				errors = append(errors, fmt.Sprintf("Match %d queued for parsing", matchID))
			} else {
				log.Printf("[ValidateReportCardCurrent] Error downloading match %d: %v", matchID, err)
				errors = append(errors, fmt.Sprintf("Match %d: %v", matchID, err))
			}
			continue
		}

		lock.Unlock()

		if _, err := os.Stat(filePath); err == nil {
			downloaded = append(downloaded, matchID)
			log.Printf("[ValidateReportCardCurrent] Successfully downloaded match %d", matchID)
		} else {
			log.Printf("[ValidateReportCardCurrent] Match %d download reported success but file not found, may be queued", matchID)
			errors = append(errors, fmt.Sprintf("Match %d queued for parsing", matchID))
		}
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"downloaded": downloaded,
		"skipped":   skipped,
		"errors":    errors,
		"total":     len(rankedMatches),
		"directory": reportCardsCurrentDir,
	})
}

type ReplayInfo struct {
	FileName string    `json:"fileName"`
	Date     time.Time `json:"date"`
}

func getProfileReplayDir(profileName string) string {
	if profileName == "" {
		return config.ReplayDir
	}
	return filepath.Join(config.ReplayDir, sanitizeFileName(profileName))
}

func getFatalReplayDir(profileName string) string {
	return filepath.Join(getProfileReplayDir(profileName), "fatal")
}

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "..", "_")
	return name
}

type BrowseItem struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	IsDir    bool      `json:"isDir"`
	IsFile   bool      `json:"isFile"`
	Date     time.Time `json:"date,omitempty"`
	FileSize int64     `json:"fileSize,omitempty"`
}

func handleBrowse(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	profileName := r.URL.Query().Get("profile")
	subPath := r.URL.Query().Get("path")
	if subPath != "" {
		subPath = strings.TrimPrefix(subPath, "/")
	}

	baseDir := getProfileReplayDir(profileName)
	targetDir := baseDir
	if subPath != "" {
		targetDir = filepath.Join(baseDir, subPath)
		if !strings.HasPrefix(targetDir, baseDir) {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
	}

	files, err := ioutil.ReadDir(targetDir)
	if err != nil {
		if os.IsNotExist(err) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]BrowseItem{})
			return
		}
		http.Error(w, "Could not read directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	items := []BrowseItem{}
	for _, file := range files {
		item := BrowseItem{
			Name:   file.Name(),
			Path:   filepath.Join(subPath, file.Name()),
			IsDir:  file.IsDir(),
			IsFile: !file.IsDir(),
		}

		if item.IsFile && strings.HasSuffix(file.Name(), ".dem") {
			item.Date = file.ModTime()
			item.FileSize = file.Size()
			filePath := filepath.Join(targetDir, file.Name())
			replayFile, err := os.Open(filePath)
			if err == nil {
				if date, err := parser.GetReplayDate(replayFile); err == nil {
					item.Date = date
				}
				replayFile.Close()
			}
		} else if item.IsDir {
			item.Date = file.ModTime()
		}

		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir
		}
		if !items[i].Date.IsZero() && !items[j].Date.IsZero() {
			if items[i].Date.After(items[j].Date) {
				return true
			}
			if items[i].Date.Before(items[j].Date) {
				return false
			}
		}
		return items[i].Name < items[j].Name
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func handleReplays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	profileName := r.URL.Query().Get("profile")
	fatalMode := r.URL.Query().Get("fatal") == "true"
	
	var replayDir string
	if fatalMode {
		replayDir = getFatalReplayDir(profileName)
	} else {
		replayDir = getProfileReplayDir(profileName)
	}

	replays := []ReplayInfo{}
	
	if fatalMode {
		files, err := ioutil.ReadDir(replayDir)
		if err != nil {
			if os.IsNotExist(err) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode([]ReplayInfo{})
				return
			}
			http.Error(w, "Could not read replay directory: "+err.Error(), http.StatusInternalServerError)
			return
		}

		for _, file := range files {
			if file.IsDir() {
				dateSubDir := filepath.Join(replayDir, file.Name())
				subFiles, err := ioutil.ReadDir(dateSubDir)
				if err != nil {
					continue
				}
				for _, subFile := range subFiles {
					if !subFile.IsDir() && strings.HasSuffix(subFile.Name(), ".dem") {
						replayDate := subFile.ModTime()
						filePath := filepath.Join(dateSubDir, subFile.Name())
						replayFile, err := os.Open(filePath)
						if err == nil {
							if date, err := parser.GetReplayDate(replayFile); err == nil {
								replayDate = date
							}
							replayFile.Close()
						}
						replays = append(replays, ReplayInfo{
							FileName: filepath.Join(file.Name(), subFile.Name()),
							Date:     replayDate,
						})
					}
				}
			} else if strings.HasSuffix(file.Name(), ".dem") {
				replayDate := file.ModTime()
				filePath := filepath.Join(replayDir, file.Name())
				replayFile, err := os.Open(filePath)
				if err == nil {
					if date, err := parser.GetReplayDate(replayFile); err == nil {
						replayDate = date
					}
					replayFile.Close()
				}
				replays = append(replays, ReplayInfo{
					FileName: file.Name(),
					Date:     replayDate,
				})
			}
		}
	} else {
		files, err := ioutil.ReadDir(replayDir)
		if err != nil {
			if os.IsNotExist(err) {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode([]ReplayInfo{})
				return
			}
			http.Error(w, "Could not read replay directory: "+err.Error(), http.StatusInternalServerError)
			return
		}

		for _, file := range files {
			if !file.IsDir() && strings.HasSuffix(file.Name(), ".dem") {
				replayDate := file.ModTime()
				filePath := filepath.Join(replayDir, file.Name())
				replayFile, err := os.Open(filePath)
				if err == nil {
					if date, err := parser.GetReplayDate(replayFile); err == nil {
						replayDate = date
					}
					replayFile.Close()
				}
				replays = append(replays, ReplayInfo{
					FileName: file.Name(),
					Date:     replayDate,
				})
			}
		}
	}

	sort.Slice(replays, func(i, j int) bool {
		return replays[i].Date.After(replays[j].Date)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(replays)
}

type DeleteRequest struct {
	MatchID     string `json:"matchId"`
	FilePath    string `json:"filePath"`
	ProfileName string `json:"profileName"`
}

func handleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	replayDir := getProfileReplayDir(req.ProfileName)
	var filePath string
	
	if req.FilePath != "" {
		filePath = filepath.Join(replayDir, req.FilePath)
	} else if req.MatchID != "" {
		fileName := req.MatchID
		if !strings.HasSuffix(fileName, ".dem") {
			fileName = fileName + ".dem"
		}
		filePath = filepath.Join(replayDir, fileName)
	} else {
		http.Error(w, "Invalid request: matchId or filePath required", http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(filePath, ".dem") {
		http.Error(w, "Can only delete .dem files", http.StatusBadRequest)
		return
	}

	absReplayDir, _ := filepath.Abs(replayDir)
	absFilePath, _ := filepath.Abs(filePath)

	if !strings.HasPrefix(absFilePath, absReplayDir) {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	if err := os.Remove(filePath); err != nil {
		log.Printf("Error deleting file %s: %v", filePath, err)
		http.Error(w, fmt.Sprintf("Error deleting file: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully deleted replay file: %s", filePath)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Replay file %s deleted successfully", filepath.Base(filePath)),
	})
}

type PlayerInfoRequest struct {
	MatchID     string `json:"matchId"`
	FilePath    string `json:"filePath,omitempty"`
	ProfileName string `json:"profileName"`
}

func handlePlayerInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PlayerInfoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	matchID, err := strconv.ParseInt(req.MatchID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid match ID", http.StatusBadRequest)
		return
	}

	var filePath string
	
	// If filePath is provided, use it directly (supports subfolders)
	if req.FilePath != "" {
		replayDir := getProfileReplayDir(req.ProfileName)
		fatalDir := getFatalReplayDir(req.ProfileName)
		
		// Try regular directory first
		regularPath := filepath.Join(replayDir, req.FilePath)
		if _, err := os.Stat(regularPath); err == nil {
			filePath = regularPath
		} else {
			// Try fatal directory
			fatalPath := filepath.Join(fatalDir, req.FilePath)
			if _, err := os.Stat(fatalPath); err == nil {
				filePath = fatalPath
			} else {
				http.Error(w, fmt.Sprintf("Replay file not found: %s", req.FilePath), http.StatusNotFound)
				return
			}
		}
	} else {
		// Fallback: search for the file by match ID
		replayDir := getProfileReplayDir(req.ProfileName)
		fatalDir := getFatalReplayDir(req.ProfileName)
		
		// Try regular directory first
		regularPath := filepath.Join(replayDir, req.MatchID+".dem")
		if _, err := os.Stat(regularPath); err == nil {
			filePath = regularPath
		} else {
			// Search in fatal directory (including subfolders)
			found := false
			files, err := ioutil.ReadDir(fatalDir)
			if err == nil {
				for _, file := range files {
					if file.IsDir() {
						dateSubDir := filepath.Join(fatalDir, file.Name())
						subFiles, err := ioutil.ReadDir(dateSubDir)
						if err != nil {
							continue
						}
						for _, subFile := range subFiles {
							if !subFile.IsDir() && subFile.Name() == req.MatchID+".dem" {
								filePath = filepath.Join(dateSubDir, subFile.Name())
								found = true
								break
							}
						}
						if found {
							break
						}
					} else if file.Name() == req.MatchID+".dem" {
						filePath = filepath.Join(fatalDir, file.Name())
						found = true
						break
					}
				}
			}
			
			if !found {
				http.Error(w, fmt.Sprintf("Replay file not found: %s.dem", req.MatchID), http.StatusNotFound)
				return
			}
		}
	}
	
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Could not open replay file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	players, err := parser.ExtractPlayerInfo(matchID, file)
	if err != nil {
		http.Error(w, "Error extracting player info: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(players)
}

func handleHeroIcon(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	heroId := strings.TrimPrefix(r.URL.Path, "/api/hero-icon/")
	if heroId == "" {
		http.Error(w, "Missing hero ID", http.StatusBadRequest)
		return
	}

	heroId = strings.TrimSuffix(heroId, "_icon.png")
	heroId = strings.TrimSuffix(heroId, "_full.png")
	heroId = strings.TrimSuffix(heroId, "_icon")
	heroId = strings.TrimSuffix(heroId, "_full")
	heroId = strings.Split(heroId, "?")[0]
	heroId = strings.TrimSpace(heroId)

	if heroId == "" {
		log.Printf("Invalid hero ID from path: %s", r.URL.Path)
		http.Error(w, "Invalid hero ID", http.StatusBadRequest)
		return
	}

	heroIdMapping := map[string]string{
		"zeus": "zuus",
	}
	if mappedId, ok := heroIdMapping[heroId]; ok {
		heroId = mappedId
	}

	heroesWithoutIcon := map[string]bool{
		"primal_beast": true,
		"ringmaster":   true,
		"marci":        true,
		"muerta":       true,
	}

	var iconUrl string
	if heroId == "kez" || heroesWithoutIcon[heroId] {
		iconUrl = fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/apps/dota2/images/heroes/%s_full.png", heroId)
	} else {
		iconUrl = fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/apps/dota2/images/heroes/%s_icon.png", heroId)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Get(iconUrl)
	if err != nil {
		log.Printf("Failed to fetch hero icon %s: %v", iconUrl, err)
		http.Error(w, fmt.Sprintf("Failed to fetch icon: %v", err), http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if !heroesWithoutIcon[heroId] && heroId != "kez" {
			log.Printf("Hero icon %s returned status %d, trying full image", iconUrl, resp.StatusCode)
			fullUrl := fmt.Sprintf("https://cdn.cloudflare.steamstatic.com/apps/dota2/images/heroes/%s_full.png", heroId)
			resp2, err2 := client.Get(fullUrl)
			if err2 != nil {
				log.Printf("Failed to fetch hero full image %s: %v", fullUrl, err2)
				http.Error(w, fmt.Sprintf("Failed to fetch icon: status %d", resp.StatusCode), http.StatusInternalServerError)
				return
			}
			defer resp2.Body.Close()
			if resp2.StatusCode == http.StatusOK {
				w.Header().Set("Content-Type", resp2.Header.Get("Content-Type"))
				w.Header().Set("Cache-Control", "public, max-age=86400")
				io.Copy(w, resp2.Body)
				return
			}
			log.Printf("Hero full image %s also returned status %d", fullUrl, resp2.StatusCode)
		}
		http.Error(w, fmt.Sprintf("Failed to fetch icon: status %d", resp.StatusCode), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	w.Header().Set("Cache-Control", "public, max-age=86400")
	io.Copy(w, resp.Body)
}

type ParseRequest struct {
	MatchID         string `json:"matchId"`
	FilePath        string `json:"filePath"`
	ReportedSlot    int    `json:"reportedSlot"`
	ReportedSteamID string `json:"reportedSteamId"`
	ProfileName     string `json:"profileName"`
}

func handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	matchID, err := strconv.ParseInt(req.MatchID, 10, 64)
	if err != nil {
		http.Error(w, "Invalid match ID", http.StatusBadRequest)
		return
	}

	var reportedSteamID uint64
	if req.ReportedSteamID != "" {
		reportedSteamID, err = strconv.ParseUint(req.ReportedSteamID, 10, 64)
		if err != nil {
			http.Error(w, "Invalid reportedSteamId", http.StatusBadRequest)
			return
		}
		// Ensure we're using SteamID64 for the parser comparison
		// The parser checks against e.GetUint64("m_iPlayerSteamID") which is typically SteamID64 in Replays
		reportedSteamID = convertSteamID(reportedSteamID, true)
	}

	replayDir := getProfileReplayDir(req.ProfileName)
	
	var filePath string
	if req.FilePath != "" {
		filePath = filepath.Join(replayDir, req.FilePath)
		if !strings.HasPrefix(filePath, replayDir) {
			http.Error(w, "Invalid file path", http.StatusBadRequest)
			return
		}
	} else {
		filePath = filepath.Join(replayDir, req.MatchID+".dem")
	}
	
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Replay file not found: %s - Make sure the replay file exists in your replay directory", req.FilePath), http.StatusNotFound)
		return
	}
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Could not open replay file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	result, err := parser.ParseReplay(matchID, file, req.ReportedSlot, reportedSteamID)
	if err != nil {
		http.Error(w, "Error parsing replay: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(result)
}

func handleHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	steamIDStr := r.URL.Query().Get("steamId")
	if steamIDStr == "" {
		http.Error(w, "Missing steamId parameter", http.StatusBadRequest)
		return
	}

	steamID, err := strconv.ParseInt(steamIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid steamId", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err == nil && l > 0 {
			limit = l
		}
	}

	turboOnly := r.URL.Query().Get("turboOnly") == "true"

	if gcClient == nil {
		http.Error(w, "GC client not available", http.StatusInternalServerError)
		return
	}

	status := gcClient.GetStatus()
	if status != botclient.StatusGCReady && status != botclient.StatusConnected {
		http.Error(w, "GC not ready. Please connect to Steam first.", http.StatusBadRequest)
		return
	}

	log.Printf("Using GC for steamID: %d, limit: %d, turboOnly: %v", steamID, limit, turboOnly)
	matches, err := gcClient.GetPlayerMatchHistory(steamID, limit, turboOnly)
	if err != nil {
		log.Printf("Error fetching matches from GC: %v", err)
		http.Error(w, fmt.Sprintf("Error fetching matches: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(matches)
}

type FatalSearchRequest struct {
	SteamIDStr    string `json:"steamId"`
	MaxDepth      int    `json:"maxDepth"`
	ProfileName   string `json:"profileName"`
	GamesPerFatal int    `json:"gamesPerFatal"`
}

func handleFatalSearch(w http.ResponseWriter, r *http.Request) {
	log.Printf("[FATAL_SEARCH] Received request: Method=%s", r.Method)
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FatalSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[FATAL_SEARCH] Error decoding request: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	steamID, err := strconv.ParseInt(req.SteamIDStr, 10, 64)
	if err != nil {
		log.Printf("[FATAL_SEARCH] Error parsing steamID: %v", err)
		http.Error(w, "Invalid steamId", http.StatusBadRequest)
		return
	}

	log.Printf("[FATAL_SEARCH] Request decoded: steamID=%d, maxDepth=%d, profileName=%s", steamID, req.MaxDepth, req.ProfileName)

	if gcClient == nil {
		log.Printf("[FATAL_SEARCH] ERROR: GC client is nil")
		http.Error(w, "GC client not available", http.StatusInternalServerError)
		return
	}

	status := gcClient.GetStatus()
	log.Printf("[FATAL_SEARCH] GC client status: %d", status)
	if status != botclient.StatusGCReady && status != botclient.StatusConnected {
		log.Printf("[FATAL_SEARCH] ERROR: GC not ready (status=%d, need %d or %d)", status, botclient.StatusGCReady, botclient.StatusConnected)
		http.Error(w, "GC not ready. Please connect to Steam first.", http.StatusBadRequest)
		return
	}

	if req.MaxDepth < 1 {
		log.Printf("[FATAL_SEARCH] ERROR: Invalid maxDepth: %d", req.MaxDepth)
		http.Error(w, "maxDepth must be at least 1", http.StatusBadRequest)
		return
	}

	log.Printf("[FATAL_SEARCH] Starting fatal search: steamID=%d, maxDepth=%d, gamesPerFatal=%d, profile=%s", steamID, req.MaxDepth, req.GamesPerFatal, req.ProfileName)
	matches, err := gcClient.FindFatalGames(steamID, req.MaxDepth, req.GamesPerFatal)
	if err != nil {
		log.Printf("[FATAL_SEARCH] ERROR finding fatal games: %v", err)
		http.Error(w, fmt.Sprintf("Error finding fatal games: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("[FATAL_SEARCH] Success: found %d fatal matches", len(matches))
	json.NewEncoder(w).Encode(map[string]interface{}{
		"matches": matches,
		"count":   len(matches),
	})

	// DEPRECATED: Stratz/OpenDota code paths removed - Steam GC is now the only active method
	// The following code is kept for reference but is no longer called:
	/*
	// Stratz API typically expects SteamID3 (32-bit) OR SteamID64 depending on the endpoint.
	// For `player(steamAccountId: ...)` GraphQL query, it typically expects the SteamID3 (Account ID).
	// Let's convert to 32-bit (SteamID3) just to be safe, as that is 'steamAccountId'.
	steamID = int64(convertSteamID(uint64(steamID), false))

	if config.StratzAPIToken == "" {
		log.Printf("Stratz API Token is empty when trying to fetch history")
		http.Error(w, "Stratz API Token not configured", http.StatusInternalServerError)
		return
	}
	log.Printf("Using Stratz API Token (length: %d) for steamID: %d, turboOnly: %v", len(config.StratzAPIToken), steamID, turboOnly)

	client := stratz.NewClient(config.StratzAPIToken)
	matches, err := client.GetLastMatches(steamID, limit, turboOnly)
	if err != nil {
		log.Printf("Error fetching matches from Stratz: %v", err)
		http.Error(w, fmt.Sprintf("Error fetching matches: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(matches)
	*/
}

type DownloadRequest struct {
	MatchID            int64  `json:"matchId"`
	ProfileName        string `json:"profileName"`
	Fatal              bool   `json:"fatal"`
	SingleDraftMatchID int64  `json:"singleDraftMatchId,omitempty"`
	SingleDraftDate    uint32 `json:"singleDraftDate,omitempty"`
	GamesPerFatal      int     `json:"gamesPerFatal,omitempty"` // Number of games to download per fatal (default: 2)
	SteamID            int64   `json:"steamId,omitempty"`        // Steam ID needed to fetch match history for additional games
	AdditionalMatchIDs []int64 `json:"additionalMatchIds,omitempty"` // List of additional match IDs to download (pre-calculated)
}

func handleDownload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req DownloadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.MatchID == 0 {
		http.Error(w, "Invalid match ID", http.StatusBadRequest)
		return
	}

	gamesPerFatal := req.GamesPerFatal
	if gamesPerFatal < 1 {
		gamesPerFatal = 2 // Default to 2 ranked games before singledraft
	}
	if gamesPerFatal > 15 {
		gamesPerFatal = 15 // Cap at 15
	}
	
	lockInterface, _ := handlerLocks.LoadOrStore(req.MatchID, &sync.Mutex{})
	handlerLock := lockInterface.(*sync.Mutex)
	handlerLock.Lock()
	defer handlerLock.Unlock()
	
	log.Printf("Starting download process for match %d (profile: %s, fatal: %v, gamesPerFatal: %d)", req.MatchID, req.ProfileName, req.Fatal, gamesPerFatal)

	var replayDir string
	if req.Fatal {
		if req.SingleDraftDate > 0 {
			dateTime := time.Unix(int64(req.SingleDraftDate), 0)
			dateFolder := dateTime.Format("2006-01-02")
			replayDir = filepath.Join(getFatalReplayDir(req.ProfileName), dateFolder)
		} else {
			replayDir = getFatalReplayDir(req.ProfileName)
		}
	} else {
		replayDir = getProfileReplayDir(req.ProfileName)
	}
	if err := os.MkdirAll(replayDir, os.ModePerm); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create profile directory: %v", err), http.StatusInternalServerError)
		return
	}

	if req.Fatal && req.SingleDraftMatchID > 0 {
		tempDir := getFatalReplayDir(req.ProfileName)
		
		// Fetch ranked games before singledraft
		// gamesPerFatal = number of ranked games in sequence before the singledraft game
		// So if gamesPerFatal=2, we need: 2 ranked games before singledraft
		// If gamesPerFatal=5, we need: 5 ranked games before singledraft
		var additionalMatchIDs []int64
		var fetchFailed bool
		neededRankedGames := gamesPerFatal // gamesPerFatal is the count of ranked games before singledraft
		
		if len(req.AdditionalMatchIDs) > 0 {
			additionalMatchIDs = req.AdditionalMatchIDs
			log.Printf("Using provided additional ranked games list: %v", additionalMatchIDs)
			if len(additionalMatchIDs) < neededRankedGames && req.SteamID > 0 && gcClient != nil {
				log.Printf("Provided list has %d games but need %d, fetching additional games", len(additionalMatchIDs), neededRankedGames)
				moreGames := fetchAdditionalGames(req.SteamID, req.SingleDraftMatchID, req.MatchID, neededRankedGames-len(additionalMatchIDs), gcClient)
				for _, gameID := range moreGames {
					found := false
					for _, existingID := range additionalMatchIDs {
						if gameID == existingID {
							found = true
							break
						}
					}
					if !found {
						additionalMatchIDs = append(additionalMatchIDs, gameID)
					}
				}
				log.Printf("Extended additional ranked games list to %d games: %v", len(additionalMatchIDs), additionalMatchIDs)
			}
		} else if neededRankedGames > 0 && req.SteamID > 0 && gcClient != nil {
			log.Printf("Fetching ranked games before singledraft: gamesPerFatal=%d, need %d ranked games before singledraft", gamesPerFatal, neededRankedGames)
			additionalMatchIDs = fetchAdditionalGames(req.SteamID, req.SingleDraftMatchID, req.MatchID, neededRankedGames, gcClient)
			if len(additionalMatchIDs) == 0 && neededRankedGames > 0 {
				// If we need additional games but couldn't fetch them, we can't verify they exist
				// So we should proceed with download attempt
				fetchFailed = true
				log.Printf("Could not fetch additional ranked games list, will attempt to download them")
			} else {
				log.Printf("Found %d ranked games before singledraft to check: %v", len(additionalMatchIDs), additionalMatchIDs)
			}
		}
		
		// Check for existing files if we successfully fetched the additional games list
		if !fetchFailed {
			var existingDateFolder string
			files, _ := ioutil.ReadDir(tempDir)
			for _, file := range files {
				if file.IsDir() {
					checkFatalPath := filepath.Join(tempDir, file.Name(), fmt.Sprintf("%d.dem", req.MatchID))
					fatalExists := false
					
					if _, err := os.Stat(checkFatalPath); err == nil {
						fatalExists = true
					}
					
					// Count how many additional ranked games actually exist
					existingAdditionalCount := 0
					for _, additionalMatchID := range additionalMatchIDs {
						checkAdditionalPath := filepath.Join(tempDir, file.Name(), fmt.Sprintf("%d.dem", additionalMatchID))
						if _, err := os.Stat(checkAdditionalPath); err == nil {
							existingAdditionalCount++
						}
					}
					
					// We need gamesPerFatal ranked games before singledraft
					requiredAdditional := neededRankedGames
					if fatalExists && existingAdditionalCount >= requiredAdditional {
						existingDateFolder = file.Name()
						log.Printf("Found all required games in %s: fatal=%v, ranked before singledraft=%d/%d", file.Name(), fatalExists, existingAdditionalCount, requiredAdditional)
						break
					}
				}
			}
			
			if existingDateFolder != "" {
				totalGames := gamesPerFatal
				if totalGames == 0 {
					totalGames = 1 + len(additionalMatchIDs)
				}
				log.Printf("All %d replays already exist in %s: fatal=%d, ranked before singledraft=%v, skipping download", totalGames, existingDateFolder, req.MatchID, additionalMatchIDs)
				w.Header().Set("Content-Type", "application/json")
				message := fmt.Sprintf("Replays already exist in %s: fatal=%d.dem", existingDateFolder, req.MatchID)
				if len(additionalMatchIDs) > 0 {
					message += fmt.Sprintf(" + %d ranked games before singledraft", len(additionalMatchIDs))
				}
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": true,
					"message": message,
				})
				return
			}
		}
		
		// Check if fatal match already exists in any date folder or tempDir root
		var fatalPath string
		var fatalExists bool
		files, _ := ioutil.ReadDir(tempDir)
		for _, file := range files {
			if file.IsDir() {
				checkPath := filepath.Join(tempDir, file.Name(), fmt.Sprintf("%d.dem", req.MatchID))
				if _, err := os.Stat(checkPath); err == nil {
					fatalPath = checkPath
					fatalExists = true
					log.Printf("Fatal replay %d already exists in %s, will reuse it", req.MatchID, file.Name())
					break
				}
			}
		}
		
		// If not found in date folders, check tempDir root
		if !fatalExists {
			fatalPath = filepath.Join(tempDir, fmt.Sprintf("%d.dem", req.MatchID))
			if _, err := os.Stat(fatalPath); err == nil {
				fatalExists = true
				log.Printf("Fatal replay %d already exists in tempDir, will reuse it", req.MatchID)
			}
		}
		
		var fatalDownloaded bool
		
		if !fatalExists {
			// Get or create lock for fatal match ID
			lockInterface, _ := downloadLocks.LoadOrStore(req.MatchID, &sync.Mutex{})
			lock := lockInterface.(*sync.Mutex)
			lock.Lock()
			
			// Double-check existence after acquiring lock
			if _, err := os.Stat(fatalPath); err == nil {
				fatalExists = true
				log.Printf("Fatal replay %d was downloaded by another request, reusing it", req.MatchID)
				lock.Unlock()
			} else {
				if err := downloader.DownloadReplay(req.MatchID, tempDir, config.StratzAPIToken, config.SteamAPIKey, gcClient); err != nil {
					lock.Unlock()
					if !strings.Contains(err.Error(), "queued for parsing") {
						log.Printf("Error downloading fatal replay for match %d: %v", req.MatchID, err)
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(http.StatusInternalServerError)
						errorMsg := fmt.Sprintf("Error downloading fatal replay: %v", err)
						// Check for expired replay (404) first
						if strings.Contains(err.Error(), "replay has likely expired") || strings.Contains(err.Error(), "replay not found (404)") {
							errorMsg = "Replay has expired (7-14 day limit). The replay is no longer available on Valve's servers."
						} else if strings.Contains(err.Error(), "failed to download replay after trying all URLs") {
							errorMsg = "Failed to download replay: All download URLs failed. The replay may be unavailable or the servers are down."
						}
						json.NewEncoder(w).Encode(map[string]interface{}{
							"success": false,
							"error":   errorMsg,
						})
						return
					}
				}
				fatalDownloaded = true
				lock.Unlock()
				// Update fatalPath to the newly downloaded file
				fatalPath = filepath.Join(tempDir, fmt.Sprintf("%d.dem", req.MatchID))
			}
		}

		// If we need additional ranked games and we haven't fetched them yet (or fetch failed earlier), try again
		if neededRankedGames > 0 && req.SteamID > 0 && gcClient != nil && len(additionalMatchIDs) == 0 {
			log.Printf("Fetching ranked games before singledraft for download: gamesPerFatal=%d, need %d ranked games", gamesPerFatal, neededRankedGames)
			additionalMatchIDs = fetchAdditionalGames(req.SteamID, req.SingleDraftMatchID, req.MatchID, neededRankedGames, gcClient)
			if len(additionalMatchIDs) == 0 {
				log.Printf("Warning: Could not fetch additional ranked games, will only download fatal match (1 game instead of %d)", gamesPerFatal)
			} else {
				log.Printf("Found %d ranked games before singledraft to download: %v", len(additionalMatchIDs), additionalMatchIDs)
			}
		}

		// Download additional ranked games before singledraft if any
		var additionalDownloaded []int64
		for _, additionalMatchID := range additionalMatchIDs {
			// Get or create lock for this match ID
			lockInterface, _ := downloadLocks.LoadOrStore(additionalMatchID, &sync.Mutex{})
			lock := lockInterface.(*sync.Mutex)
			lock.Lock()
			
			// Double-check existence after acquiring lock (another request might have downloaded it)
			additionalExists := false
			additionalPath := ""
			
			// Check in target date folder first
			if req.SingleDraftDate > 0 {
				checkPath := filepath.Join(replayDir, fmt.Sprintf("%d.dem", additionalMatchID))
				if _, err := os.Stat(checkPath); err == nil {
					additionalExists = true
					additionalPath = checkPath
				}
			}
			
			// If not found in target date folder, check all date subdirectories
			if !additionalExists {
				files, _ := ioutil.ReadDir(tempDir)
				for _, file := range files {
					if file.IsDir() {
						checkPath := filepath.Join(tempDir, file.Name(), fmt.Sprintf("%d.dem", additionalMatchID))
						if _, err := os.Stat(checkPath); err == nil {
							additionalExists = true
							additionalPath = checkPath
							break
						}
					}
				}
			}
			
			// If not found in date folders, check tempDir root
			if !additionalExists {
				checkPath := filepath.Join(tempDir, fmt.Sprintf("%d.dem", additionalMatchID))
				if _, err := os.Stat(checkPath); err == nil {
					additionalExists = true
					additionalPath = checkPath
				}
			}
			
			if additionalExists {
				log.Printf("Additional ranked game %d already exists in %s, skipping", additionalMatchID, additionalPath)
				additionalDownloaded = append(additionalDownloaded, additionalMatchID)
				lock.Unlock()
			} else {
				log.Printf("Downloading additional ranked game %d (before singledraft)", additionalMatchID)
				if err := downloader.DownloadReplay(additionalMatchID, tempDir, config.StratzAPIToken, config.SteamAPIKey, gcClient); err != nil {
					lock.Unlock()
					if !strings.Contains(err.Error(), "queued for parsing") {
						log.Printf("Error downloading additional ranked game %d: %v (continuing with other games)", additionalMatchID, err)
					}
				} else {
					additionalDownloaded = append(additionalDownloaded, additionalMatchID)
					log.Printf("Successfully downloaded additional ranked game %d", additionalMatchID)
					lock.Unlock()
				}
			}
		}

		_, fatalCheckErr := os.Stat(fatalPath)
		if fatalCheckErr != nil && os.IsNotExist(fatalCheckErr) {
			log.Printf("Fatal replay still missing, may be queued")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"status":  "queued",
				"message": fmt.Sprintf("Fatal match queued for parsing, will be downloaded automatically when ready"),
			})
			return
		}

		// Extract date from fatal match (not singledraft) to create date folder
		fatalFile, err := os.Open(fatalPath)
		if err != nil {
			log.Printf("Error reading fatal replay: %v", err)
			if fatalDownloaded {
				os.Remove(fatalPath)
			}
			for _, additionalMatchID := range additionalDownloaded {
				additionalPath := filepath.Join(tempDir, fmt.Sprintf("%d.dem", additionalMatchID))
				os.Remove(additionalPath)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Error reading fatal replay: %v", err),
			})
			return
		}

		replayDate, err := parser.GetReplayDate(fatalFile)
		fatalFile.Close()
		
		if err != nil {
			log.Printf("Error extracting date from fatal replay: %v, using file mod time", err)
			if fileInfo, err := os.Stat(fatalPath); err == nil {
				replayDate = fileInfo.ModTime()
			} else {
				replayDate = time.Now()
			}
		}

		dateFolder := replayDate.Format("2006-01-02")
		finalDir := filepath.Join(tempDir, dateFolder)
		if err := os.MkdirAll(finalDir, os.ModePerm); err != nil {
			log.Printf("Error creating date folder: %v", err)
			if fatalDownloaded {
				os.Remove(fatalPath)
			}
			for _, additionalMatchID := range additionalDownloaded {
				additionalPath := filepath.Join(tempDir, fmt.Sprintf("%d.dem", additionalMatchID))
				os.Remove(additionalPath)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   fmt.Sprintf("Error creating date folder: %v", err),
			})
			return
		}

		finalFatalPath := filepath.Join(finalDir, fmt.Sprintf("%d.dem", req.MatchID))

		// Check if file is already in the correct final location
		if _, err := os.Stat(finalFatalPath); err == nil {
			log.Printf("Fatal replay %d already exists in final destination %s, skipping move", req.MatchID, dateFolder)
			// If it was in a different location, we can leave it (or optionally remove the old one)
			// For now, we'll just use the existing file
		} else if fatalPath != finalFatalPath {
			if err := os.Rename(fatalPath, finalFatalPath); err != nil {
				log.Printf("Error moving fatal replay to date folder: %v", err)
			} else {
				log.Printf("Moved fatal replay %d to date folder %s", req.MatchID, dateFolder)
			}
		}

		// Move additional ranked games to date folder
		for _, additionalMatchID := range additionalDownloaded {
			finalAdditionalPath := filepath.Join(finalDir, fmt.Sprintf("%d.dem", additionalMatchID))
			if _, err := os.Stat(finalAdditionalPath); err == nil {
				log.Printf("Additional ranked game %d already exists in final destination %s, skipping move", additionalMatchID, finalAdditionalPath)
				continue
			}
			
			additionalPath := filepath.Join(tempDir, fmt.Sprintf("%d.dem", additionalMatchID))
			if _, err := os.Stat(additionalPath); err != nil {
				log.Printf("Additional ranked game %d not found in tempDir, may already be in date folder, skipping move", additionalMatchID)
				continue
			}
			
			if additionalPath != finalAdditionalPath {
				if err := os.Rename(additionalPath, finalAdditionalPath); err != nil {
					log.Printf("Error moving additional ranked game %d to date folder: %v", additionalMatchID, err)
				} else {
					log.Printf("Moved additional ranked game %d to date folder", additionalMatchID)
				}
			}
		}

		totalGames := gamesPerFatal
		if totalGames == 0 {
			totalGames = len(additionalDownloaded) + 1
		}
		log.Printf("Successfully downloaded %d replays to %s: fatal=%d, ranked before singledraft=%v", totalGames, dateFolder, req.MatchID, additionalDownloaded)
		w.Header().Set("Content-Type", "application/json")
		message := fmt.Sprintf("Replays downloaded successfully to %s: fatal=%d.dem", dateFolder, req.MatchID)
		if len(additionalDownloaded) > 0 {
			message += fmt.Sprintf(" + %d ranked games before singledraft", len(additionalDownloaded))
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": message,
			"dateFolder": dateFolder,
		})
		return
	}

	if err := downloader.DownloadReplay(req.MatchID, replayDir, config.StratzAPIToken, config.SteamAPIKey, gcClient); err != nil {
		// Check if match was queued for parsing (background processing)
		if strings.Contains(err.Error(), "queued for parsing") {
			log.Printf("Match %d queued for parsing, will be processed in background", req.MatchID)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusAccepted)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"status":  "queued",
				"message": fmt.Sprintf("Match %d queued for parsing, will be downloaded automatically when ready", req.MatchID),
			})
			return
		}
		log.Printf("Error downloading replay for match %d: %v", req.MatchID, err)
		http.Error(w, fmt.Sprintf("Error downloading replay: %v", err), http.StatusInternalServerError)
		return
	}

	// Verify file actually exists before returning success
	filePath := filepath.Join(replayDir, fmt.Sprintf("%d.dem", req.MatchID))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Printf("Match %d download reported success but file not found, may be queued", req.MatchID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"status":  "queued",
			"message": fmt.Sprintf("Match %d queued for parsing, will be downloaded automatically when ready", req.MatchID),
		})
		return
	}

	log.Printf("Successfully downloaded replay for match %d", req.MatchID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Replay downloaded successfully: %d.dem", req.MatchID),
	})
}

// fetchAdditionalGamesGC fetches match history using the GC client as a fallback
func fetchAdditionalGamesGC(steamID int64, singleDraftMatchID int64, fatalMatchID int64, count int, gcClient *botclient.Client) []int64 {
	if gcClient == nil {
		return []int64{}
	}

	limit := 20

	log.Printf("Fetching match history (GC) starting from singledraft match %d (limit=%d)", singleDraftMatchID, limit)
	
	// Ensure GC is ready before attempting fetch
	status := gcClient.GetStatus()
	if status != botclient.StatusGCReady {
		log.Printf("GC not ready (status: %d), refreshing connection...", status)
		time.Sleep(2 * time.Second)
		status = gcClient.GetStatus()
		if status != botclient.StatusGCReady {
			log.Printf("GC still not ready after refresh, proceeding anyway")
		}
	}
	
	matches, err := gcClient.GetPlayerMatchHistoryPaginated(steamID, limit, false, uint64(singleDraftMatchID))
	if err != nil {
		log.Printf("Error fetching match history from singledraft match %d via GC: %v", singleDraftMatchID, err)
		if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "context deadline exceeded") {
			log.Printf("GC timeout detected, refreshing connection and retrying...")
			time.Sleep(3 * time.Second)
			status = gcClient.GetStatus()
			if status != botclient.StatusGCReady {
				log.Printf("GC not ready after timeout, waiting for ready state...")
				for i := 0; i < 15; i++ {
					time.Sleep(1 * time.Second)
					status = gcClient.GetStatus()
					if status == botclient.StatusGCReady {
						log.Printf("GC is now ready, retrying...")
						break
					}
				}
			}
		}
		// Fallback: try starting from fatal match ID
		log.Printf("Retrying GC from fatal match %d", fatalMatchID)
		matches, err = gcClient.GetPlayerMatchHistoryPaginated(steamID, limit, false, uint64(fatalMatchID))
		if err != nil {
			log.Printf("Error fetching match history from fatal match %d via GC: %v", fatalMatchID, err)
			if strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "context deadline exceeded") {
				log.Printf("GC timeout detected again, refreshing connection and retrying...")
				time.Sleep(3 * time.Second)
				status = gcClient.GetStatus()
				if status != botclient.StatusGCReady {
					log.Printf("GC not ready after timeout, waiting for ready state...")
					for i := 0; i < 15; i++ {
						time.Sleep(1 * time.Second)
						status = gcClient.GetStatus()
						if status == botclient.StatusGCReady {
							log.Printf("GC is now ready, retrying...")
							break
						}
					}
				}
			}
			// Fallback: try starting from 0 (newest)
			log.Printf("Retrying GC from start (0)")
			matches, err = gcClient.GetPlayerMatchHistoryPaginated(steamID, limit, false, 0)
			if err != nil {
				log.Printf("Error fetching match history from start via GC: %v", err)
				return []int64{}
			}
		}
	}

	if len(matches) == 0 {
		log.Printf("No matches returned from GC history fetch")
		return []int64{}
	}

	// Find the singledraft match in the history
	singleDraftIndex := -1
	for i, m := range matches {
		if m.ID == singleDraftMatchID {
			singleDraftIndex = i
			break
		}
	}

	if singleDraftIndex == -1 {
		log.Printf("Could not find singledraft match %d in GC history. Searching linearly...", singleDraftMatchID, len(matches))
		// If we started from 0, maybe we can just process the list if we find the fatal match?
		// Or just return whatever ranked matches we found?
		// Let's try to find the fatal match at least
		for i, m := range matches {
			if m.ID == fatalMatchID {
				singleDraftIndex = i - 1 // Singledraft should be before (newer than) fatal in history? No, single draft is NEWER than fatal in our logic?
				// Wait.
				// Fatal Match: The game that caused the Low Priority.
				// Single Draft Match: The FIRST game of Low Priority.
				// Chronologically: Fatal Game -> ... -> Single Draft Game.
				// Match History (Newest First): [Single Draft, ..., Fatal Game]
				
				// So Single Draft is at a lower index (newer) than Fatal Game.
				// We want ranked games BEFORE the Single Draft Game (chronologically).
				// In Match History (Newest First): These are matches at HIGHER indices than Single Draft Game.
				
				// Correct.
				break
			}
		}
	}
	
	startIndex := 0
	if singleDraftIndex != -1 {
		startIndex = singleDraftIndex + 1
	} else {
		// If we can't find the singledraft game, but we fetched matches, 
		// we might be looking at a list that doesn't include it (e.g. started from 0 and SD is old).
		// But we are trying to find games BEFORE the SD game (older than it).
		// So if we don't have the SD game, we can't be sure where to start.
		// However, if we started fetching FROM the SD game, index 0 should be the SD game (or close to it).
		// If we started from 0, we might be too far ahead.
		log.Printf("Warning: Singledraft match not found in GC results. Using all fetched matches.")
	}

	// Get ranked matches before the singledraft (history is ordered newest first)
	// We want matches AFTER the singledraft in the array (which are older games)
	// Only include ranked matches (LobbyType == 7), skip singledraft and other game modes
	const LobbyTypeRanked = 7
	
	var additionalMatches []int64
	for i := startIndex; i < len(matches) && len(additionalMatches) < count; i++ {
		// Skip the fatal match if we encounter it
		if matches[i].ID == fatalMatchID {
			continue
		}
		
		// Only include ranked matches (LobbyType == 7)
		if matches[i].LobbyType == LobbyTypeRanked {
			additionalMatches = append(additionalMatches, matches[i].ID)
		}
	}

	log.Printf("Found %d additional games via GC before singledraft %d: %v", len(additionalMatches), singleDraftMatchID, additionalMatches)
	return additionalMatches
}

// fetchAdditionalGames fetches match history and finds additional games before the singledraft match
// uses Steam Web API first, then GC fallback
func fetchAdditionalGames(steamID int64, singleDraftMatchID int64, fatalMatchID int64, count int, gcClient *botclient.Client) []int64 {
	if count <= 0 {
		return []int64{}
	}

	// Use Steam Web API via steamapi client
	// Convert SteamID64 to AccountID (32-bit)
	accountID := steamID - 76561197960265728
	client := steamapi.NewClient(config.SteamAPIKey)

	// Strategy: Start from the singledraft match ID to fetch matches around it
	// Match history is ordered newest first, so starting from singledraft will give us:
	// [matches newer than singledraft, singledraft, matches older than singledraft]
	
	limit := 20
	
	log.Printf("Fetching match history (Web API) starting from singledraft match %d (limit=%d) to find %d additional games", singleDraftMatchID, limit, count)
	
	// Start from the singledraft match ID
	matches, err := client.GetPlayerMatchHistory(accountID, singleDraftMatchID, limit)
	if err != nil {
		log.Printf("Error fetching match history from singledraft match %d via Web API: %v", singleDraftMatchID, err)
		// Fallback: try starting from fatal match ID
		log.Printf("Retrying from fatal match %d", fatalMatchID)
		matches, err = client.GetPlayerMatchHistory(accountID, fatalMatchID, limit)
		if err != nil {
			log.Printf("Error fetching match history from fatal match %d via Web API: %v", fatalMatchID, err)
			
			// Fallback to GC if Web API fails
			log.Printf("Web API failed, attempting fallback to GC...")
			return fetchAdditionalGamesGC(steamID, singleDraftMatchID, fatalMatchID, count, gcClient)
		}
	}

	if len(matches) == 0 {
		log.Printf("No matches returned from history fetch")
		// Fallback to GC if Web API returns no matches (unexpected if user has matches)
		log.Printf("Web API returned no matches, attempting fallback to GC...")
		return fetchAdditionalGamesGC(steamID, singleDraftMatchID, fatalMatchID, count, gcClient)
	}

	// Find the singledraft match in the history
	singleDraftIndex := -1
	for i, m := range matches {
		if m.MatchID == singleDraftMatchID {
			singleDraftIndex = i
			break
		}
	}

	if singleDraftIndex == -1 {
		log.Printf("Could not find singledraft match %d in history (fetched %d matches). The singledraft may be too far back.", singleDraftMatchID, len(matches))
		// Fallback to GC if index not found in Web API (maybe private profile partial result?)
		log.Printf("Web API did not return singledraft match, attempting fallback to GC...")
		return fetchAdditionalGamesGC(steamID, singleDraftMatchID, fatalMatchID, count, gcClient)
	}

	// Get ranked matches before the singledraft (history is ordered newest first)
	// We want matches AFTER the singledraft in the array (which are older games)
	// Only include ranked matches (LobbyType == 7), skip singledraft and other game modes
	const LobbyTypeRanked = 7
	
	var additionalMatches []int64
	startIndex := singleDraftIndex + 1
	for i := startIndex; i < len(matches) && len(additionalMatches) < count; i++ {
		// Skip the fatal match if we encounter it
		if matches[i].MatchID == fatalMatchID {
			continue
		}
		// Web API doesn't provide GameMode in match history list
		// But since we only want Ranked matches (LobbyType 7), this implicitly filters out most non-ranked modes
		
		// Only include ranked matches (LobbyType == 7)
		if matches[i].LobbyType == LobbyTypeRanked {
			additionalMatches = append(additionalMatches, matches[i].MatchID)
		}
	}

	log.Printf("Found %d additional games before singledraft %d: %v", len(additionalMatches), singleDraftMatchID, additionalMatches)
	return additionalMatches
}

func handleProgress(w http.ResponseWriter, r *http.Request) {
	matchIDStr := r.URL.Query().Get("matchId")
	if matchIDStr == "" {
		http.Error(w, "Missing matchId parameter", http.StatusBadRequest)
		return
	}

	matchID, err := strconv.ParseInt(matchIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid matchId", http.StatusBadRequest)
		return
	}

	// Use Server-Sent Events (SSE) for real-time progress
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create a ticker to poll progress
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	// Timeout after 5 minutes (just in case)
	timeout := time.After(5 * time.Minute)

	for {
		select {
		case <-r.Context().Done():
			return
		case <-timeout:
			return
		case <-ticker.C:
			progress := downloader.GetProgress(matchID)

			// Send event
			fmt.Fprintf(w, "data: %.2f\n\n", progress)
			w.(http.Flusher).Flush()

			if progress >= 100 {
				return
			}
			// If progress is 0 but it might have finished or not started.
			// We rely on the frontend to close the connection when the download API returns.
		}
	}
}
