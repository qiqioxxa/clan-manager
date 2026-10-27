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
			Timeout: 10 * time.Second,
		},
		sem: make(chan struct{}, 10),
	}
}

func (client *CoCClient) GetCWLGroup(clanTag string) (CWLGroup, error) {
	var group CWLGroup
	endpoint := fmt.Sprintf("v1/clans/%s/currentwar/leaguegroup", clanTag)

	err := client.doRequest(endpoint, &group)
	if err != nil {
		return CWLGroup{}, err
	}

	return group, nil
}
func (client *CoCClient) GetWar(warTag string) (War, error) {
	var war War
	endpoint := fmt.Sprintf("v1/clanwarleagues/wars/%s", warTag)

	err := client.doRequest(endpoint, &war)
	if err != nil {
		return War{}, err
	}

	return war, nil
}
func (client *CoCClient) GetScoredClan(clan CWLClan, teamSize int) (ScoredCWLClan, error) {
	currentMembers, err := client.GetClanMembers(clan.Tag)
	if err != nil {
		return ScoredCWLClan{}, fmt.Errorf("failed to fetch current members of clan %q: %w", clan.Name, err)
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
		return ScoredCWLClan{}, fmt.Errorf("failed to fetch players info for %q: %w", clan.Name, err)
	}

	SortPlayers(players, ByScore, true, false)

	limit := min(teamSize, len(players))
	if limit <= 0 {
		return ScoredCWLClan{
			CWLClan:        clan,
			Score:          0,
			PresentMembers: presentMembers,
		}, nil
	}

	clanScore := 0.0
	for j := range limit {
		clanScore += players[j].Score(true)
	}

	return ScoredCWLClan{
		CWLClan:        clan,
		Score:          clanScore / float64(limit),
		PresentMembers: presentMembers,
	}, nil
}
func (client *CoCClient) GetCWLWarStats(rounds []Round) (map[string]WarStats, error) {
	var warTags []string
	for _, round := range rounds {
		for _, warTag := range round.WarTags {
			if warTag != "#0" {
				warTags = append(warTags, warTag)
			}
		}
	}

	wars := make([]War, len(warTags))

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i, warTag := range warTags {
		wg.Go(func() {
			war, err := client.GetWar(warTag)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			wars[i] = war
		})
	}

	wg.Wait()

	if firstErr != nil {
		return map[string]WarStats{}, firstErr
	}

	warStats := make(map[string]WarStats)

	for _, war := range wars {
		clanStats := warStats[war.Clan.Tag]
		clanStats.TeamSize = war.TeamSize
		clanStats.Attacks += war.Clan.Attacks
		clanStats.Stars += war.Clan.Stars
		if war.Clan.Stars > war.Opponent.Stars {
			clanStats.WinStars += 10
		}
		warStats[war.Clan.Tag] = clanStats

		oppStats := warStats[war.Opponent.Tag]
		oppStats.TeamSize = war.TeamSize
		oppStats.Attacks += war.Opponent.Attacks
		oppStats.Stars += war.Opponent.Stars
		if war.Opponent.Stars > war.Clan.Stars {
			oppStats.WinStars += 10
		}
		warStats[war.Opponent.Tag] = oppStats
	}

	return warStats, nil
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

			player.CalculateSums()
			player.ProcessAchievements()

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

	const maxRetries = 5
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		req, err := http.NewRequest(http.MethodGet, reqURL, nil)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+client.token)

		resp, err := client.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("network request failed: %w", err)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			lastErr = fmt.Errorf("API request failed with status code: %s", resp.Status)
			if resp.StatusCode >= http.StatusInternalServerError {
				time.Sleep(time.Duration(attempt) * time.Second)
				continue
			}
			return lastErr
		}

		err = json.NewDecoder(resp.Body).Decode(target)
		resp.Body.Close()

		if err != nil {
			lastErr = fmt.Errorf("failed to decode JSON response: %w", err)
			time.Sleep(time.Duration(attempt) * time.Second)
			continue
		}

		return nil
	}

	return fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}
