package main

import (
	"cmp"
	"fmt"
	"os"
	"slices"
	"sync"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func runRosterReport(client *CoCClient, cfg Config) error {
	clanMembers, err := client.GetClanMembers(cfg.ClanTag)
	if err != nil {
		return fmt.Errorf("failed to fetch clan data: %v", err)
	}

	players, err := client.GetPlayersInfo(clanMembers)
	if err != nil {
		return fmt.Errorf("failed to fetch players info: %v", err)
	}

	fmt.Printf("Fetched %d players info\n", len(players))

	sortPlayers(players, cfg.SortBy, cfg.Report, cfg.Reversed)

	formatter := rosterFormatters[cfg.View]

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(formatter.Header)

	totalScore := 0
	for i, player := range players {
		t.AppendRow(formatter.GenerateRow(i+1, player))
		totalScore += player.Score(cfg.Report)
	}
	t.Render()

	if cfg.View == ViewProgression && len(players) != 0 {
		fmt.Printf("Avg clan %s score: %d\n", cfg.View, totalScore/len(players))
	}

	return nil
}
func runCWLGroupReport(client *CoCClient, cfg Config) error {
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
		warStats, ok := warStats[scoredClans[i].Tag]
		if ok {
			scoredClans[i].Stars = warStats.Stars
			scoredClans[i].Attacks = warStats.Attacks
			scoredClans[i].WinStars = warStats.WinStars
		}
	}

	sortClans(scoredClans, ByStars, false)
	for i := range scoredClans {
		scoredClans[i].Rank = i + 1
	}
	sortClans(scoredClans, cfg.SortBy, cfg.Reversed)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Rank", "Name", "Stars", "Attacks", "Stars/atk", "Score", "Listed", "Present", "Tag"})
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:  "Stars/atk",
			Align: text.AlignRight,
		},
	})

	for _, clan := range scoredClans {
		avgStars := 0.0
		if clan.Attacks > 0 {
			avgStars = float64(clan.Stars) / float64(clan.Attacks)
		}
		t.AppendRow(table.Row{
			clan.Rank,
			clan.Name,
			clan.Stars + clan.WinStars,
			clan.Attacks,
			fmt.Sprintf("%.2f", avgStars),
			clan.Score,
			len(clan.ListedMembers),
			len(clan.PresentMembers),
			clan.Tag,
		})
	}
	t.Render()

	return nil
}
func runCWLLayoutReport(client *CoCClient, cfg Config) error {
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

		fmt.Printf("=== DAY %d ===\n", i+1)

		t := table.NewWriter()
		t.SetOutputMirror(os.Stdout)
		t.AppendHeader(table.Row{"Clan", "<11", "11", "12", "13", "14", "15", "16", "17", "18", "Sum"})

		for _, clan := range clans {
			row := make(table.Row, 0, len(clan.THs)+2)
			row = append(row, clan.Name)
			for _, count := range clan.THs {
				row = append(row, count)
			}
			row = append(row, clan.THSum)

			t.AppendRow(row)
		}

		tagByName := make(map[string]string, len(clans))
		for _, clan := range clans {
			tagByName[clan.Name] = clan.Tag
		}

		t.SetRowPainter(func(row table.Row) text.Colors {
			clanName, ok := row[0].(string)
			if ok {
				if tagByName[clanName] == enemyTag {
					return text.Colors{text.BgRed, text.Bold}
				}
				if tagByName[clanName] == cfg.ClanTag {
					return text.Colors{text.BgGreen, text.Bold}
				}
			}
			return nil
		})
		t.Render()
	}

	return nil
}

type WarClanLayout struct {
	WarClan
	THs   []int
	THSum int
}
