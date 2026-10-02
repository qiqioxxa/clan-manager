package main

import (
	"cmp"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

func main() {
	cfg, err := parseFlags()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	client := NewCoCClient(cfg.APIToken)
	clanMembers, err := client.GetClanMembers(cfg.ClanTag)
	if err != nil {
		log.Fatalf("Failed to fetch clan data: %v", err)
	}

	players, err := client.GetPlayersInfo(clanMembers)
	if err != nil {
		log.Fatalf("Failed to fetch players info: %v", err)
	}

	fmt.Printf("Fetched %d players info\n", len(players))

	sortPlayers(&cfg, players)

	formatter := modes[cfg.Mode]

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
}

func sortPlayers(cfg *Config, players []Player) {
	slices.SortFunc(players, func(a, b Player) int {
		var result int

		switch cfg.SortBy {
		case ByName:
			result = cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		case ByTownHall:
			result = cmp.Compare(a.TownHall, b.TownHall)
		case ByWarStars:
			result = cmp.Compare(a.WarStars, b.WarStars)
		case ByCapitalContributions:
			result = cmp.Compare(a.ClanCapitalContributions, b.ClanCapitalContributions)
		case ByHeroSum:
			result = cmp.Compare(a.HeroSum, b.HeroSum)
		case ByEquipmentSum:
			result = cmp.Compare(a.EquipmentSum, b.EquipmentSum)
		case ByLabUpdates:
			result = cmp.Compare(a.LabUpdates, b.LabUpdates)
		case ByDonations:
			result = cmp.Compare(a.Donations, b.Donations)
		case ByClanGamesPoints:
			result = cmp.Compare(a.ClanGamesPoints, b.ClanGamesPoints)
		case ByWarLeagueStars:
			result = cmp.Compare(a.WarLeagueStars, b.WarLeagueStars)
		case ByScore:
			result = cmp.Compare(a.Score(cfg.Mode), b.Score(cfg.Mode))
		default:
			result = cmp.Compare(a.TownHall, b.TownHall)
		}

		if !cfg.Ascending {
			return -result
		}
		return result
	})
}
