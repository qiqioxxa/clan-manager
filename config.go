package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

type GlobalConfig struct {
	APIToken string
	ClanTag  string
}

type RosterConfig struct {
	GlobalConfig
	View     RosterView
	SortBy   SortColumn
	Reversed bool
	Snapshot bool
	Diff     string
}

type CWLGroupConfig struct {
	GlobalConfig
	SortBy   SortColumn
	Reversed bool
}

type CWLLayoutConfig struct {
	GlobalConfig
}

func ParseFlags() (any, error) {
	if len(os.Args) < 2 {
		return nil, fmt.Errorf("expected subcommand: roster, cwlgroup, cwllayout")
	}

	subcommand := os.Args[1]
	subArgs := os.Args[2:]

	switch subcommand {
	case "roster":
		var cfg RosterConfig
		set := flag.NewFlagSet("roster", flag.ExitOnError)

		set.StringVar(&cfg.APIToken, "token", "", "Supercell API bearer token")
		set.StringVar(&cfg.ClanTag, "tag", "", "Clan tag")
		set.Var(&cfg.View, "view", "Roster view: all, progression, activity")
		set.Var(&cfg.SortBy, "sort", "Sort column")
		set.BoolVar(&cfg.Reversed, "r", false, "Reverse sorting order")
		set.BoolVar(&cfg.Snapshot, "snapshot", false, "Make a snapshot of current roster")
		set.StringVar(&cfg.Diff, "diff", "", "Compare current roster with snapshot")

		if err := set.Parse(subArgs); err != nil {
			return nil, err
		}
		return cfg, nil

	case "cwlgroup":
		var cfg CWLGroupConfig
		set := flag.NewFlagSet("cwlgroup", flag.ExitOnError)

		set.StringVar(&cfg.APIToken, "token", "", "Supercell API bearer token")
		set.StringVar(&cfg.ClanTag, "tag", "", "Clan tag")
		set.Var(&cfg.SortBy, "sort", "Sort column")
		set.BoolVar(&cfg.Reversed, "r", false, "Reverse sorting order")

		if err := set.Parse(subArgs); err != nil {
			return nil, err
		}
		return cfg, nil

	case "cwllayout":
		var cfg CWLLayoutConfig
		set := flag.NewFlagSet("cwllayout", flag.ExitOnError)

		set.StringVar(&cfg.APIToken, "token", "", "Supercell API bearer token")
		set.StringVar(&cfg.ClanTag, "tag", "", "Clan tag")

		if err := set.Parse(subArgs); err != nil {
			return nil, err
		}
		return cfg, nil

	default:
		return nil, fmt.Errorf("unknown subcommand: %s", subcommand)
	}
}

type RosterView string

const (
	ViewAll         RosterView = "all"
	ViewProgression RosterView = "progression"
	ViewActivity    RosterView = "activity"
)

func (rv *RosterView) String() string {
	return string(*rv)
}
func (rv *RosterView) Set(val string) error {
	switch RosterView(strings.ToLower(val)) {
	case ViewAll, ViewProgression, ViewActivity:
		*rv = RosterView(strings.ToLower(val))
		return nil
	default:
		return fmt.Errorf("invalid roster view %q (allowed: all, progression, activity)", val)
	}
}

type SortColumn string

const (
	ByName                 SortColumn = "name"
	ByTownHall             SortColumn = "th"
	ByHeroSum              SortColumn = "herosum"
	ByEquipmentSum         SortColumn = "equipsum"
	ByLabUpdates           SortColumn = "lab"
	ByWarStars             SortColumn = "cw"
	ByWarLeagueStars       SortColumn = "cwl"
	ByDonations            SortColumn = "donations"
	ByClanGamesPoints      SortColumn = "clangames"
	ByCapitalContributions SortColumn = "capital"
	ByScore                SortColumn = "score"
	ByListed               SortColumn = "listed"
	ByPresent              SortColumn = "present"
	ByStars                SortColumn = "stars"
	ByAttacks              SortColumn = "attacks"
)

var validSortColumns = map[SortColumn]bool{
	ByName:                 true,
	ByTownHall:             true,
	ByHeroSum:              true,
	ByEquipmentSum:         true,
	ByLabUpdates:           true,
	ByWarStars:             true,
	ByWarLeagueStars:       true,
	ByDonations:            true,
	ByClanGamesPoints:      true,
	ByCapitalContributions: true,
	ByScore:                true,
	ByListed:               true,
	ByPresent:              true,
	ByStars:                true,
	ByAttacks:              true,
}

func (s *SortColumn) String() string {
	return string(*s)
}
func (s *SortColumn) Set(val string) error {
	v := SortColumn(strings.ToLower(val))
	if !validSortColumns[v] {
		return fmt.Errorf("invalid sort column %q (allowed: name, th, herosum, equipsum, lab, cw, cwl, donations, clangames, capital, score, listed, present, stars, attacks)", val)
	}
	*s = v
	return nil
}
