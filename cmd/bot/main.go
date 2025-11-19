package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"

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
	Status int `json:"status"`
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
	http.HandleFunc("/status", handleStatus)
	http.HandleFunc("/replay-info", handleReplayInfo)
	http.HandleFunc("/player-match-history", handlePlayerMatchHistory)

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
	defer mu.Unlock()

	if client != nil {
		client.Close()
	}

	log.Printf("Initializing bot for user %s", req.Username)
	client = dota2gc.NewClient(req.Username, req.Password)
	if err := client.Connect(); err != nil {
		log.Printf("Failed to connect: %v", err)
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
	client.SubmitCode(req.Code)
	w.WriteHeader(http.StatusOK)
}

func handleStatus(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	status := 0 // Disconnected
	if client != nil {
		status = int(client.GetStatus())
	}

	json.NewEncoder(w).Encode(StatusResponse{Status: status})
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
	SteamID64 int64 `json:"steamId64"`
	Limit     int   `json:"limit"`
	TurboOnly bool  `json:"turboOnly"`
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

	matches, err := c.GetPlayerMatchHistory(req.SteamID64, req.Limit, req.TurboOnly)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(matches)
}

