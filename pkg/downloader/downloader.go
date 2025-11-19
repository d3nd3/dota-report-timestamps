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
	"time"

	"github.com/klauspost/compress/zstd"
)

var httpClient = &http.Client{
	Timeout: 30 * time.Second,
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func DownloadReplay(matchID int64, replayDir string) error {
	replayURL, err := getReplayURL(matchID)
	if err != nil {
		return fmt.Errorf("failed to get replay URL: %w", err)
	}

	if replayURL == "" {
		return fmt.Errorf("replay URL not available for match %d (match may not be parsed yet)", matchID)
	}

	return downloadAndExtractReplay(replayURL, matchID, replayDir)
}

func RequestParsing(matchID int64) (int, error) {
	url := fmt.Sprintf("https://api.opendota.com/api/request/%d", matchID)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("failed to request parsing: %w", err)
	}
	defer resp.Body.Close()

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
	}

	if err := json.Unmarshal(body, &jobResp); err != nil {
		return 0, fmt.Errorf("failed to unmarshal job response: %w", err)
	}

	return jobResp.Job.JobID, nil
}

func GetReplayURL(matchID int64) (string, error) {
	return getReplayURL(matchID)
}

func getReplayURL(matchID int64) (string, error) {
	url := fmt.Sprintf("https://api.opendota.com/api/matches/%d", matchID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7")
	req.Header.Set("Accept-Encoding", "gzip, zstd")
	req.Header.Set("Accept-Language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("DNT", "1")
	req.Header.Set("Sec-Ch-Ua", `"Google Chrome";v="131", "Chromium";v="131", "Not_A Brand";v="24"`)
	req.Header.Set("Sec-Ch-Ua-Mobile", "?0")
	req.Header.Set("Sec-Ch-Ua-Platform", `"Linux"`)
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Priority", "u=0,i")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch match data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	contentEncoding := resp.Header.Get("Content-Encoding")

	// OpenDota often uses zstd compression even if Content-Encoding header is not set
	// Read the body first, then try to decompress if needed
	rawContent, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", fmt.Errorf("failed to read response body: %w", readErr)
	}

	var bodyContent []byte

	// Check Content-Encoding header first
	switch contentEncoding {
	case "zstd":
		zstdReader, err := zstd.NewReader(bytes.NewReader(rawContent))
		if err != nil {
			return "", fmt.Errorf("failed to create zstd reader: %w", err)
		}
		defer zstdReader.Close()
		bodyContent, err = io.ReadAll(zstdReader)
		if err != nil {
			return "", fmt.Errorf("failed to read zstd decompressed body: %w", err)
		}
	case "gzip":
		gzReader, err := gzip.NewReader(bytes.NewReader(rawContent))
		if err != nil {
			return "", fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzReader.Close()
		bodyContent, err = io.ReadAll(gzReader)
		if err != nil {
			return "", fmt.Errorf("failed to read gzip decompressed body: %w", err)
		}
	default:
		// No Content-Encoding header - try to detect compression
		// Check if it looks like compressed data (doesn't start with { or [)
		if len(rawContent) > 0 && rawContent[0] != '{' && rawContent[0] != '[' {
			// Try zstd decompression first (OpenDota's default)
			zstdReader, zstdErr := zstd.NewReader(bytes.NewReader(rawContent))
			if zstdErr == nil {
				defer zstdReader.Close()
				decompressed, readErr := io.ReadAll(zstdReader)
				if readErr == nil && len(decompressed) > 0 && (decompressed[0] == '{' || decompressed[0] == '[') {
					bodyContent = decompressed
				} else {
					// zstd read failed, try gzip
					gzReader, gzErr := gzip.NewReader(bytes.NewReader(rawContent))
					if gzErr == nil {
						defer gzReader.Close()
						decompressed, readErr := io.ReadAll(gzReader)
						if readErr == nil && len(decompressed) > 0 && (decompressed[0] == '{' || decompressed[0] == '[') {
							bodyContent = decompressed
						} else {
							return "", fmt.Errorf("response appears compressed but decompression failed (tried zstd and gzip), first byte: 0x%02x", rawContent[0])
						}
					} else {
						return "", fmt.Errorf("response appears compressed but decompression failed (tried zstd and gzip), first byte: 0x%02x", rawContent[0])
					}
				}
			} else {
				// zstd reader creation failed, try gzip
				gzReader, gzErr := gzip.NewReader(bytes.NewReader(rawContent))
				if gzErr == nil {
					defer gzReader.Close()
					decompressed, readErr := io.ReadAll(gzReader)
					if readErr == nil && len(decompressed) > 0 && (decompressed[0] == '{' || decompressed[0] == '[') {
						bodyContent = decompressed
					} else {
						return "", fmt.Errorf("response appears compressed but gzip decompression failed, first byte: 0x%02x", rawContent[0])
					}
				} else {
					return "", fmt.Errorf("response appears compressed but cannot create decompression readers, first byte: 0x%02x", rawContent[0])
				}
			}
		} else {
			// Looks like JSON already
			bodyContent = rawContent
		}
	}

	// Check if response looks like JSON (starts with { or [)
	if len(bodyContent) > 0 && bodyContent[0] != '{' && bodyContent[0] != '[' {
		// Try to find JSON in the response (might be embedded in HTML)
		start := -1
		for i := 0; i < len(bodyContent)-1; i++ {
			if bodyContent[i] == '{' {
				start = i
				break
			}
		}
		if start == -1 {
			return "", fmt.Errorf("response is not JSON, content-type: %s, first 100 bytes: %q", resp.Header.Get("Content-Type"), string(bodyContent[:min(100, len(bodyContent))]))
		}
		bodyContent = bodyContent[start:]
	}

	var apiResp struct {
		ReplayURL string `json:"replay_url"`
		OdData    struct {
			HasParsed bool `json:"has_parsed"`
		} `json:"od_data"`
	}

	if err := json.Unmarshal(bodyContent, &apiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response (content-type: %s, first 200 bytes: %q): %w", resp.Header.Get("Content-Type"), string(bodyContent[:min(200, len(bodyContent))]), err)
	}

	if apiResp.ReplayURL != "" {
		return apiResp.ReplayURL, nil
	}

	if !apiResp.OdData.HasParsed {
		return "", nil
	}

	return apiResp.ReplayURL, nil
}

func downloadAndExtractReplay(replayURL string, matchID int64, replayDir string) error {
	if err := os.MkdirAll(replayDir, os.ModePerm); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	bz2FilePath := filepath.Join(replayDir, fmt.Sprintf("%d.bz2", matchID))

	err := func() error {
		getResp, err := httpClient.Get(replayURL)
		if err != nil {
			return fmt.Errorf("failed to download replay: %w", err)
		}
		defer getResp.Body.Close()

		if getResp.StatusCode != http.StatusOK {
			return fmt.Errorf("download failed with status: %s", getResp.Status)
		}

		bz2File, err := os.Create(bz2FilePath)
		if err != nil {
			return fmt.Errorf("failed to create .bz2 file: %w", err)
		}
		defer bz2File.Close()

		if _, err := io.Copy(bz2File, getResp.Body); err != nil {
			return fmt.Errorf("failed to save .bz2 file: %w", err)
		}
		return nil
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

	return nil
}

func WaitForParsing(matchID int64, jobID int, maxWaitTime time.Duration) error {
	deadline := time.Now().Add(maxWaitTime)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Check job status first
			if jobID > 0 {
				jobURL := fmt.Sprintf("https://api.opendota.com/api/request/%d", jobID)
				jobResp, err := httpClient.Get(jobURL)
				if err == nil {
					jobResp.Body.Close()
				}
			}

			// Check if match is parsed by checking for replay URL
			replayURL, err := getReplayURL(matchID)
			if err != nil {
				log.Printf("Error checking replay URL for match %d: %v", matchID, err)
				continue
			}
			if replayURL != "" {
				return nil
			}
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for match %d to be parsed", matchID)
			}
		case <-time.After(maxWaitTime):
			return fmt.Errorf("timeout waiting for match %d to be parsed", matchID)
		}
	}
}
