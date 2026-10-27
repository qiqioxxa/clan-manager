package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

type PlayerSnapshot struct {
	Name                     string `json:"name"`
	TownHall                 int    `json:"townHallLevel"`
	HeroSum                  int    `json:"heroSum"`
	EquipmentSum             int    `json:"equipSum"`
	LabUpdates               int    `json:"lab"`
	WarStars                 int    `json:"warStars"`
	WarLeagueStars           int    `json:"warLeagueStars"`
	Donations                int    `json:"donations"`
	ClanGamesPoints          int    `json:"clanGamesPoints"`
	ClanCapitalContributions int    `json:"clanCapitalContributions"`
}

type ClanSnapshot struct {
	Timestamp string                    `json:"timestamp"`
	ClanTag   string                    `json:"clanTag"`
	Players   map[string]PlayerSnapshot `json:"players"`
}

func SaveClanSnapshot(clanTag string, players []Player) error {
	cleanTag := strings.TrimPrefix(clanTag, "#")
	dir := filepath.Join("snapshots", cleanTag)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create snapshots directory: %w", err)
	}

	nowUTC := time.Now().UTC()
	filename := nowUTC.Format("2006-01-02_15-04-05") + ".json"
	filePath := filepath.Join(dir, filename)

	snapshot := ClanSnapshot{
		Timestamp: nowUTC.Format(time.RFC3339),
		ClanTag:   clanTag,
		Players:   make(map[string]PlayerSnapshot),
	}

	for _, p := range players {
		snapshot.Players[p.Tag] = PlayerSnapshot{
			Name:                     p.Name,
			TownHall:                 p.TownHall,
			HeroSum:                  p.HeroSum,
			EquipmentSum:             p.EquipmentSum,
			LabUpdates:               p.LabUpdates,
			WarStars:                 p.WarStars,
			WarLeagueStars:           p.WarLeagueStars,
			Donations:                p.Donations,
			ClanGamesPoints:          p.ClanGamesPoints,
			ClanCapitalContributions: p.ClanCapitalContributions,
		}
	}

	fileData, err := json.MarshalIndent(snapshot, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal snapshot: %w", err)
	}

	err = os.WriteFile(filePath, fileData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write snapshot file: %w", err)
	}

	return nil
}

func ResolveSnapshot(clanTag, flagValue string) (ClanSnapshot, error) {
	cleanTag := strings.TrimPrefix(clanTag, "#")
	dir := filepath.Join("snapshots", cleanTag)

	files, err := os.ReadDir(dir)
	if err != nil {
		return ClanSnapshot{}, fmt.Errorf("failed to read snapshots directory: %w", err)
	}

	files = slices.DeleteFunc(files, func(e os.DirEntry) bool {
		return strings.HasPrefix(e.Name(), ".")
	})

	depth, err := strconv.Atoi(flagValue)
	if err != nil {
		return ClanSnapshot{}, fmt.Errorf("invalid depth value: %w", err)
	}

	if depth <= 0 || depth > len(files) {
		return ClanSnapshot{}, fmt.Errorf("entered depth out of range [1,%d]", len(files))
	}

	slices.SortFunc(files, func(a, b os.DirEntry) int {
		return cmp.Compare(b.Name(), a.Name())
	})

	path := filepath.Join(dir, files[depth-1].Name())
	var snapshot ClanSnapshot

	file, err := os.Open(path)
	if err != nil {
		return ClanSnapshot{}, fmt.Errorf("failed to open file %q: %w", files[depth-1].Name(), err)
	}
	defer file.Close()

	json.NewDecoder(file).Decode(&snapshot)

	return snapshot, nil
}

func CalculateDiff(players []Player, clanSnapshot ClanSnapshot) []Player {
	diffedPlayers := make([]Player, 0, len(players))

	for _, p := range players {
		saved, ok := clanSnapshot.Players[p.Tag]
		if !ok {
			continue
		}
		diffedPlayer := Player{
			Tag:                      p.Tag,
			Name:                     p.Name,
			TownHall:                 p.TownHall - saved.TownHall,
			HeroSum:                  p.HeroSum - saved.HeroSum,
			EquipmentSum:             p.EquipmentSum - saved.EquipmentSum,
			LabUpdates:               p.LabUpdates - saved.LabUpdates,
			WarStars:                 p.WarStars - saved.WarStars,
			WarLeagueStars:           p.WarLeagueStars - saved.WarLeagueStars,
			Donations:                p.Donations - saved.Donations,
			ClanGamesPoints:          p.ClanGamesPoints - saved.ClanGamesPoints,
			ClanCapitalContributions: p.ClanCapitalContributions - saved.ClanCapitalContributions,
		}

		diffedPlayers = append(diffedPlayers, diffedPlayer)
	}

	return diffedPlayers
}
