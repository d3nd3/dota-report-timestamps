package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/d3nd3/dota-report-timestamps/pkg/downloader"
	"github.com/d3nd3/dota-report-timestamps/pkg/parser"
	"github.com/d3nd3/dota-report-timestamps/pkg/stratz"
)

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
		} else {
			log.Printf("Stratz API Token was empty in request (ReplayDir provided: %v)", newConfig.ReplayDir != "")
		}
		json.NewEncoder(w).Encode(config)
	}
}

type ReplayInfo struct {
	FileName string    `json:"fileName"`
	Date     time.Time `json:"date"`
}

func handleReplays(w http.ResponseWriter, r *http.Request) {
	files, err := ioutil.ReadDir(config.ReplayDir)
	if err != nil {
		http.Error(w, "Could not read replay directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var replays []ReplayInfo
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".dem") {
			replayDate := file.ModTime()

			filePath := filepath.Join(config.ReplayDir, file.Name())
			replayFile, err := os.Open(filePath)
			if err == nil {
				if date, err := parser.GetReplayDate(replayFile); err == nil {
					replayDate = date
				} else {
					// log.Printf("Could not extract date from %s: %v, using mod time", file.Name(), err)
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
	MatchID string `json:"matchId"`
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

	filePath := filepath.Join(config.ReplayDir, fileName)
	absReplayDir, _ := filepath.Abs(config.ReplayDir)
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

type ParseRequest struct {
	MatchID         string `json:"matchId"`
	ReportedSlot    int    `json:"reportedSlot"`
	ReportedSteamID string `json:"reportedSteamId"`
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
	}

	filePath := filepath.Join(config.ReplayDir, req.MatchID+".dem")
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

	if config.StratzAPIToken == "" {
		log.Printf("Stratz API Token is empty when trying to fetch history")
		http.Error(w, "Stratz API Token not configured", http.StatusInternalServerError)
		return
	}
	log.Printf("Using Stratz API Token (length: %d) for steamID: %d", len(config.StratzAPIToken), steamID)

	client := stratz.NewClient(config.StratzAPIToken)
	matches, err := client.GetLastMatches(steamID, limit)
	if err != nil {
		log.Printf("Error fetching matches from Stratz: %v", err)
		http.Error(w, fmt.Sprintf("Error fetching matches: %v", err), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(matches)
}

type DownloadRequest struct {
	MatchID int64 `json:"matchId"`
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

	log.Printf("Starting download process for match %d", req.MatchID)

	// Check if replay URL is already available (match already parsed)
	replayURL, err := downloader.GetReplayURL(req.MatchID)
	if err != nil {
		log.Printf("Error checking replay URL: %v", err)
		http.Error(w, fmt.Sprintf("Error checking replay URL: %v", err), http.StatusInternalServerError)
		return
	}

	if replayURL == "" {
		log.Printf("Match %d not parsed yet, requesting parsing...", req.MatchID)
		jobID, err := downloader.RequestParsing(req.MatchID)
		if err != nil {
			log.Printf("Error requesting parsing: %v", err)
			http.Error(w, fmt.Sprintf("Error requesting parsing: %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("Parsing requested for match %d, job ID: %d, waiting...", req.MatchID, jobID)

		log.Printf("Waiting for match %d to be parsed (this may take a few minutes)...", req.MatchID)
		if err := downloader.WaitForParsing(req.MatchID, jobID, 5*time.Minute); err != nil {
			log.Printf("Error waiting for parsing: %v", err)
			http.Error(w, fmt.Sprintf("Error waiting for parsing: %v", err), http.StatusInternalServerError)
			return
		}
		log.Printf("Match %d parsed successfully", req.MatchID)
	} else {
		log.Printf("Match %d already parsed, proceeding with download", req.MatchID)
	}

	log.Printf("Downloading replay for match %d", req.MatchID)
	if err := downloader.DownloadReplay(req.MatchID, config.ReplayDir); err != nil {
		log.Printf("Error downloading replay: %v", err)
		http.Error(w, fmt.Sprintf("Error downloading replay: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully downloaded replay for match %d", req.MatchID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("Replay downloaded successfully: %d.dem", req.MatchID),
	})
}
