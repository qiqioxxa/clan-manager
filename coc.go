package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type CoCClient struct {
	token      string
	httpClient *http.Client
	sem        chan struct{}
}

func NewCoCClient(token string) *CoCClient {
	return &CoCClient{
		token: token,
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
		sem: make(chan struct{}, 10),
	}
}

func (client *CoCClient) GetLeagueGroup(clanTag string) (LeagueGroupResponse, error) {
	var response LeagueGroupResponse
	endpoint := fmt.Sprintf("v1/clans/%s/currentwar/leaguegroup", clanTag)

	err := client.doRequest(endpoint, &response)
	if err != nil {
		return LeagueGroupResponse{}, err
	}

	return response, nil
}
func (client *CoCClient) GetWarInfo(warTag string) (War, error) {
	var war War
	endpoint := fmt.Sprintf("v1/clanwarleagues/wars/%s", warTag)

	err := client.doRequest(endpoint, &war)
	if err != nil {
		return War{}, err
	}

	return war, nil
}
func (client *CoCClient) GetScoredClan(clan ClanCWL, teamSize int, mode Mode) (ScoredClan, error) {
	currentMembers, err := client.GetClanMembers(clan.Tag)
	if err != nil {
		return ScoredClan{}, fmt.Errorf("failed to fetch current members for clan %q: %w", clan.Name, err)
	}

	inClan := make(map[string]bool, len(currentMembers))
	for _, m := range currentMembers {
		inClan[m.Tag] = true
	}

	presentMembers := make([]ClanMember, 0, len(clan.ListedMembers))
	for _, regMem := range clan.ListedMembers {
		if inClan[regMem.Tag] {
			presentMembers = append(presentMembers, regMem)
		}
	}

	players, err := client.GetPlayersInfo(presentMembers)
	if err != nil {
		return ScoredClan{}, fmt.Errorf("failed to fetch players info for %q: %w", clan.Name, err)
	}

	sortPlayers(players, ByScore, CWLGroup, false)

	limit := min(teamSize, len(players))
	clanScore := 0
	for j := range limit {
		clanScore += players[j].Score(mode)
	}

	return ScoredClan{
		ClanCWL:        clan,
		Score:          clanScore,
		PresentMembers: presentMembers,
	}, nil
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

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i, member := range clanMembers {
		wg.Go(func() {
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
func (client *CoCClient) doRequest(endpoint string, target any) error {
	client.sem <- struct{}{}
	defer func() { <-client.sem }()

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

type LeagueGroupResponse struct {
	Clans  []ClanCWL `json:"clans"`
	Rounds []Round   `json:"rounds"`
}
type ClanCWL struct {
	Tag           string       `json:"tag"`
	Name          string       `json:"name"`
	ListedMembers []ClanMember `json:"members"`
}
type Round struct {
	WarTags [4]string `json:"warTags"`
}
type War struct {
	TeamSize int `json:"teamSize"`
}
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

func (p *Player) Score(mode Mode) int {
	switch mode {
	case Progression:
		return p.progressionScore()
	case ClanWar:
		return p.clanWarScore()
	case ClanWarLeague, CWLGroup:
		return p.clanWarLeagueScore()
	default:
		return 0
	}
}
func (p *Player) progressionScore() int {
	weightedSum := p.HeroSum*10 + p.EquipmentSum*12 + p.LabUpdates*3
	return weightedSum / 8
}
func (p *Player) clanWarScore() int {
	maxLimit, okMax := heroSumLimits[p.TownHall]
	minLimit, okMin := heroSumLimits[p.TownHall-3]

	if !okMax {
		maxLimit = p.TownHall * 25
	}
	if !okMin {
		minLimit = 0
	}

	threeTownhallRange := maxLimit - minLimit
	if threeTownhallRange == 0 {
		threeTownhallRange = 1
	}

	rushCoef := float64(p.HeroSum-minLimit) / float64(threeTownhallRange)

	if rushCoef < 0.0 {
		rushCoef = 0.0
	}

	weightedSum := p.HeroSum*10 + p.EquipmentSum*12 + p.LabUpdates*3

	score := (float64(weightedSum) + float64(p.WarStars*5)) * rushCoef / float64(p.TownHall)

	return int(score)
}
func (p *Player) clanWarLeagueScore() int {
	weightedSum := p.HeroSum*10 + p.EquipmentSum*12 + p.LabUpdates*3
	return (weightedSum/8 + p.TownHall*100 + p.WarLeagueStars) / 5
}

var heroSumLimits = map[int]int{
	7:  10,
	8:  30,
	9:  70,
	10: 100,
	11: 150,
	12: 210,
	13: 275,
	14: 320,
	15: 365,
	16: 400,
	17: 435,
	18: 465,
}
