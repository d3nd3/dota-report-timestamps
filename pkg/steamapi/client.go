package steamapi

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

type MatchDetailsResponse struct {
	Result struct {
		Cluster    int   `json:"cluster"`
		MatchID    int64 `json:"match_id"`
		ReplaySalt int64 `json:"replay_salt"`
		Error      string `json:"error"`
	} `json:"result"`
}

func (c *Client) GetReplayInfo(matchID int64) (clusterID int, replaySalt int64, err error) {
	url := fmt.Sprintf("https://api.steampowered.com/IDOTA2Match_570/GetMatchDetails/v1/?key=%s&match_id=%d", c.apiKey, matchID)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to fetch match details: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden {
		return 0, 0, fmt.Errorf("invalid Steam API key")
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// Special handling for known 500 error on new matches
		if resp.StatusCode == http.StatusInternalServerError {
			return 0, 0, fmt.Errorf("Valve WebAPI returned 500 (known issue for post-7.36 matches): %s", string(body))
		}
		return 0, 0, fmt.Errorf("unexpected status code from Steam API: %d, body: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var details MatchDetailsResponse
	if err := json.Unmarshal(body, &details); err != nil {
		return 0, 0, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if details.Result.Error != "" {
		return 0, 0, fmt.Errorf("steam api error: %s", details.Result.Error)
	}

	// Check if we got valid data
	// Note: Sometimes Steam API returns empty result for new matches or matches without replay salt
	if details.Result.MatchID == 0 {
		return 0, 0, fmt.Errorf("match not found or details unavailable")
	}
	
	// Some matches might not have replay salt (e.g. too old, or practice lobbies)
	if details.Result.ReplaySalt == 0 {
		return 0, 0, fmt.Errorf("replay salt not found in match details")
	}

	return details.Result.Cluster, details.Result.ReplaySalt, nil
}

