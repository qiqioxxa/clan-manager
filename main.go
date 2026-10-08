package main

import (
	"cmp"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"sync"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func main() {
	cfg, err := parseFlags()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	client := NewCoCClient(cfg.APIToken)

	switch cfg.Report {
	case ReportRoster:
		err = runRosterReport(client, cfg)
	case ReportCWLGroup:
		err = runCWLGroupReport(client, cfg)
	case ReportCWLLayout:
		err = runCWLLayoutReport(client, cfg)
	}

	if err != nil {
		log.Fatalf("Execution error: %v", err)
	}
}

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
			clan.Stars,
			clan.Attacks,
			fmt.Sprintf("%.2f", avgStars),
			clan.Score / teamSize,
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
		if round.WarTags[0] == "#0" {
			break
		}

		var enemyTag string
		clans := make([]WarClanLayout, 0, 8)

		for j, warTag := range round.WarTags {
			war, err := client.GetWar(warTag)
			if err != nil {
				return fmt.Errorf("failed to fetch war №%d of day %d", j, i)
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
			if ok && tagByName[clanName] == enemyTag {
				return text.Colors{text.BgRed, text.Bold}
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

func sortPlayers(players []Player, sortColumn SortColumn, report ReportType, reversed bool) {
	slices.SortFunc(players, func(a, b Player) int {
		var result int

		switch sortColumn {
		case ByName:
			result = cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		case ByTownHall:
			result = cmp.Compare(b.TownHall, a.TownHall)
		case ByWarStars:
			result = cmp.Compare(b.WarStars, a.WarStars)
		case ByCapitalContributions:
			result = cmp.Compare(b.ClanCapitalContributions, a.ClanCapitalContributions)
		case ByHeroSum:
			result = cmp.Compare(b.HeroSum, a.HeroSum)
		case ByEquipmentSum:
			result = cmp.Compare(b.EquipmentSum, a.EquipmentSum)
		case ByLabUpdates:
			result = cmp.Compare(b.LabUpdates, a.LabUpdates)
		case ByDonations:
			result = cmp.Compare(b.Donations, a.Donations)
		case ByClanGamesPoints:
			result = cmp.Compare(b.ClanGamesPoints, a.ClanGamesPoints)
		case ByWarLeagueStars:
			result = cmp.Compare(b.WarLeagueStars, a.WarLeagueStars)
		case ByScore:
			result = cmp.Compare(b.Score(report), a.Score(report))
		default:
			result = cmp.Compare(b.TownHall, a.TownHall)
		}

		if reversed {
			return -result
		}
		return result
	})
}
func sortClans(clans []ScoredCWLClan, sortColumn SortColumn, reversed bool) {
	slices.SortFunc(clans, func(a, b ScoredCWLClan) int {
		var result int

		switch sortColumn {
		case ByName:
			result = cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		case ByListed:
			result = cmp.Compare(len(b.ListedMembers), len(a.ListedMembers))
		case ByPresent:
			result = cmp.Compare(len(b.PresentMembers), len(a.PresentMembers))
		case ByScore:
			result = cmp.Compare(b.Score, a.Score)
		case ByAttacks:
			result = cmp.Compare(b.Attacks, a.Attacks)
		case ByStars:
			fallthrough
		default:
			if a.Stars != b.Stars {
				result = cmp.Compare(b.Stars, a.Stars)
			} else {
				result = cmp.Compare(b.Score, a.Score)
			}
		}

		if reversed {
			return -result
		}
		return result
	})
}
