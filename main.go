package main

import (
	"fmt"
	"log"
	"os"

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

	fmt.Printf("Fetched %d players info\n", len(clanMembers))

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)

	t.AppendHeader(table.Row{"Name", "Townhall", "Hero sum", "CW+CWL stars", "CWL stars", "Donations"})

	for _, player := range players {
		t.AppendRow(table.Row{player.Name, player.TownHall, player.HeroSum, player.WarStars, player.TotalWarLeagueStars, player.TotalDonations})
	}

	t.Render()
}
