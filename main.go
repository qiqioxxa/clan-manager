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

	slices.SortFunc(players, func(a, b Player) int {
		var result int

		switch cfg.SortBy {
		case ByTag:
			result = cmp.Compare(a.Tag, b.Tag)
		case ByName:
			result = cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		case ByTownHall:
			result = cmp.Compare(a.TownHall, b.TownHall)
		case ByWarStars:
			result = cmp.Compare(a.WarStars, b.WarStars)
		case ByCapitalContributions:
			result = cmp.Compare(a.ClanCapitalContributions, b.ClanCapitalContributions)
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
		case ByHeroSum:
			fallthrough
		default:
			result = cmp.Compare(a.HeroSum, b.HeroSum)
		}

		if !cfg.Ascending {
			return -result
		}
		return result
	})

	fmt.Printf("Fetched %d players info\n", len(clanMembers))

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	t.AppendHeader(table.Row{"ID", "Name", "Townhall", "Hero sum", "CW+CWL stars", "CWL stars", "Donations", "Equip sum", "Lab updates"})

	for i, player := range players {
		t.AppendRow(table.Row{i + 1, player.Name, player.TownHall, player.HeroSum, player.WarStars, player.WarLeagueStars, player.Donations, player.EquipmentSum, player.LabUpdates})
	}

	t.Render()
}
