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

	if cfg.Mode == CWLGroup {
		err = runCWLGroupReport(client, cfg)
	} else {
		err = runClanReport(client, cfg)
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
			currentMembers, err := client.GetClanMembers(clan.Tag)
			if err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("failed to fetch current members for clan %q: %w", clan.Name, err)
				}
				mu.Unlock()
				return
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
				mu.Lock()
				if firstErr == nil {
					firstErr = fmt.Errorf("failed to fetch players info for %q: %w", clan.Name, err)
				}
				mu.Unlock()
				return
			}

			sortPlayers(players, ByScore, CWLGroup, false)

			limit := min(war.TeamSize, len(players))

			clanScore := 0
			for j := range limit {
				clanScore += players[j].Score(cfg.Mode)
			}

			scoredClans[i] = ScoredClan{
				ClanCWL:        clan,
				Score:          clanScore,
				PresentMembers: presentMembers,
			}
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
	t.AppendHeader(table.Row{"Rank", "Name", "Tag", "Listed", "Present", fmt.Sprintf("Top-%d score", war.TeamSize)})

	for _, clan := range scoredClans {
		t.AppendRow(table.Row{clan.Rank, clan.Name, clan.Tag, len(clan.ListedMembers), len(clan.PresentMembers), clan.Score})
	}

	t.Render()

	return nil
}
func runClanReport(client *CoCClient, cfg Config) error {
	clanMembers, err := client.GetClanMembers(cfg.ClanTag)
	if err != nil {
		return fmt.Errorf("failed to fetch clan data: %v", err)
	}

	players, err := client.GetPlayersInfo(clanMembers)
	if err != nil {
		return fmt.Errorf("failed to fetch players info: %v", err)
	}

	fmt.Printf("Fetched %d players info\n", len(players))

	sortPlayers(players, cfg.SortBy, cfg.Mode, cfg.Reversed)

	formatter := rosterFormatters[cfg.Mode]

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(formatter.Header)

	totalScore := 0
	for i, player := range players {
		t.AppendRow(formatter.GenerateRow(i+1, player))
		totalScore += player.Score(cfg.Mode)
	}

	t.Render()

	if cfg.Mode == Progression || cfg.Mode == ClanWar || cfg.Mode == ClanWarLeague {
		fmt.Printf("Total clan %s score: %d, avg = %d\n", cfg.Mode, totalScore, totalScore/len(players))
	}

	return nil
}

func sortPlayers(players []Player, sortColumn SortColumn, mode Mode, reversed bool) {
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
			result = cmp.Compare(b.Score(mode), a.Score(mode))
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
