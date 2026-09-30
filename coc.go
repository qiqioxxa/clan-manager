package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type ClanMembersResponse struct {
	Items []ClanMember `json:"items"`
}

type ClanMember struct {
	Tag string `json:"tag"`
}

type Player struct {
	Tag      string  `json:"tag"`
	Name     string  `json:"name"`
	TownHall int     `json:"townHallLevel"`
	WarStars int     `json:"warStars"`
	Heroes   []Hero  `json:"heroes"`
	Troops   []Troop `json:"troops"`
	Spells   []Spell `json:"spells"`
}

type Hero struct {
	Name      string      `json:"name"`
	Level     int         `json:"level"`
	Village   string      `json:"village"`
	Equipment []Equipment `json:"equipment"`
}

type Equipment struct {
	Name    string `json:"name"`
	Level   int    `json:"level"`
	Village string `json:"village"`
}

type Troop struct {
	Name    string `json:"name"`
	Level   int    `json:"level"`
	Village string `json:"village"`
}

type Spell struct {
	Name    string `json:"name"`
	Level   int    `json:"level"`
	Village string `json:"village"`
}

type CoCClient struct {
	token      string
	httpClient *http.Client
}

func NewCoCClient(token string) *CoCClient {
	return &CoCClient{
		token: token,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (client *CoCClient) doRequest(endpoint string, target any) error {
	reqURL, err := url.JoinPath("https://api.clashofclans.com", endpoint)
	if err != nil {
		return fmt.Errorf("failed to construct request URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+client.token)

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("network request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed with status code: %s", resp.Status)
	}

	err = json.NewDecoder(resp.Body).Decode(target)
	if err != nil {
		return fmt.Errorf("failed to decode JSON response: %w", err)
	}

	return nil
}

func (client *CoCClient) GetClanMembers(clanTag string) ([]ClanMember, error) {
	var response ClanMembersResponse
	endpoint := fmt.Sprintf("v1/clans/%s/members", clanTag)

	err := client.doRequest(endpoint, &response)
	if err != nil {
		return nil, err
	}

	return response.Items, nil
}

func (client *CoCClient) GetPlayersInfo(clanMembers []ClanMember) ([]Player, error) {
	players := make([]Player, 0, len(clanMembers))

	for _, member := range clanMembers {
		var player Player
		endpoint := fmt.Sprintf("v1/players/%s", member.Tag)

		err := client.doRequest(endpoint, &player)
		if err != nil {
			return nil, err
		}

		players = append(players, player)
	}

	return players, nil
}
