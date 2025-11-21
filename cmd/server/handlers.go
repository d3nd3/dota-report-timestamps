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
	"time"

	"github.com/d3nd3/dota-report-timestamps/pkg/botclient"
	"github.com/d3nd3/dota-report-timestamps/pkg/downloader"
	"github.com/d3nd3/dota-report-timestamps/pkg/parser"
	"github.com/d3nd3/dota-report-timestamps/pkg/stratz"
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
			http.Error(w, "Failed to submit code: "+err.Error(), http.StatusInternalServerError)
			return
		}
		// Give it a moment to process
		time.Sleep(500 * time.Millisecond)
	} else {
		// Standard login request
		// Check current status first - if already connecting/connected, return current status
		// StatusNeedGuardCode is NOT included here - we allow re-init to start fresh connection
		currentStatus := gcClient.GetStatus()
		if currentStatus == botclient.StatusConnecting ||
			currentStatus == botclient.StatusConnected ||
			currentStatus == botclient.StatusGCReady {
			log.Printf("Steam client already in state %d, returning current status", currentStatus)
			// Return current status without re-initializing
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": true,
				"status":  int(currentStatus),
				"message": "Already connected or connecting",
			})
			return
		}

		// Init if disconnected or StatusNeedGuardCode (allows fresh connection attempt)
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
	if gcClient != nil {
		status = gcClient.GetStatus()
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
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     int(status),
		"statusText": statusText,
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

func sanitizeFileName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "/", "_")
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "..", "_")
	return name
}

func handleReplays(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	profileName := r.URL.Query().Get("profile")
	replayDir := getProfileReplayDir(profileName)

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

	replays := []ReplayInfo{}
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

	sort.Slice(replays, func(i, j int) bool {
		return replays[i].Date.After(replays[j].Date)
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(replays)
}

type DeleteRequest struct {
	MatchID     string `json:"matchId"`
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

	if req.MatchID == "" {
		http.Error(w, "Invalid match ID", http.StatusBadRequest)
		return
	}

	fileName := req.MatchID
	if !strings.HasSuffix(fileName, ".dem") {
		fileName = fileName + ".dem"
	}

	if strings.Contains(fileName, "..") || strings.Contains(fileName, "/") || strings.Contains(fileName, "\\") {
		http.Error(w, "Invalid file name", http.StatusBadRequest)
		return
	}

	if !strings.HasSuffix(fileName, ".dem") {
		http.Error(w, "Can only delete .dem files", http.StatusBadRequest)
		return
	}

	replayDir := getProfileReplayDir(req.ProfileName)
	filePath := filepath.Join(replayDir, fileName)
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
		"message": fmt.Sprintf("Replay file %s deleted successfully", fileName),
	})
}

type PlayerInfoRequest struct {
	MatchID     string `json:"matchId"`
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

	replayDir := getProfileReplayDir(req.ProfileName)
	filePath := filepath.Join(replayDir, req.MatchID+".dem")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Replay file not found: %s.dem", req.MatchID), http.StatusNotFound)
		return
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

	log.Printf("Fetching hero icon for: %s", heroId)

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
	filePath := filepath.Join(replayDir, req.MatchID+".dem")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		http.Error(w, fmt.Sprintf("Replay file not found: %s.dem - Make sure the replay file exists in your replay directory", req.MatchID), http.StatusNotFound)
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
	useGC := r.URL.Query().Get("useGC") == "true"

	if useGC {
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
		return
	}

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
}

type DownloadRequest struct {
	MatchID     int64  `json:"matchId"`
	ProfileName string `json:"profileName"`
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

	log.Printf("Starting download process for match %d (profile: %s)", req.MatchID, req.ProfileName)

	replayDir := getProfileReplayDir(req.ProfileName)
	if err := os.MkdirAll(replayDir, os.ModePerm); err != nil {
		http.Error(w, fmt.Sprintf("Failed to create profile directory: %v", err), http.StatusInternalServerError)
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
