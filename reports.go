package main

import (
	"cmp"
	"fmt"
	"slices"
	"sync"
)

func runRosterReport(client *CoCClient, cfg RosterConfig) error {
	clanMembers, err := client.GetClanMembers(cfg.ClanTag)
	if err != nil {
		return fmt.Errorf("failed to fetch clan data: %w", err)
	}

	players, err := client.GetPlayersInfo(clanMembers)
	if err != nil {
		return fmt.Errorf("failed to fetch players info: %w", err)
	}

	sortPlayers(players, cfg.SortBy, false, cfg.Reversed)

	renderRoster(players, cfg.View)
	return nil
}
func runCWLGroupReport(client *CoCClient, cfg CWLGroupConfig) error {
	group, err := client.GetCWLGroup(cfg.ClanTag)
	if err != nil {
		return fmt.Errorf("failed to fetch CWL group: %w", err)
	}

	warStats, err := client.GetCWLWarStats(group.Rounds)
	if err != nil {
		return fmt.Errorf("failed to fetch war stats: %w", err)
	}
	teamSize := warStats[group.Clans[0].Tag].TeamSize

	scoredClans := make([]ScoredCWLClan, len(group.Clans))

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i, clan := range group.Clans {
		wg.Go(func() {
			scoredClan, err := client.GetScoredClan(clan, teamSize)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				mu.Unlock()
				return
			}
			scoredClans[i] = scoredClan
		})
	}

	wg.Wait()

	if firstErr != nil {
		return firstErr
	}

	for i := range scoredClans {
		stats, ok := warStats[scoredClans[i].Tag]
		if ok {
			scoredClans[i].Stars = stats.Stars
			scoredClans[i].Attacks = stats.Attacks
			scoredClans[i].WinStars = stats.WinStars
		}
	}

	sortClans(scoredClans, ByStars, false)
	for i := range scoredClans {
		scoredClans[i].Rank = i + 1
	}
	sortClans(scoredClans, cfg.SortBy, cfg.Reversed)

	renderCWLGroup(scoredClans)
	return nil
}
func runCWLLayoutReport(client *CoCClient, cfg CWLLayoutConfig) error {
	group, err := client.GetCWLGroup(cfg.ClanTag)
	if err != nil {
		return fmt.Errorf("failed to fetch CWL group: %w", err)
	}

	for i, round := range group.Rounds {
		var enemyTag string
		clans := make([]WarClanLayout, 0, 8)

		for j, warTag := range round.WarTags {
			if warTag == "#0" {
				continue
			}

			war, err := client.GetWar(warTag)
			if err != nil {
				return fmt.Errorf("failed to fetch war №%d of day %d: %w", j+1, i+1, err)
			}

			if war.Clan.Tag == cfg.ClanTag {
				enemyTag = war.Opponent.Tag
			} else if war.Opponent.Tag == cfg.ClanTag {
				enemyTag = war.Clan.Tag
			}

			clanTHs := make([]int, 9)
			oppTHs := make([]int, 9)
			clanTHSum := 0
			oppTHSum := 0

			for _, member := range war.Clan.Members {
				th := min(18, max(10, member.TownHall))
				clanTHs[th-10]++
				clanTHSum += th
			}
			for _, member := range war.Opponent.Members {
				th := min(18, max(10, member.TownHall))
				oppTHs[th-10]++
				oppTHSum += th
			}

			clans = append(clans, WarClanLayout{
				WarClan: war.Clan,
				THs:     clanTHs,
				THSum:   clanTHSum,
			})
			clans = append(clans, WarClanLayout{
				WarClan: war.Opponent,
				THs:     oppTHs,
				THSum:   oppTHSum,
			})
		}

		if len(clans) == 0 {
			continue
		}

		slices.SortFunc(clans, func(a, b WarClanLayout) int {
			if a.Tag == cfg.ClanTag {
				return -1
			}
			if b.Tag == cfg.ClanTag {
				return 1
			}
			return cmp.Compare(a.Name, b.Name)
		})

		renderCWLLayout(clans, i+1, cfg.ClanTag, enemyTag)
	}

	return nil
}

type WarClanLayout struct {
	WarClan
	THs   []int
	THSum int
}
