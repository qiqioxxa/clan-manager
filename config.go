package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Config struct {
	APIToken   string
	ClanTag    string
	Report     ReportType
	View       RosterView
	SortBy     SortColumn
	Reversed   bool
	OutputFile string
}

func parseFlags() (Config, error) {
	var cfg Config

	cfg.Report = ReportRoster
	cfg.View = ViewAll
	cfg.SortBy = ByScore

	flag.StringVar(&cfg.APIToken, "token", "", "Supercell API bearer token")
	flag.StringVar(&cfg.ClanTag, "tag", "", "Clan tag")
	flag.Var(&cfg.Report, "report", "Report type: roster, cwlgroup, cwllayout")
	flag.Var(&cfg.View, "view", "Roster view (for -report roster): all, progression, activity")
	flag.Var(&cfg.SortBy, "sort", "Sort column")
	flag.BoolVar(&cfg.Reversed, "r", false, "Reverse sorting order")
	flag.StringVar(&cfg.OutputFile, "out", "clan_stats.xlsx", "Output Excel file")

	flag.Parse()

	err := cfg.Validate()
	if err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.APIToken == "" || c.ClanTag == "" {
		return errors.New("-token and -tag flags are required")
	}

	switch c.Report {
	case ReportRoster:
		hasScore := c.View == ViewProgression
		if c.SortBy == ByScore && !hasScore {
			sortExplicitlySet := false
			flag.Visit(func(f *flag.Flag) {
				if f.Name == "sort" {
					sortExplicitlySet = true
				}
			})
			if !sortExplicitlySet {
				c.SortBy = ByTownHall
			}
		}

		allowedSort := map[SortColumn]bool{
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
			ByScore:                hasScore,
		}

		if !allowedSort[c.SortBy] {
			allowed := "name, th, herosum, equipsum, lab, cw, cwl, donations, clangames, capital"
			if hasScore {
				allowed += ", score"
			}
			return fmt.Errorf("invalid -sort %q for -view %q (allowed: %s)", c.SortBy, c.View, allowed)
		}

	case ReportCWLGroup:
		viewExplicitlySet := false
		flag.Visit(func(f *flag.Flag) {
			if f.Name == "view" {
				viewExplicitlySet = true
			}
		})
		if viewExplicitlySet {
			return fmt.Errorf("flag -view cannot be set for cwlgroup")
		}

		allowedSort := map[SortColumn]bool{
			ByName:    true,
			ByListed:  true,
			ByPresent: true,
			ByScore:   true,
		}
		if !allowedSort[c.SortBy] {
			return fmt.Errorf("invalid -sort %q for -report %q (allowed: name, listed, present, score)", c.SortBy, ReportCWLGroup)
		}

	case ReportCWLLayout:
	}

	return nil
}

type ReportType string

const (
	ReportRoster    ReportType = "roster"
	ReportCWLGroup  ReportType = "cwlgroup"
	ReportCWLLayout ReportType = "cwllayout"
)

func (rt *ReportType) String() string {
	return string(*rt)
}
func (rt *ReportType) Set(val string) error {
	switch ReportType(strings.ToLower(val)) {
	case ReportRoster, ReportCWLGroup, ReportCWLLayout:
		*rt = ReportType(strings.ToLower(val))
		return nil
	default:
		return fmt.Errorf("invalid report type %q (allowed: roster, cwlgroup, cwllayout)", val)
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
)

func (s *SortColumn) String() string {
	return string(*s)
}
func (s *SortColumn) Set(val string) error {
	switch SortColumn(strings.ToLower(val)) {
	case ByName, ByTownHall, ByHeroSum, ByEquipmentSum, ByLabUpdates, ByWarStars, ByWarLeagueStars, ByDonations, ByClanGamesPoints, ByCapitalContributions, ByScore, ByListed, ByPresent:
		*s = SortColumn(strings.ToLower(val))
		return nil
	default:
		return fmt.Errorf("invalid sort column %q (allowed: name, th, herosum, equipsum, lab, cw, cwl, donations, clangames, capital, score, listed, present)", val)
	}
}

type RosterFormatter struct {
	Header      table.Row
	GenerateRow func(id int, p Player) table.Row
}

var rosterFormatters = map[RosterView]RosterFormatter{
	ViewAll: {
		Header: table.Row{"ID", "Name", "Townhall", "Hero sum", "Equip sum", "Lab updates", "CW stars", "CWL stars", "Donations", "Clan Games points", "Capital contributions"},
		GenerateRow: func(id int, p Player) table.Row {
			return table.Row{id, p.Name, p.TownHall, p.HeroSum, p.EquipmentSum, p.LabUpdates, p.WarStars, p.WarLeagueStars, p.Donations, p.ClanGamesPoints, p.ClanCapitalContributions}
		},
	},
	ViewProgression: {
		Header: table.Row{"ID", "Name", "Townhall", "Hero sum", "Equip sum", "Lab updates", "Progress score"},
		GenerateRow: func(id int, p Player) table.Row {
			return table.Row{id, p.Name, p.TownHall, p.HeroSum, p.EquipmentSum, p.LabUpdates, p.Score(ReportRoster)}
		},
	},
	ViewActivity: {
		Header: table.Row{"ID", "Name", "Townhall", "CW+CWL stars", "CWL stars", "Donations", "Clan Games points", "Capital contributions"},
		GenerateRow: func(id int, p Player) table.Row {
			return table.Row{id, p.Name, p.TownHall, p.WarStars, p.WarLeagueStars, p.Donations, p.ClanGamesPoints, p.ClanCapitalContributions}
		},
	},
}
