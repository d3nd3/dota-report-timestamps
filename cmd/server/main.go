package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/d3nd3/dota-report-timestamps/pkg/botclient"
)

type Config struct {
	ReplayDir      string `json:"replayDir"`
	StratzAPIToken string `json:"stratzApiToken"`
	SteamAPIKey    string `json:"steamApiKey"`
	SteamUser      string `json:"steamUser"`
	SteamPass      string `json:"steamPass"`
}

var config Config
var gcClient *botclient.Client

func main() {
	// Default config
	homeDir, _ := os.UserHomeDir()
	config.ReplayDir = filepath.Join(homeDir, ".steam/debian-installation/steamapps/common/dota 2 beta/game/dota/replays/")
	config.StratzAPIToken = os.Getenv("STRATZ_API_TOKEN")
	config.SteamAPIKey = os.Getenv("STEAM_API_KEY")
	config.SteamUser = os.Getenv("STEAM_USER")
	config.SteamPass = os.Getenv("STEAM_PASS")

	// Initialize Bot Client
	gcClient = botclient.NewClient("8082")

	// If credentials are provided via env, try to init the bot
	if config.SteamUser != "" && config.SteamPass != "" {
		log.Printf("Initializing Dota 2 GC Bot for user: %s", config.SteamUser)
		// Run in background as bot process might take a moment to start up
		go func() {
			// Wait for bot process to be ready (run.sh starts it)
			for i := 0; i < 10; i++ {
				if err := gcClient.Init(config.SteamUser, config.SteamPass); err == nil {
					log.Println("Dota 2 GC Bot initialized successfully via env vars")
					return
				}
				time.Sleep(1 * time.Second)
			}
			log.Println("Failed to auto-initialize bot (bot process might not be ready)")
		}()
	}

	// Serve static files
	fs := http.FileServer(http.Dir("./cmd/server/static"))
	http.Handle("/", fs)

	// API endpoints
	http.HandleFunc("/api/config", handleConfig)
	http.HandleFunc("/api/replays", handleReplays)
	http.HandleFunc("/api/parse", handleParse)
	http.HandleFunc("/api/history", handleHistory)
	http.HandleFunc("/api/download", handleDownload)
	http.HandleFunc("/api/progress", handleProgress)
	http.HandleFunc("/api/delete", handleDelete)
	
	// Steam GC endpoints
	http.HandleFunc("/api/steam/login", handleSteamLogin)
	http.HandleFunc("/api/steam/status", handleSteamStatus)

	fmt.Println("Server started at http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
