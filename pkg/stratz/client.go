package stratz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/shurcooL/graphql"
)

type Client struct {
	client *graphql.Client
	token  string
}

func NewClient(token string) *Client {
	httpClient := &http.Client{
		Transport: &authedTransport{
			token:   token,
			wrapped: http.DefaultTransport,
		},
	}
	return &Client{
		client: graphql.NewClient("https://api.stratz.com/graphql", httpClient),
		token:  token,
	}
}

type authedTransport struct {
	token   string
	wrapped http.RoundTripper
}

func (t *authedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.token)
	req.Header.Set("User-Agent", "STRATZ_API")
	return t.wrapped.RoundTrip(req)
}

type Match struct {
	ID int64 `json:"id"`
}

type ReplayInfo struct {
	ClusterID  int   `json:"clusterId"`
	ReplaySalt int64 `json:"replaySalt"`
}

func (c *Client) GetReplayInfo(matchID int64) (*ReplayInfo, error) {
	queryStr := `query GetReplayInfo($matchId: Long!) {
		match(id: $matchId) {
			clusterId
			replaySalt
		}
	}`

	variables := map[string]interface{}{
		"matchId": matchID,
	}

	reqBody := map[string]interface{}{
		"query":     queryStr,
		"variables": variables,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", "https://api.stratz.com/graphql", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", "STRATZ_API")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("stratz API error: non-200 status %d", resp.StatusCode)
	}

	var result struct {
		Data struct {
			Match struct {
				ClusterID  int   `json:"clusterId"`
				ReplaySalt int64 `json:"replaySalt"`
			} `json:"match"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %v", result.Errors)
	}

	return &ReplayInfo{
		ClusterID:  result.Data.Match.ClusterID,
		ReplaySalt: result.Data.Match.ReplaySalt,
	}, nil
}

func (c *Client) GetLastMatches(steamID int64, limit int, turboOnly bool) ([]Match, error) {
	var queryStr string
	if turboOnly {
		queryStr = `query GetMatches($steamAccountId: Long!, $take: Int!) {
			player(steamAccountId: $steamAccountId) {
				matches(request: {take: $take, gameModes: [23]}) {
					id
				}
			}
		}`
	} else {
		queryStr = `query GetMatches($steamAccountId: Long!, $take: Int!) {
			player(steamAccountId: $steamAccountId) {
				matches(request: {take: $take}) {
					id
				}
			}
		}`
	}

	variables := map[string]interface{}{
		"steamAccountId": steamID,
		"take":           limit,
	}

	reqBody := map[string]interface{}{
		"query":     queryStr,
		"variables": variables,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(context.Background(), "POST", "https://api.stratz.com/graphql", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("User-Agent", "STRATZ_API")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		// Often Stratz returns 500 for invalid SteamID format, but let's just log it clearly
		return nil, fmt.Errorf("stratz API error (steamID: %d, limit: %d): non-200 OK status code: %d %s body: %s", steamID, limit, resp.StatusCode, resp.Status, string(body))
	}

	var result struct {
		Data struct {
			Player *struct { // Player can be null if ID is invalid or hidden
				Matches []struct {
					ID int64 `json:"id"`
				} `json:"matches"`
			} `json:"player"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	if len(result.Errors) > 0 {
		return nil, fmt.Errorf("graphql errors: %v", result.Errors)
	}

	if result.Data.Player == nil {
		return nil, fmt.Errorf("player not found (SteamID: %d)", steamID)
	}

	matches := make([]Match, len(result.Data.Player.Matches))
	for i, m := range result.Data.Player.Matches {
		matches[i] = Match{ID: m.ID}
	}

	return matches, nil
}
