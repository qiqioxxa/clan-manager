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

type ScoredClan struct {
	ClanCWL
	Rank           int
	Score          int
	PresentMembers []ClanMember
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
	lgr, err := client.GetLeagueGroup(cfg.ClanTag)
	if err != nil {
		return fmt.Errorf("failed to fetch league group: %w", err)
	}

	war, err := client.GetWarInfo(lgr.Rounds[0].WarTags[0])
	if err != nil {
		return fmt.Errorf("failed to fetch team size: %w", err)
	}

	scoredClans := make([]ScoredClan, len(lgr.Clans))

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for i, clan := range lgr.Clans {
		wg.Go(func() {
			scoredClan, err := client.GetScoredClan(clan, war.TeamSize)
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

	sortClans(scoredClans, ByScore, false)
	for i := range scoredClans {
		scoredClans[i].Rank = i + 1
	}
	sortClans(scoredClans, cfg.SortBy, cfg.Reversed)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Rank", "Name", "Tag", "Listed", "Present", fmt.Sprintf("Top-%d avg", war.TeamSize)})

	for _, clan := range scoredClans {
		t.AppendRow(table.Row{clan.Rank, clan.Name, clan.Tag, len(clan.ListedMembers), len(clan.PresentMembers), clan.Score / war.TeamSize})
	}
	t.Render()

	return nil
}
func runCWLLayoutReport(client *CoCClient, cfg Config) error {
	return nil
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
func sortClans(clans []ScoredClan, sortColumn SortColumn, reversed bool) {
	slices.SortFunc(clans, func(a, b ScoredClan) int {
		var result int

		switch sortColumn {
		case ByName:
			result = cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		case ByListed:
			result = cmp.Compare(len(b.ListedMembers), len(a.ListedMembers))
		case ByPresent:
			result = cmp.Compare(len(b.PresentMembers), len(a.PresentMembers))
		case ByScore:
			fallthrough
		default:
			result = cmp.Compare(b.Score, a.Score)
		}

		if reversed {
			return -result
		}
		return result
	})
}
