package downloader

import (
	"bytes"
	"compress/bzip2"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/d3nd3/dota-report-timestamps/pkg/botclient"
	"github.com/d3nd3/dota-report-timestamps/pkg/parser"
	"github.com/d3nd3/dota-report-timestamps/pkg/steamapi"
	"github.com/d3nd3/dota-report-timestamps/pkg/stratz"
	"github.com/klauspost/compress/zstd"
)

// ProgressCallback is a function type for reporting download progress (0-100)
type ProgressCallback func(matchID int64, progress float64)

// Global map to store progress updates
var (
	downloadProgress = make(map[int64]float64)
	progressMu       sync.RWMutex
)

// GetProgress returns the current download progress for a match
func GetProgress(matchID int64) float64 {
	progressMu.RLock()
	defer progressMu.RUnlock()
	return downloadProgress[matchID]
}

// SetProgress updates the progress for a match
func SetProgress(matchID int64, progress float64) {
	progressMu.Lock()
	defer progressMu.Unlock()
	downloadProgress[matchID] = progress
}

// ClearProgress removes progress for a match (e.g. when done)
func ClearProgress(matchID int64) {
	progressMu.Lock()
	defer progressMu.Unlock()
	delete(downloadProgress, matchID)
}

// ProgressWriter counts bytes written and reports progress
type ProgressWriter struct {
	Total      int64
	Written    int64
	MatchID    int64
	LastUpdate time.Time
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	n := len(p)
	pw.Written += int64(n)

	// Update progress at most every 100ms to avoid lock contention
	if time.Since(pw.LastUpdate) > 100*time.Millisecond {
		if pw.Total > 0 {
			percentage := float64(pw.Written) / float64(pw.Total) * 100
			SetProgress(pw.MatchID, percentage)
		}
		pw.LastUpdate = time.Now()
	}
	return n, nil
}

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

var downloadClient = &http.Client{
	Timeout: 10 * time.Minute,
}

type pendingMatch struct {
	matchID     int64
	jobID       int
	replayDir   string
	stratzToken string
	steamAPIKey string
	gcClient    *botclient.Client
	requestedAt time.Time
}

var (
	pendingMatches = make(map[int64]*pendingMatch)
	pendingMu      sync.RWMutex
)

// Rate limiting: OpenDota free tier is ~60 calls/minute.
// Special rule: /api/request POST calls count as 10 calls.
// We'll use a token bucket where 1 token = 1 call cost.
// Capacity 60, refill 1 per second.
var rateLimiter = make(chan struct{}, 60)

func init() {
	// Fill the bucket initially
	for i := 0; i < 60; i++ {
		rateLimiter <- struct{}{}
	}

	// Refill routine: 1 token every second (60/min)
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			select {
			case rateLimiter <- struct{}{}:
			default:
				// Bucket full
			}
		}
	}()

	// Background worker to process pending matches
	go processPendingMatches()
}

// waitForCost blocks until we can consume 'cost' tokens.
func waitForCost(cost int) {
	for i := 0; i < cost; i++ {
		<-rateLimiter
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// constructReplayURLs generates primary and alternative replay URLs for a given match.
// Valve has multiple replay clusters.
// Primary format: http://replay{cluster}.valve.net/570/{matchID}_{salt}.dem.bz2
// Perfect World (China) format: http://replay{cluster}.wmsj.cn/570/{matchID}_{salt}.dem.bz2
func constructReplayURLs(cluster uint32, matchID int64, salt uint64) []string {
	// Primary URL (Valve global)
	urls := []string{
		fmt.Sprintf("http://replay%d.valve.net/570/%d_%d.dem.bz2", cluster, matchID, salt),
	}

	// If the cluster is in China (typically > 200, but can vary), Perfect World servers might be more reliable.
	// We'll add it as a fallback for all clusters just in case, or specifically if we suspect it's a PW match.
	// Note: Perfect World domains sometimes change, but wmsj.cn is the standard one.
	urls = append(urls, fmt.Sprintf("http://replay%d.wmsj.cn/570/%d_%d.dem.bz2", cluster, matchID, salt))

	return urls
}

// DownloadReplay downloads a replay for the given match ID to the specified directory.
// It tries to fetch the replay URL using Stratz (if a token is provided) or falls back to OpenDota.
func DownloadReplay(matchID int64, replayDir string, stratzToken string, steamAPIKey string, gcClient *botclient.Client) error {
	demFilePath := filepath.Join(replayDir, fmt.Sprintf("%d.dem", matchID))
	if _, err := os.Stat(demFilePath); err == nil {
		log.Printf("Replay file already exists for match %d, skipping download", matchID)
		return nil
	}

	var replayURLs []string
	// var err error // Removed this declaration to use the one from the loop/calls properly

	// 0. Try Dota 2 GC Client (Highest Priority, Direct Access)
	status := botclient.StatusDisconnected
	if gcClient != nil {
		status = gcClient.GetStatus()
	}

	if gcClient != nil && (status == botclient.StatusGCReady || status == botclient.StatusConnected) {
		log.Printf("Attempting to get replay URL from Dota 2 GC for match %d (Status: %d)...", matchID, status)
		cluster, salt, err := gcClient.GetReplayInfo(uint64(matchID))
		if err == nil && cluster > 0 && salt > 0 {
			replayURLs = constructReplayURLs(cluster, matchID, salt)
			log.Printf("Found replay URL via Dota 2 GC: %s (and %d alternates)", replayURLs[0], len(replayURLs)-1)
		} else {
			log.Printf("Dota 2 GC failed: %v. Falling back to other methods.", err)
		}
	} else if gcClient != nil {
		log.Printf("Skipping Dota 2 GC: Status is %d (Expected %d - GCReady/Connected). Bot might be connecting or stuck.", status, botclient.StatusGCReady)
	}

	// 1. Try Steam WebAPI first if key is available (Most reliable/authoritative)
	if len(replayURLs) == 0 && steamAPIKey != "" {
		log.Printf("Attempting to get replay URL from Steam WebAPI for match %d...", matchID)
		steamClient := steamapi.NewClient(steamAPIKey)
		clusterID, replaySalt, err := steamClient.GetReplayInfo(matchID)
		if err != nil {
			log.Printf("Failed to get replay URL from Steam WebAPI: %v. Falling back to Stratz.", err)
		} else if clusterID > 0 && replaySalt > 0 {
			replayURLs = constructReplayURLs(uint32(clusterID), matchID, uint64(replaySalt))
			log.Printf("Found replay URL via Steam WebAPI: %s", replayURLs[0])
		}
	}

	// 2. Try Stratz if Steam failed or no key
	if len(replayURLs) == 0 && stratzToken != "" {
		log.Printf("Attempting to get replay URL from Stratz for match %d...", matchID)
		stratzURL, err := getReplayURLFromStratz(matchID, stratzToken)
		if err != nil {
			log.Printf("Failed to get replay URL from Stratz: %v. Falling back to OpenDota.", err)
		} else if stratzURL != "" {
			// Stratz returns a full URL, so we accept it as is.
			// However, Stratz might construct it themselves.
			// Ideally we want cluster/salt to construct fallbacks, but Stratz client just gives URL here.
			// We'll just use the URL provided.
			replayURLs = []string{stratzURL}
			log.Printf("Found replay URL via Stratz: %s", stratzURL)
		}
	}

	// 3. Fallback to OpenDota if previous methods failed
	if len(replayURLs) == 0 {
		log.Printf("Attempting to get replay URL from OpenDota for match %d...", matchID)

		// First check if match is parsed using has_parsed field (most reliable)
		hasParsed, err := checkOpenDotaParsed(matchID)
		if err != nil {
			// If OpenDota is down (521) or having server issues (5xx), queue for retry
			if strings.Contains(err.Error(), "status code 521") || strings.Contains(err.Error(), "status code 5") {
				log.Printf("OpenDota is temporarily unavailable for match %d, queueing for background retry...", matchID)
				pendingMu.Lock()
				pendingMatches[matchID] = &pendingMatch{
					matchID:     matchID,
					jobID:       0, // No job ID yet, will check again later
					replayDir:   replayDir,
					stratzToken: stratzToken,
					steamAPIKey: steamAPIKey,
					gcClient:    gcClient,
					requestedAt: time.Now(),
				}
				pendingMu.Unlock()
				return fmt.Errorf("match %d queued for retry (OpenDota temporarily unavailable)", matchID)
			}
			return fmt.Errorf("failed to check OpenDota parsed status: %w", err)
		}

		if !hasParsed {
			log.Printf("Match %d not parsed yet on OpenDota, requesting parsing...", matchID)
			jobID, err := RequestParsing(matchID)
			if err != nil {
				// If OpenDota is down (521) or having server issues (5xx), queue for retry
				if strings.Contains(err.Error(), "status code 521") || strings.Contains(err.Error(), "status code 5") {
					log.Printf("OpenDota is temporarily unavailable for match %d, queueing for background retry...", matchID)
					pendingMu.Lock()
					pendingMatches[matchID] = &pendingMatch{
						matchID:     matchID,
						jobID:       0, // No job ID yet, will check again later
						replayDir:   replayDir,
						stratzToken: stratzToken,
						steamAPIKey: steamAPIKey,
						gcClient:    gcClient,
						requestedAt: time.Now(),
					}
					pendingMu.Unlock()
					return fmt.Errorf("match %d queued for retry (OpenDota temporarily unavailable)", matchID)
				}
				return fmt.Errorf("failed to request parsing: %w", err)
			}
			log.Printf("Parsing requested for match %d (Job ID: %d), queueing for background processing...", matchID, jobID)

			// Queue for background processing instead of blocking
			pendingMu.Lock()
			pendingMatches[matchID] = &pendingMatch{
				matchID:     matchID,
				jobID:       jobID,
				replayDir:   replayDir,
				stratzToken: stratzToken,
				steamAPIKey: steamAPIKey,
				gcClient:    gcClient,
				requestedAt: time.Now(),
			}
			pendingMu.Unlock()

			return fmt.Errorf("match %d queued for parsing, will be processed in background", matchID)
		}

		// Now fetch the replay URL (should be available if parsed)
		odURL, err := getReplayURL(matchID)
		if err != nil {
			// If OpenDota is down (521) or having server issues (5xx), queue for retry
			if strings.Contains(err.Error(), "status code 521") || strings.Contains(err.Error(), "status code 5") {
				log.Printf("OpenDota is temporarily unavailable for match %d, queueing for background retry...", matchID)
				pendingMu.Lock()
				pendingMatches[matchID] = &pendingMatch{
					matchID:     matchID,
					jobID:       0, // No job ID yet, will check again later
					replayDir:   replayDir,
					stratzToken: stratzToken,
					steamAPIKey: steamAPIKey,
					gcClient:    gcClient,
					requestedAt: time.Now(),
				}
				pendingMu.Unlock()
				return fmt.Errorf("match %d queued for retry (OpenDota temporarily unavailable)", matchID)
			}
			return fmt.Errorf("failed to get replay URL from OpenDota: %w", err)
		}
		if odURL == "" {
			return fmt.Errorf("replay URL is missing for match %d (may have expired)", matchID)
		}
		replayURLs = []string{odURL}
	}

	return downloadAndExtractReplay(replayURLs, matchID, replayDir)
}

func getReplayURLFromStratz(matchID int64, token string) (string, error) {
	client := stratz.NewClient(token)
	info, err := client.GetReplayInfo(matchID)
	if err != nil {
		return "", err
	}

	if info.ClusterID == 0 || info.ReplaySalt == 0 {
		// Log more details about why we're missing this data
		if info.ClusterID == 0 && info.ReplaySalt == 0 {
			return "", fmt.Errorf("missing cluster or salt info from Stratz (match %d may not have replay data available yet)", matchID)
		} else if info.ClusterID == 0 {
			return "", fmt.Errorf("missing cluster ID from Stratz for match %d (replaySalt: %d)", matchID, info.ReplaySalt)
		} else {
			return "", fmt.Errorf("missing replay salt from Stratz for match %d (clusterID: %d)", matchID, info.ClusterID)
		}
	}

	url := fmt.Sprintf("http://replay%d.valve.net/570/%d_%d.dem.bz2", info.ClusterID, matchID, info.ReplaySalt)
	return url, nil
}

func RequestParsing(matchID int64) (int, error) {
	url := fmt.Sprintf("https://api.opendota.com/api/request/%d", matchID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Cost is 10 tokens for a POST request
	waitForCost(10)
	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to request parsing: %w", err)
	}
	defer resp.Body.Close()

	// Handle Cloudflare 521 (Web Server Is Down) and other transient errors
	if resp.StatusCode == 521 {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("unexpected status code 521: OpenDota server is down (Cloudflare error). Body: %s", string(body))
	}

	// Handle other 5xx errors as potentially transient
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("unexpected status code %d: OpenDota server error (may be temporary). Body: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var jobResp struct {
		Job struct {
			JobID int `json:"jobId"`
		} `json:"job"`
		JobID int `json:"jobId"` // Sometimes it's top level
	}

	if err := json.Unmarshal(body, &jobResp); err != nil {
		return 0, fmt.Errorf("failed to unmarshal job response: %w", err)
	}

	id := jobResp.Job.JobID
	if id == 0 {
		id = jobResp.JobID
	}

	return id, nil
}

func GetReplayURL(matchID int64) (string, error) {
	return getReplayURL(matchID)
}

// checkOpenDotaParsed checks if a match has been parsed on OpenDota by checking the has_parsed field
func checkOpenDotaParsed(matchID int64) (bool, error) {
	url := fmt.Sprintf("https://api.opendota.com/api/matches/%d", matchID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Encoding", "gzip, zstd")

	waitForCost(1)
	resp, err := httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to fetch match data: %w", err)
	}
	defer resp.Body.Close()

	// Handle Cloudflare 521 (Web Server Is Down) and other transient errors
	if resp.StatusCode == 521 {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("unexpected status code 521: OpenDota server is down (Cloudflare error). Body: %s", string(body))
	}

	// Handle other 5xx errors as potentially transient
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("unexpected status code %d: OpenDota server error (may be temporary). Body: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return false, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	rawContent, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return false, fmt.Errorf("failed to read response body: %w", readErr)
	}

	// Handle compression (zstd, gzip, or none)
	var bodyContent []byte
	contentEncoding := resp.Header.Get("Content-Encoding")

	switch contentEncoding {
	case "zstd":
		if r, err := zstd.NewReader(bytes.NewReader(rawContent)); err == nil {
			defer r.Close()
			bodyContent, _ = io.ReadAll(r)
		}
	case "gzip":
		if r, err := gzip.NewReader(bytes.NewReader(rawContent)); err == nil {
			defer r.Close()
			bodyContent, _ = io.ReadAll(r)
		}
	default:
		// Auto-detect zstd magic bytes if no header
		if len(rawContent) > 0 && rawContent[0] != '{' && rawContent[0] != '[' {
			if r, err := zstd.NewReader(bytes.NewReader(rawContent)); err == nil {
				defer r.Close()
				if d, err := io.ReadAll(r); err == nil && len(d) > 0 {
					bodyContent = d
				}
			}
		}
	}

	if len(bodyContent) == 0 {
		bodyContent = rawContent
	}

	// Strip non-JSON prefix if any
	if len(bodyContent) > 0 && bodyContent[0] != '{' && bodyContent[0] != '[' {
		start := bytes.IndexByte(bodyContent, '{')
		if start != -1 {
			bodyContent = bodyContent[start:]
		}
	}

	var apiResp struct {
		OdData struct {
			HasParsed bool `json:"has_parsed"`
		} `json:"od_data"`
	}

	if err := json.Unmarshal(bodyContent, &apiResp); err != nil {
		// If we can't parse, assume not parsed
		return false, nil
	}

	return apiResp.OdData.HasParsed, nil
}

func getReplayURL(matchID int64) (string, error) {
	url := fmt.Sprintf("https://api.opendota.com/api/matches/%d", matchID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Encoding", "gzip, zstd")

	// Cost is 1 token for GET
	waitForCost(1)
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch match data: %w", err)
	}
	defer resp.Body.Close()

	// Handle Cloudflare 521 (Web Server Is Down) and other transient errors
	if resp.StatusCode == 521 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code 521: OpenDota server is down (Cloudflare error). Body: %s", string(body))
	}

	// Handle other 5xx errors as potentially transient
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code %d: OpenDota server error (may be temporary). Body: %s", resp.StatusCode, string(body))
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	rawContent, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read response body: %w", readErr)
	}

	// Handle compression (zstd, gzip, or none)
	var bodyContent []byte
	contentEncoding := resp.Header.Get("Content-Encoding")

	switch contentEncoding {
	case "zstd":
		if r, err := zstd.NewReader(bytes.NewReader(rawContent)); err == nil {
			defer r.Close()
			bodyContent, _ = io.ReadAll(r)
		}
	case "gzip":
		if r, err := gzip.NewReader(bytes.NewReader(rawContent)); err == nil {
			defer r.Close()
			bodyContent, _ = io.ReadAll(r)
		}
	default:
		// Auto-detect zstd magic bytes if no header
		if len(rawContent) > 0 && rawContent[0] != '{' && rawContent[0] != '[' {
			if r, err := zstd.NewReader(bytes.NewReader(rawContent)); err == nil {
				defer r.Close()
				if d, err := io.ReadAll(r); err == nil && len(d) > 0 {
					bodyContent = d
				}
			}
		}
	}

	if len(bodyContent) == 0 {
		bodyContent = rawContent
	}

	// Strip non-JSON prefix if any
	if len(bodyContent) > 0 && bodyContent[0] != '{' && bodyContent[0] != '[' {
		start := bytes.IndexByte(bodyContent, '{')
		if start != -1 {
			bodyContent = bodyContent[start:]
		}
	}

	var apiResp struct {
		ReplayURL string `json:"replay_url"`
		OdData    struct {
			HasParsed bool `json:"has_parsed"`
		} `json:"od_data"`
	}

	if err := json.Unmarshal(bodyContent, &apiResp); err != nil {
		// Don't fail hard on unmarshal, just return empty
		return "", nil
	}

	if apiResp.ReplayURL != "" {
		return apiResp.ReplayURL, nil
	}

	// If parsed but no URL, it might have expired
	return "", nil
}

func downloadAndExtractReplay(replayURLs []string, matchID int64, replayDir string) error {
	if err := os.MkdirAll(replayDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	bz2FilePath := filepath.Join(replayDir, fmt.Sprintf("%d.bz2", matchID))

	err := func() error {
		// Try each URL in the list
		var lastErr error
		for _, url := range replayURLs {
			log.Printf("Downloading replay from: %s", url)

			// Retry loop for 502/503 errors (max 3 retries per URL)
			for i := 0; i < 3; i++ {
				if i > 0 {
					time.Sleep(2 * time.Second) // Backoff slightly
				}

				getResp, reqErr := downloadClient.Get(url)
				if reqErr != nil {
					lastErr = reqErr
					log.Printf("Network error downloading %s: %v", url, reqErr)
					continue // Network error, retry
				}

				if getResp.StatusCode == http.StatusOK {
					// Success! Save and return
					defer getResp.Body.Close()

					// Initialize progress
					SetProgress(matchID, 0)
					defer ClearProgress(matchID)

					bz2File, err := os.Create(bz2FilePath)
					if err != nil {
						return fmt.Errorf("failed to create .bz2 file: %w", err)
					}
					defer bz2File.Close()

					// Wrap response body with progress writer
					contentLength := getResp.ContentLength
					progressWriter := &ProgressWriter{
						Total:   contentLength,
						MatchID: matchID,
					}

					// Use TeeReader to write to file and update progress
					// Note: io.Copy uses the writer (bz2File) to write, but we need to intercept reads.
					// Actually, we can wrap the reader.
					// But wait, ProgressWriter implements Write, so we can use io.TeeReader(body, progressWriter)
					// No, io.TeeReader returns a Reader. We need to Copy from (TeeReader(body, progressWriter)) to file.
					// Wait, ProgressWriter is a Writer. So TeeReader(body, progressWriter) returns a reader that writes to progressWriter as it reads.

					reader := io.TeeReader(getResp.Body, progressWriter)

					if _, err := io.Copy(bz2File, reader); err != nil {
						return fmt.Errorf("failed to save .bz2 file: %w", err)
					}

					// Ensure 100% at end
					SetProgress(matchID, 100)
					return nil
				}

				// If it's a temporary server error (502, 503, 504), retry
				if getResp.StatusCode == http.StatusBadGateway || getResp.StatusCode == http.StatusServiceUnavailable || getResp.StatusCode == http.StatusGatewayTimeout {
					getResp.Body.Close()
					lastErr = fmt.Errorf("download failed with status: %s", getResp.Status)
					log.Printf("Server error downloading %s: %s", url, getResp.Status)
					continue
				}

				// Permanent error (404, etc), try next URL immediately
				getResp.Body.Close()
				lastErr = fmt.Errorf("download failed with status: %s", getResp.Status)
				log.Printf("Permanent error downloading %s: %s", url, getResp.Status)
				break
			}
		}

		return fmt.Errorf("failed to download replay after trying all URLs: %w", lastErr)
	}()

	if err != nil {
		return err
	}

	demFilePath := filepath.Join(replayDir, fmt.Sprintf("%d.dem", matchID))

	err = func() error {
		bz2FileReader, err := os.Open(bz2FilePath)
		if err != nil {
			return fmt.Errorf("failed to open .bz2 file: %w", err)
		}
		defer bz2FileReader.Close()

		bzip2Reader := bzip2.NewReader(bz2FileReader)

		demFile, err := os.Create(demFilePath)
		if err != nil {
			return fmt.Errorf("failed to create .dem file: %w", err)
		}
		defer demFile.Close()

		if _, err := io.Copy(demFile, bzip2Reader); err != nil {
			return fmt.Errorf("failed to write decompressed data: %w", err)
		}
		return nil
	}()

	if err != nil {
		return err
	}

	if err := os.Remove(bz2FilePath); err != nil {
		log.Printf("Warning: failed to remove temp file %s: %v", bz2FilePath, err)
	}

	demFile, err := os.Open(demFilePath)
	if err == nil {
		if date, err := parser.GetReplayDate(demFile); err == nil {
			log.Printf("Extracted match date for match %d: %v", matchID, date)
		} else {
			log.Printf("Could not extract date from replay %d: %v (using file mod time)", matchID, err)
		}
		demFile.Close()
	}

	return nil
}

func WaitForParsing(matchID int64, jobID int, maxWaitTime time.Duration) error {
	deadline := time.Now().Add(maxWaitTime)
	// Poll every 10s (polite)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	log.Printf("Entering wait loop for match %d (Job ID: %d), max wait: %v", matchID, jobID, maxWaitTime)

	for {
		select {
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for match %d", matchID)
			}

			// 1. Poll job status first
			if jobID > 0 {
				url := fmt.Sprintf("https://api.opendota.com/api/request/%d", jobID)
				req, _ := http.NewRequest("GET", url, nil)
				waitForCost(1)
				resp, err := httpClient.Do(req)
				if err == nil {
					resp.Body.Close()
				}
			}

			// 2. Check if match is ready via main endpoint using has_parsed (most reliable)
			hasParsed, err := checkOpenDotaParsed(matchID)
			if err != nil {
				log.Printf("Error checking parsed status for match %d: %v", matchID, err)
				continue
			}

			if hasParsed {
				// Match is parsed, now check for replay URL
				replayURL, err := getReplayURL(matchID)
				if err == nil && replayURL != "" {
					log.Printf("Match %d parsed! Replay URL: %s", matchID, replayURL)
					return nil
				}
				// If parsed but no URL, might have expired - but parsing is done
				log.Printf("Match %d is parsed but replay URL is missing (may have expired)", matchID)
				return nil
			}

			log.Printf("Match %d still pending...", matchID)

		case <-time.After(maxWaitTime):
			return fmt.Errorf("timeout waiting for match %d", matchID)
		}
	}
}

func processPendingMatches() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		pendingMu.Lock()
		var toProcess []*pendingMatch
		for matchID, pm := range pendingMatches {
			if time.Since(pm.requestedAt) > 10*time.Minute {
				log.Printf("Removing expired pending match %d", matchID)
				delete(pendingMatches, matchID)
				continue
			}
			toProcess = append(toProcess, pm)
		}
		pendingMu.Unlock()

		for _, pm := range toProcess {
			hasParsed, err := checkOpenDotaParsed(pm.matchID)
			if err != nil {
				// If OpenDota is still down, log and continue (will retry next cycle)
				if strings.Contains(err.Error(), "status code 521") || strings.Contains(err.Error(), "status code 5") {
					log.Printf("OpenDota still unavailable for pending match %d: %v (will retry)", pm.matchID, err)
					continue
				}
				log.Printf("Error checking parsed status for pending match %d: %v", pm.matchID, err)
				continue
			}

			if hasParsed {
				log.Printf("Pending match %d is now parsed, checking if already exists...", pm.matchID)

				demFilePath := filepath.Join(pm.replayDir, fmt.Sprintf("%d.dem", pm.matchID))
				if _, err := os.Stat(demFilePath); err == nil {
					log.Printf("Replay file already exists for match %d, skipping download", pm.matchID)
					pendingMu.Lock()
					delete(pendingMatches, pm.matchID)
					pendingMu.Unlock()
					continue
				}

				log.Printf("Completing download for match %d...", pm.matchID)
				pendingMu.Lock()
				delete(pendingMatches, pm.matchID)
				pendingMu.Unlock()

				replayURL, err := getReplayURL(pm.matchID)
				if err != nil {
					log.Printf("Error getting replay URL for match %d: %v", pm.matchID, err)
					continue
				}
				if replayURL == "" {
					log.Printf("Replay URL missing for match %d (may have expired)", pm.matchID)
					continue
				}

				if err := downloadAndExtractReplay([]string{replayURL}, pm.matchID, pm.replayDir); err != nil {
					log.Printf("Error downloading replay for match %d: %v", pm.matchID, err)
				} else {
					log.Printf("Successfully downloaded replay for match %d", pm.matchID)
				}
			} else {
				log.Printf("Pending match %d still waiting for parsing...", pm.matchID)
			}
		}
	}
}
