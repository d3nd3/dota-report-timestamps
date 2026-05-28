package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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
var downloadLocks sync.Map // Map[int64]*sync.Mutex to prevent concurrent downloads of the same match
var handlerLocks sync.Map // Map[int64]*sync.Mutex to prevent concurrent handler execution for the same match
var configPath string

func defaultReplayDir() string {
	homeDir, _ := os.UserHomeDir()
	candidates := []string{
		`C:\Program Files (x86)\Steam\steamapps\common\dota 2 beta\game\dota\replays`,
		`C:\Program Files\Steam\steamapps\common\dota 2 beta\game\dota\replays`,
		filepath.Join(homeDir, ".steam/debian-installation/steamapps/common/dota 2 beta/game/dota/replays"),
	}
	for _, p := range candidates {
		if info, err := os.Stat(p); err == nil && info.IsDir() {
			return p
		}
	}
	if runtime.GOOS == "windows" {
		return candidates[0]
	}
	return candidates[2]
}

func initConfigPath() string {
	if dir, err := os.UserConfigDir(); err == nil {
		return filepath.Join(dir, "dota-report-timestamps", "config.json")
	}
	return "config.json"
}

func loadConfigFromDisk() {
	b, err := os.ReadFile(configPath)
	if err != nil {
		return
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		log.Printf("Warning: invalid config file %s: %v", configPath, err)
		return
	}
	if c.ReplayDir != "" {
		config.ReplayDir = c.ReplayDir
	}
	if c.StratzAPIToken != "" {
		config.StratzAPIToken = c.StratzAPIToken
	}
	if c.SteamAPIKey != "" {
		config.SteamAPIKey = c.SteamAPIKey
	}
	if c.SteamUser != "" {
		config.SteamUser = c.SteamUser
	}
	if c.SteamPass != "" {
		config.SteamPass = c.SteamPass
	}
}

func saveConfigToDisk() error {
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(configPath, b, 0o600)
}

func findStaticDir() string {
	possibleDirs := []string{
		"./cmd/server/static",
		"./static",
	}
	
	execPath, err := os.Executable()
	if err == nil {
		execDir := filepath.Dir(execPath)
		possibleDirs = append(possibleDirs, filepath.Join(execDir, "static"))
		possibleDirs = append(possibleDirs, filepath.Join(execDir, "cmd", "server", "static"))
	}
	
	for _, dir := range possibleDirs {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			if _, err := os.Stat(filepath.Join(dir, "index.html")); err == nil {
				return dir
			}
		}
	}
	return ""
}

func noCacheJS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".js") {
			w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
			w.Header().Set("Pragma", "no-cache")
			w.Header().Set("Expires", "0")
			w.Header().Del("ETag")
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	// Default config
	config.ReplayDir = defaultReplayDir()
	config.StratzAPIToken = os.Getenv("STRATZ_API_TOKEN")
	config.SteamAPIKey = os.Getenv("STEAM_API_KEY")
	config.SteamUser = os.Getenv("STEAM_USER")
	config.SteamPass = os.Getenv("STEAM_PASS")
	configPath = initConfigPath()
	loadConfigFromDisk()
	log.Printf("Using replay dir: %s", config.ReplayDir)

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

	// Find static files directory (check multiple possible locations)
	staticDir := findStaticDir()
	if staticDir == "" {
		log.Fatal("Could not find static files directory. Tried: ./cmd/server/static, ./static")
	}
	log.Printf("Serving static files from: %s", staticDir)
	
	// Serve static files
	fs := http.FileServer(http.Dir(staticDir))
	http.Handle("/", noCacheJS(fs))

	// API endpoints
	http.HandleFunc("/api/config", handleConfig)
	http.HandleFunc("/api/replays", handleReplays)
	http.HandleFunc("/api/browse", handleBrowse)
	http.HandleFunc("/api/player-info", handlePlayerInfo)
	http.HandleFunc("/api/parse", handleParse)
	http.HandleFunc("/api/history", handleHistory)
	http.HandleFunc("/api/download", handleDownload)
	http.HandleFunc("/api/progress", handleProgress)
	http.HandleFunc("/api/delete", handleDelete)
	http.HandleFunc("/api/hero-icon/", handleHeroIcon)
	http.HandleFunc("/api/fatal-search", handleFatalSearch)
	
	// Steam GC endpoints
	http.HandleFunc("/api/steam/login", handleSteamLogin)
	http.HandleFunc("/api/steam/disconnect", handleSteamDisconnect)
	http.HandleFunc("/api/steam/status", handleSteamStatus)
	http.HandleFunc("/api/steam/conduct-scorecard", handleConductScorecard)
	http.HandleFunc("/api/steam/validate-report-card", handleValidateReportCard)
	http.HandleFunc("/api/steam/validate-report-card-current", handleValidateReportCardCurrent)
	http.HandleFunc("/api/steam/validate-report-card-status", handleValidateReportCardStatus)

	fmt.Println("Server started at http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
