package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

type Config struct {
	ReplayDir      string `json:"replayDir"`
	StratzAPIToken string `json:"stratzApiToken"`
}

var config Config

func main() {
	// Default config
	homeDir, _ := os.UserHomeDir()
	config.ReplayDir = filepath.Join(homeDir, ".steam/debian-installation/steamapps/common/dota 2 beta/game/dota/replays/")
	config.StratzAPIToken = os.Getenv("STRATZ_API_TOKEN")

	// Serve static files
	fs := http.FileServer(http.Dir("./cmd/server/static"))
	http.Handle("/", fs)

	// API endpoints
	http.HandleFunc("/api/config", handleConfig)
	http.HandleFunc("/api/replays", handleReplays)
	http.HandleFunc("/api/parse", handleParse)
	http.HandleFunc("/api/history", handleHistory)
	http.HandleFunc("/api/download", handleDownload)
	http.HandleFunc("/api/delete", handleDelete)
	fmt.Println("Server started at http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
