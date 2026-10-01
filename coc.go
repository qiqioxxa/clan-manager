package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type ClanMembersResponse struct {
	Items []ClanMember `json:"items"`
}

type ClanMember struct {
	Tag string `json:"tag"`
}

type Player struct {
	Tag                      string        `json:"tag"`
	Name                     string        `json:"name"`
	TownHall                 int           `json:"townHallLevel"`
	WarStars                 int           `json:"warStars"`
	ClanCapitalContributions int           `json:"clanCapitalContributions"`
	Achievements             []Achievement `json:"achievements"`
	Heroes                   []Hero        `json:"heroes"`
	Troops                   []Troop       `json:"troops"`
	Spells                   []Spell       `json:"spells"`

	HeroSum         int `json:"-"`
	EquipmentSum    int `json:"-"`
	LabUpdates      int `json:"-"`
	Donations       int `json:"-"`
	ClanGamesPoints int `json:"-"`
	WarLeagueStars  int `json:"-"`
}

type Achievement struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
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
			Timeout: 20 * time.Second,
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
	players := make([]Player, len(clanMembers))

	const maxConcurrency = 10
	sem := make(chan struct{}, maxConcurrency)

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i, member := range clanMembers {
		wg.Go(func() {
			sem <- struct{}{}
			defer func() { <-sem }()

			endpoint := fmt.Sprintf("v1/players/%s", member.Tag)
			var player Player

			err := client.doRequest(endpoint, &player)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("failed to fetch player %s: %w", member.Tag, err)
				}
				mu.Unlock()
				return
			}

			player.calculateSums()
			player.processAchievements()

			players[i] = player
		})
	}

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	return players, nil
}

func (p *Player) calculateSums() {
	p.HeroSum = 0
	p.EquipmentSum = 0
	p.LabUpdates = 0

	for _, hero := range p.Heroes {
		if hero.Village == "home" {
			p.HeroSum += hero.Level
			for _, equipment := range hero.Equipment {
				p.EquipmentSum += equipment.Level
			}
		}
	}

	for _, troop := range p.Troops {
		if troop.Village == "home" {
			p.LabUpdates += troop.Level - 1
		}
	}

	for _, spell := range p.Spells {
		if spell.Village == "home" {
			p.LabUpdates += spell.Level - 1
		}
	}
}

func (p *Player) processAchievements() {
	for _, a := range p.Achievements {
		switch a.Name {
		case "Friend in Need":
			p.Donations = a.Value
		case "Games Champion":
			p.ClanGamesPoints = a.Value
		case "War League Legend":
			p.WarLeagueStars = a.Value
		}
	}
	p.Achievements = nil
}

func (p *Player) ProgressionScore() int {
	return 0
}

func (p *Player) ClanWarScore() int {
	return 0
}

func (p *Player) ClanWarLeagueScore() int {
	return 0
}
