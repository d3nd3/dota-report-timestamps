package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/d3nd3/dota-report-timestamps/pkg/dota2gc"
)

var (
	client *dota2gc.Client
	mu     sync.Mutex
)

type InitRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type SubmitCodeRequest struct {
	Code string `json:"code"`
}

type StatusResponse struct {
	Status       int    `json:"status"`
	ErrorMessage string `json:"errorMessage,omitempty"`
}

type ReplayInfoResponse struct {
	Cluster uint32 `json:"cluster"`
	Salt    uint64 `json:"salt"`
	Error   string `json:"error,omitempty"`
}

func main() {
	port := os.Getenv("BOT_PORT")
	if port == "" {
		port = "8082"
	}

	http.HandleFunc("/init", handleInit)
	http.HandleFunc("/submit-code", handleSubmitCode)
	http.HandleFunc("/disconnect", handleDisconnect)
	http.HandleFunc("/status", handleStatus)
	http.HandleFunc("/replay-info", handleReplayInfo)
	http.HandleFunc("/player-match-history", handlePlayerMatchHistory)
	http.HandleFunc("/fatal-search", handleFatalSearch)
	http.HandleFunc("/conduct-scorecard", handleConductScorecard)

	log.Printf("Dota 2 GC Bot Service starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleInit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req InitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()

	// Close existing client if any, and wait for cleanup
	if client != nil {
		log.Printf("Closing existing client before reinitializing...")
		client.Close()
		client = nil
		// Give a moment for cleanup to complete
		time.Sleep(500 * time.Millisecond)
	}

	log.Printf("Initializing bot for user %s", req.Username)
	client = dota2gc.NewClient(req.Username, req.Password)

	mu.Unlock() // Release lock before potentially blocking Connect()

	if err := client.Connect(); err != nil {
		log.Printf("Failed to connect: %v", err)
		mu.Lock()
		client = nil
		mu.Unlock()
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func handleSubmitCode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SubmitCodeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if client == nil {
		http.Error(w, "Client not initialized", http.StatusBadRequest)
		return
	}

	log.Printf("Submitting code: %s", req.Code)
	if err := client.SubmitCode(req.Code); err != nil {
		log.Printf("Failed to submit code: %v", err)
		http.Error(w, "Failed to submit code: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleDisconnect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if client == nil {
		// Already disconnected
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Already disconnected",
		})
		return
	}

	log.Printf("Disconnecting Steam client...")
	client.Close()
	client = nil
	log.Printf("Steam client disconnected successfully")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Disconnected successfully",
	})
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	status := 0 // Disconnected
	var errorMessage string
	if client != nil {
		status = int(client.GetStatus())
		errorMessage = client.GetLastErrorMessage()
	}

	json.NewEncoder(w).Encode(StatusResponse{
		Status:       status,
		ErrorMessage: errorMessage,
	})
}

func handleReplayInfo(w http.ResponseWriter, r *http.Request) {
	matchIDStr := r.URL.Query().Get("match_id")
	matchID, err := strconv.ParseUint(matchIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid match_id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	c := client
	mu.Unlock() // Don't hold lock during network call

	if c == nil {
		json.NewEncoder(w).Encode(ReplayInfoResponse{Error: "Client not initialized"})
		return
	}

	// Wait for readiness if needed? The client.GetReplayInfo checks status.
	cluster, salt, err := c.GetReplayInfo(matchID)
	resp := ReplayInfoResponse{
		Cluster: cluster,
		Salt:    salt,
	}
	if err != nil {
		resp.Error = err.Error()
	}

	json.NewEncoder(w).Encode(resp)
}

type PlayerMatchHistoryRequest struct {
	SteamID64      int64  `json:"steamId64"`
	Limit          int    `json:"limit"`
	TurboOnly      bool   `json:"turboOnly"`
	StartAtMatchID uint64 `json:"startAtMatchId,omitempty"`
}

func handlePlayerMatchHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req PlayerMatchHistoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	c := client
	mu.Unlock()

	if c == nil {
		http.Error(w, "Client not initialized", http.StatusBadRequest)
		return
	}

	var matches []dota2gc.Match
	var err error
	if req.StartAtMatchID > 0 {
		matches, err = c.GetPlayerMatchHistoryPaginated(req.SteamID64, req.Limit, req.TurboOnly, req.StartAtMatchID)
	} else {
		matches, err = c.GetPlayerMatchHistory(req.SteamID64, req.Limit, req.TurboOnly)
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(matches)
}

type FatalSearchRequest struct {
	SteamID64     int64 `json:"steamId64"`
	MaxDepth      int   `json:"maxDepth"`
	GamesPerFatal int   `json:"gamesPerFatal"`
}

func handleFatalSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req FatalSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	mu.Lock()
	c := client
	mu.Unlock()

	if c == nil {
		http.Error(w, "Client not initialized", http.StatusBadRequest)
		return
	}

	matches, err := c.FindFatalGames(req.SteamID64, req.MaxDepth, req.GamesPerFatal)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(matches)
}

func handleConductScorecard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	mu.Lock()
	c := client
	mu.Unlock()

	if c == nil {
		http.Error(w, "Client not initialized", http.StatusBadRequest)
		return
	}

	scorecard, err := c.GetPlayerConductScorecard()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(scorecard)
}
