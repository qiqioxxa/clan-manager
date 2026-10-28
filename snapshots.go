package main

import (
	"cmp"
	"encoding/json"
	"fmt"
	"math"
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

	if len(files) == 0 {
		return ClanSnapshot{}, fmt.Errorf("no snapshots found for clan %s", clanTag)
	}

	var filePath string
	diffType, value := checkDiffType(flagValue)

	switch diffType {
	case TypeDepth:
		depth := value.(int)
		if depth <= 0 || depth > len(files) {
			return ClanSnapshot{}, fmt.Errorf("entered depth out of range [1,%d]", len(files))
		}

		slices.SortFunc(files, func(a, b os.DirEntry) int {
			return cmp.Compare(b.Name(), a.Name())
		})

		filePath = filepath.Join(dir, files[depth-1].Name())

	case TypeDate:
		inputDate := value.(time.Time)

		var bestFile os.DirEntry
		minOffset := math.MaxFloat64

		for _, file := range files {
			timestamp, err := time.Parse("2006-01-02_15-04-05", strings.TrimSuffix(file.Name(), ".json"))
			if err != nil {
				continue
			}

			offset := math.Abs(float64(inputDate.Unix() - timestamp.Unix()))
			if offset < minOffset {
				minOffset = offset
				bestFile = file
			}
		}

		if bestFile == nil {
			return ClanSnapshot{}, fmt.Errorf("no valid snapshot found")
		}

		filePath = filepath.Join(dir, bestFile.Name())

		if minOffset != 0 {
			timestamp, _ := time.Parse("2006-01-02_15-04-05", strings.TrimSuffix(bestFile.Name(), ".json"))
			fmt.Printf("Exact snapshot not found, used closest (%.2fd off): %v\n", minOffset/3600/24, timestamp.Local().Format(time.DateTime))
		}

	case TypeInvalid:
		fallthrough
	default:
		return ClanSnapshot{}, fmt.Errorf("invalid -diff value (expected depth number or YYYY-MM-DD)")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return ClanSnapshot{}, fmt.Errorf("failed to open file %q: %w", filePath, err)
	}
	defer file.Close()

	var snapshot ClanSnapshot
	err = json.NewDecoder(file).Decode(&snapshot)
	if err != nil {
		return ClanSnapshot{}, fmt.Errorf("failed to decode snapshor JSON: %w", err)
	}

	return snapshot, nil
}
func checkDiffType(flagValue string) (diffType, any) {
	timestamp, err := time.ParseInLocation("2006-01-02T15:04:05", flagValue, time.Local)
	if err == nil {
		return TypeDate, timestamp
	}
	timestamp, err = time.ParseInLocation(time.DateOnly, flagValue, time.Local)
	if err == nil {
		return TypeDate, timestamp
	}

	depth, err := strconv.Atoi(flagValue)
	if err == nil {
		return TypeDepth, depth
	}

	return TypeInvalid, nil
}

type diffType string

const (
	TypeDate    diffType = "DATE"
	TypeDepth   diffType = "DEPTH"
	TypeInvalid diffType = "INVALID"
)

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
