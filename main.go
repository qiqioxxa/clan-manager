package main

import (
	"fmt"
	"log"
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

	for _, player := range players {
		fmt.Println(player.Tag, player.Name, player.TownHall)
	}
}
