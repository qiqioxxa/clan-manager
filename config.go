package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Mode string

const (
	Progression   Mode = "progression"
	Activity      Mode = "activity"
	ClanWar       Mode = "cw"
	ClanWarLeague Mode = "cwl"
	All           Mode = "all"
)

func (p *Mode) String() string {
	return string(*p)
}
func (p *Mode) Set(val string) error {
	switch Mode(strings.ToLower(val)) {
	case Progression, Activity, ClanWar, ClanWarLeague, All:
		*p = Mode(strings.ToLower(val))
		return nil
	default:
		return fmt.Errorf("invalid table mode %q (allowed: progression, activity, cw, cwl, all)", val)
	}
}

type ModeFormatter struct {
	Header      table.Row
	GenerateRow func(id int, p Player) table.Row
}

var modes = map[Mode]ModeFormatter{
	Progression: {
		Header: table.Row{"ID", "Name", "Townhall", "Hero sum", "Equip sum", "Lab updates", "Progress score"},
		GenerateRow: func(id int, p Player) table.Row {
			return table.Row{id, p.Name, p.TownHall, p.HeroSum, p.EquipmentSum, p.LabUpdates, p.Score(Progression)}
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
			return table.Row{id, p.Name, p.TownHall, p.HeroSum, p.EquipmentSum, p.LabUpdates, p.WarStars, p.Score(ClanWar)}
		},
	},
	ClanWarLeague: {
		Header: table.Row{"ID", "Name", "Townhall", "Hero sum", "Equip sum", "Lab updates", "CWL stars", "CWL score"},
		GenerateRow: func(id int, p Player) table.Row {
			return table.Row{id, p.Name, p.TownHall, p.HeroSum, p.EquipmentSum, p.LabUpdates, p.WarLeagueStars, p.Score(ClanWarLeague)}
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
	ByScore                SortColumn = "score"
)

func (s *SortColumn) String() string {
	return string(*s)
}
func (s *SortColumn) Set(val string) error {
	switch SortColumn(strings.ToLower(val)) {
	case ByName, ByTownHall, ByWarStars, ByCapitalContributions, ByHeroSum, ByEquipmentSum, ByLabUpdates, ByDonations, ByClanGamesPoints, ByWarLeagueStars, ByScore:
		*s = SortColumn(strings.ToLower(val))
		return nil
	default:
		return fmt.Errorf("invalid sort column %q (allowed: name, th, herosum, equipsum, lab, cw, cwl, donations, clangames, capital, score)", val)
	}
}

type Config struct {
	APIToken   string
	ClanTag    string
	Mode       Mode
	SortBy     SortColumn
	Ascending  bool
	OutputFile string
}

func parseFlags() (Config, error) {
	var cfg Config

	cfg.Mode = All
	cfg.SortBy = ByScore

	flag.StringVar(&cfg.APIToken, "token", "", "Supercell API bearer token")
	flag.StringVar(&cfg.ClanTag, "tag", "", "Clan tag")
	flag.Var(&cfg.Mode, "mode", "Choose table mode: progression, activity, cw, cwl, all")
	flag.Var(&cfg.SortBy, "sort", "Sort by column: name, th, herosum, equipsum, lab, cw, cwl, donations, clangames, capital, score")
	flag.BoolVar(&cfg.Ascending, "asc", false, "Sorting in ascending order")
	flag.StringVar(&cfg.OutputFile, "out", "clan_stats.xlsx", "Output Excel file")

	flag.Parse()

	if cfg.APIToken == "" || cfg.ClanTag == "" {
		return Config{}, errors.New("-token and -tag flags are required")
	}

	if cfg.SortBy == ByScore && cfg.Mode != Progression && cfg.Mode != ClanWar && cfg.Mode != ClanWarLeague {
		cfg.SortBy = ByTownHall
	}

	return cfg, nil
}
