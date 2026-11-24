package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"slices"
	"strings"
	"time"
)

const (
	developerPortalURL = "https://developer.clashofclans.com"
	generatedKeyName   = "auto-key"
	keyDateLayout      = "02.01.2006"
)

type CoCAPIKey struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	CIDRRanges  []string `json:"cidrRanges"`
	Token       string   `json:"key"`
}

type CoCDeveloperClient struct {
	email, password string
	httpClient      *http.Client
}

func NewCoCDeveloperClient(email, password string) *CoCDeveloperClient {
	jar, _ := cookiejar.New(nil)
	return &CoCDeveloperClient{
		email:      email,
		password:   password,
		httpClient: &http.Client{Jar: jar, Timeout: 10 * time.Second},
	}
}

func (client *CoCDeveloperClient) postJSON(path string, body, out any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("developer portal %s: encode request: %w", path, err)
	}
	req, err := http.NewRequest(http.MethodPost, developerPortalURL+path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("developer portal %s: create request: %w", path, err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("developer portal %s: %w", path, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("developer portal %s: read response: %w", path, err)
	}

	if resp.StatusCode != http.StatusOK {
		message := strings.TrimSpace(string(raw))
		return fmt.Errorf("developer portal %s: request failed with status code %s: %s", path, resp.Status, message)
	}

	var status struct {
		Status struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"status"`
	}
	err = json.Unmarshal(raw, &status)
	if err != nil {
		return fmt.Errorf("developer portal %s: decode response status: %w", path, err)
	}
	if status.Status.Code != 0 {
		return fmt.Errorf("developer portal %s: status %d: %s", path, status.Status.Code, status.Status.Message)
	}

	if out != nil {
		err = json.Unmarshal(raw, out)
		if err != nil {
			return fmt.Errorf("developer portal %s: decode response: %w", path, err)
		}
	}
	return nil
}

func (client *CoCDeveloperClient) login() (string, error) {
	var response struct {
		TemporaryAPIToken string `json:"temporaryAPIToken"`
	}
	err := client.postJSON("/api/login", map[string]string{"email": client.email, "password": client.password}, &response)
	if err != nil {
		return "", err
	}
	if response.TemporaryAPIToken == "" {
		return "", errors.New("login response has no temporaryAPIToken")
	}
	return response.TemporaryAPIToken, nil
}

func clientIPFromToken(temporaryToken string) (string, error) {
	parts := strings.Split(temporaryToken, ".")
	if len(parts) != 3 {
		return "", errors.New("unexpected temporaryAPIToken format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	var claims struct {
		Limits []struct {
			Type  string   `json:"type"`
			Cidrs []string `json:"cidrs"`
		} `json:"limits"`
	}
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return "", err
	}
	for _, limit := range claims.Limits {
		if limit.Type == "client" && len(limit.Cidrs) > 0 {
			return strings.TrimSuffix(limit.Cidrs[0], "/32"), nil
		}
	}
	return "", errors.New("no client CIDR in temporaryAPIToken")
}

func (client *CoCDeveloperClient) listAPIKeys() ([]CoCAPIKey, error) {
	var resp struct {
		Keys []CoCAPIKey `json:"keys"`
	}
	err := client.postJSON("/api/apikey/list", struct{}{}, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Keys, nil
}

func (client *CoCDeveloperClient) createAPIKey(clientIP string) (CoCAPIKey, error) {
	var resp struct {
		Key CoCAPIKey `json:"key"`
	}
	err := client.postJSON("/api/apikey/create", map[string]any{
		"name":        generatedKeyName,
		"description": time.Now().Format(keyDateLayout),
		"cidrRanges":  []string{clientIP},
		"scopes":      nil,
	}, &resp)
	return resp.Key, err
}

func (client *CoCDeveloperClient) revokeAPIKey(id string) error {
	return client.postJSON("/api/apikey/revoke", map[string]string{"id": id}, nil)
}

func (client *CoCDeveloperClient) APITokenForCurrentIP() (string, error) {
	temporaryToken, err := client.login()
	if err != nil {
		return "", err
	}
	clientIP, err := clientIPFromToken(temporaryToken)
	if err != nil {
		return "", fmt.Errorf("developer portal login token: %w", err)
	}
	keys, err := client.listAPIKeys()
	if err != nil {
		return "", err
	}
	for _, key := range keys {
		if slices.Contains(key.CIDRRanges, clientIP) {
			return key.Token, nil
		}
	}

	if len(keys) >= 10 {
		var oldest *CoCAPIKey
		var oldestDate time.Time
		for i := range keys {
			if keys[i].Name != generatedKeyName {
				continue
			}

			date, _ := time.Parse(keyDateLayout, keys[i].Description)

			if oldest == nil || date.Before(oldestDate) {
				oldest = &keys[i]
				oldestDate = date
			}
		}
		if oldest == nil {
			return "", fmt.Errorf("key limit reached and no %q keys to revoke", generatedKeyName)
		}
		err = client.revokeAPIKey(oldest.ID)
		if err != nil {
			return "", err
		}
	}

	key, err := client.createAPIKey(clientIP)
	if err != nil {
		return "", err
	}
	return key.Token, nil
}
