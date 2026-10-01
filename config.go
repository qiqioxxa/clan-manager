package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Preset string

const (
	Progression   Preset = "progression"
	Activity      Preset = "activity"
	ClanWar       Preset = "cw"
	ClanWarLeague Preset = "cwl"
	All           Preset = "all"
)

func (p *Preset) String() string {
	return string(*p)
}
func (p *Preset) Set(val string) error {
	switch Preset(strings.ToLower(val)) {
	case Progression, Activity, ClanWar, ClanWarLeague, All:
		*p = Preset(strings.ToLower(val))
		return nil
	default:
		return fmt.Errorf("invalid table preset %q (allowed: progression, activity, all)", val)
	}
}

type PresetFormatter struct {
	Header      table.Row
	GenerateRow func(id int, p Player) table.Row
}

var presets = map[Preset]PresetFormatter{
	Progression: {
		Header: table.Row{"ID", "Name", "Townhall", "Hero sum", "Equip sum", "Lab updates", "Progress score"},
		GenerateRow: func(id int, p Player) table.Row {
			return table.Row{id, p.Name, p.TownHall, p.HeroSum, p.EquipmentSum, p.LabUpdates, p.ProgressionScore()}
		},
	},
	Activity: {
		Header: table.Row{"ID", "Name", "Townhall", "CW+CWL stars", "CWL stars", "Donations", "Clan Games points", "Capital contibutions"},
		GenerateRow: func(id int, p Player) table.Row {
			return table.Row{id, p.Name, p.TownHall, p.WarStars, p.WarLeagueStars, p.Donations, p.ClanGamesPoints, p.ClanCapitalContributions}
		},
	},
	ClanWar: {
		Header: table.Row{"ID", "Name", "Townhall", "Hero sum", "Equip sum", "Lab updates", "CW+CWL stars", "CW score"},
		GenerateRow: func(id int, p Player) table.Row {
			return table.Row{id, p.Name, p.TownHall, p.HeroSum, p.EquipmentSum, p.LabUpdates, p.WarStars, p.ClanWarScore()}
		},
	},
	ClanWarLeague: {
		Header: table.Row{"ID", "Name", "Townhall", "Hero sum", "Equip sum", "Lab updates", "CWL stars", "CWL score"},
		GenerateRow: func(id int, p Player) table.Row {
			return table.Row{id, p.Name, p.TownHall, p.HeroSum, p.EquipmentSum, p.LabUpdates, p.WarLeagueStars, p.ClanWarLeagueScore()}
		},
	},
	All: {
		Header: table.Row{"ID", "Name", "Townhall", "Hero sum", "Equip sum", "Lab updates", "CW+CWL stars", "CWL stars", "Donations", "Clan Games points", "Capital contributions"},
		GenerateRow: func(id int, p Player) table.Row {
			return table.Row{id, p.Name, p.TownHall, p.HeroSum, p.EquipmentSum, p.LabUpdates, p.WarStars, p.WarLeagueStars, p.Donations, p.ClanGamesPoints, p.ClanCapitalContributions}
		},
	},
}

type SortColumn string

const (
	ByName                 SortColumn = "name"
	ByTownHall             SortColumn = "th"
	ByWarStars             SortColumn = "cw"
	ByCapitalContributions SortColumn = "capital"
	ByHeroSum              SortColumn = "herosum"
	ByEquipmentSum         SortColumn = "equipsum"
	ByLabUpdates           SortColumn = "lab"
	ByDonations            SortColumn = "donations"
	ByClanGamesPoints      SortColumn = "clangames"
	ByWarLeagueStars       SortColumn = "cwl"
)

func (s *SortColumn) String() string {
	return string(*s)
}
func (s *SortColumn) Set(val string) error {
	switch SortColumn(strings.ToLower(val)) {
	case ByName, ByTownHall, ByWarStars, ByCapitalContributions, ByHeroSum, ByEquipmentSum, ByLabUpdates, ByDonations, ByClanGamesPoints, ByWarLeagueStars:
		*s = SortColumn(strings.ToLower(val))
		return nil
	default:
		return fmt.Errorf("invalid sort column %q (allowed: name, th, herosum, equipsum, lab, cw, cwl, donations, clangames, capital)", val)
	}
}

type Config struct {
	APIToken   string
	ClanTag    string
	Preset     Preset
	SortBy     SortColumn
	Ascending  bool
	OutputFile string
}

func parseFlags() (Config, error) {
	var cfg Config

	cfg.Preset = All
	cfg.SortBy = ByTownHall

	flag.StringVar(&cfg.APIToken, "token", "", "Supercell API bearer token")
	flag.StringVar(&cfg.ClanTag, "tag", "", "Clan tag")
	flag.Var(&cfg.Preset, "preset", "Choose table preset: progression, activity, all")
	flag.Var(&cfg.SortBy, "sort", "Sort by column: name, th, herosum, equipsum, lab, cw, cwl, donations, clangames, capital")
	flag.BoolVar(&cfg.Ascending, "asc", false, "Sorting in ascending order")
	flag.StringVar(&cfg.OutputFile, "out", "clan_stats.xlsx", "Output Excel file")

	flag.Parse()

	if cfg.APIToken == "" || cfg.ClanTag == "" {
		return Config{}, errors.New("-token and -tag flags are required")
	}

	return cfg, nil
}
