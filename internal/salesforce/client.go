package salesforce

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Client struct {
	clientID     string
	clientSecret string
	loginURL     string
	httpClient   *http.Client

	mu          sync.RWMutex
	accessToken string
	instanceURL string
	expiresAt   time.Time
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	InstanceURL string `json:"instance_url"`
	TokenType   string `json:"token_type"`
	IssuedAt    string `json:"issued_at"`
}

func NewClient(clientID, clientSecret, loginURL string) *Client {
	return &Client{
		clientID:     clientID,
		clientSecret: clientSecret,
		loginURL:     loginURL,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) authenticate() error {
	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", c.clientID)
	data.Set("client_secret", c.clientSecret)

	req, err := http.NewRequest("POST", c.loginURL+"/services/oauth2/token", strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("creating auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("auth failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return fmt.Errorf("parsing token response: %w", err)
	}

	c.mu.Lock()
	c.accessToken = tokenResp.AccessToken
	c.instanceURL = tokenResp.InstanceURL
	c.expiresAt = time.Now().Add(1 * time.Hour) // Conservative expiry
	c.mu.Unlock()

	return nil
}

func (c *Client) ensureAuth() error {
	c.mu.RLock()
	valid := c.accessToken != "" && time.Now().Before(c.expiresAt)
	c.mu.RUnlock()

	if valid {
		return nil
	}
	return c.authenticate()
}

func (c *Client) InstanceURL() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.instanceURL
}

type QueryResult struct {
	TotalSize      int              `json:"totalSize"`
	Done           bool             `json:"done"`
	Records        []map[string]any `json:"records"`
	NextRecordsURL string           `json:"nextRecordsUrl"`
}

func (c *Client) Query(soql string) (*QueryResult, error) {
	if err := c.ensureAuth(); err != nil {
		return nil, err
	}

	c.mu.RLock()
	instanceURL := c.instanceURL
	token := c.accessToken
	c.mu.RUnlock()

	queryURL := fmt.Sprintf("%s/services/data/v62.0/query?q=%s", instanceURL, url.QueryEscape(soql))

	req, err := http.NewRequest("GET", queryURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating query request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("query request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusUnauthorized {
		// Token expired, re-auth and retry once
		if err := c.authenticate(); err != nil {
			return nil, err
		}
		return c.Query(soql)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result QueryResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parsing query result: %w", err)
	}

	return &result, nil
}
