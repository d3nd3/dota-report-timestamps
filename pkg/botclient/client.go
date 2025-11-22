package botclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type ConnectionStatus int

const (
	StatusDisconnected ConnectionStatus = iota
	StatusConnecting
	StatusNeedGuardCode
	StatusConnected
	StatusGCReady
	StatusRateLimited
)

type Client struct {
	baseURL string
}

func NewClient(port string) *Client {
	if port == "" {
		port = "8082"
	}
	return &Client{
		baseURL: "http://localhost:" + port,
	}
}

func (c *Client) Init(user, pass string) error {
	payload := map[string]string{"username": user, "password": pass}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(c.baseURL+"/init", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("init failed: %s", resp.Status)
	}
	return nil
}

func (c *Client) SubmitCode(code string) error {
	payload := map[string]string{"code": code}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(c.baseURL+"/submit-code", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("submit code failed: %s", string(body))
	}
	return nil
}

func (c *Client) Disconnect() error {
	resp, err := http.Post(c.baseURL+"/disconnect", "application/json", bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("disconnect failed: %s", resp.Status)
	}
	return nil
}

func (c *Client) GetStatus() ConnectionStatus {
	resp, err := http.Get(c.baseURL + "/status")
	if err != nil {
		return StatusDisconnected
	}
	defer resp.Body.Close()

	var res struct {
		Status       int    `json:"status"`
		ErrorMessage string `json:"errorMessage,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return StatusDisconnected
	}
	return ConnectionStatus(res.Status)
}

func (c *Client) GetStatusWithError() (ConnectionStatus, string) {
	resp, err := http.Get(c.baseURL + "/status")
	if err != nil {
		return StatusDisconnected, ""
	}
	defer resp.Body.Close()

	var res struct {
		Status       int    `json:"status"`
		ErrorMessage string `json:"errorMessage,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return StatusDisconnected, ""
	}
	return ConnectionStatus(res.Status), res.ErrorMessage
}

func (c *Client) GetReplayInfo(matchID uint64) (uint32, uint64, error) {
	resp, err := http.Get(fmt.Sprintf("%s/replay-info?match_id=%d", c.baseURL, matchID))
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var res struct {
		Cluster uint32 `json:"cluster"`
		Salt    uint64 `json:"salt"`
		Error   string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return 0, 0, err
	}

	if res.Error != "" {
		return 0, 0, fmt.Errorf(res.Error)
	}
	return res.Cluster, res.Salt, nil
}

type Match struct {
	ID        int64  `json:"id"`
	GameMode  uint32 `json:"gameMode"`
	LobbyType uint32 `json:"lobbyType"`
	StartTime uint32 `json:"startTime"`
}

func (c *Client) GetPlayerMatchHistory(steamID64 int64, limit int, turboOnly bool) ([]Match, error) {
	return c.GetPlayerMatchHistoryPaginated(steamID64, limit, turboOnly, 0)
}

func (c *Client) GetPlayerMatchHistoryPaginated(steamID64 int64, limit int, turboOnly bool, startAtMatchID uint64) ([]Match, error) {
	payload := map[string]interface{}{
		"steamId64": steamID64,
		"limit":     limit,
		"turboOnly": turboOnly,
	}
	if startAtMatchID > 0 {
		payload["startAtMatchId"] = startAtMatchID
	}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(c.baseURL+"/player-match-history", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf(errResp.Error)
		}
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var matches []Match
	if err := json.NewDecoder(resp.Body).Decode(&matches); err != nil {
		return nil, err
	}
	return matches, nil
}

type FatalMatchInfo struct {
	FatalMatchID       int64   `json:"fatalMatchId"`
	SingleDraftMatchID int64   `json:"singleDraftMatchId"`
	SingleDraftDate    uint32  `json:"singleDraftDate"`
	AdditionalMatchIDs []int64 `json:"additionalMatchIds"`
}

func (c *Client) FindFatalGames(steamID64 int64, maxDepth int, gamesPerFatal int) ([]FatalMatchInfo, error) {
	payload := map[string]interface{}{
		"steamId64":     steamID64,
		"maxDepth":      maxDepth,
		"gamesPerFatal": gamesPerFatal,
	}
	data, _ := json.Marshal(payload)
	resp, err := http.Post(c.baseURL+"/fatal-search", "application/json", bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf(errResp.Error)
		}
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var matches []FatalMatchInfo
	if err := json.NewDecoder(resp.Body).Decode(&matches); err != nil {
		return nil, err
	}
	return matches, nil
}

func (c *Client) GetPlayerConductScorecard() (map[string]interface{}, error) {
	resp, err := http.Get(c.baseURL + "/conduct-scorecard")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error string `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		if errResp.Error != "" {
			return nil, fmt.Errorf(errResp.Error)
		}
		return nil, fmt.Errorf("request failed with status %d", resp.StatusCode)
	}

	var scorecard map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&scorecard); err != nil {
		return nil, err
	}
	return scorecard, nil
}
