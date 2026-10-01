package main

import (
	"errors"
	"flag"
	"fmt"
	"strings"
)

type SortColumn string

const (
	ByTag                  SortColumn = "tag"
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
	case ByTag, ByName, ByTownHall, ByWarStars, ByCapitalContributions, ByHeroSum, ByEquipmentSum, ByLabUpdates, ByDonations, ByClanGamesPoints, ByWarLeagueStars:
		*s = SortColumn(strings.ToLower(val))
		return nil
	default:
		return fmt.Errorf("invalid sort column %q (allowed: tag, name, th, cw, capital, herosum, equipsum, lab, donations, clangames, cwl)", val)
	}
}

type Config struct {
	APIToken   string
	ClanTag    string
	OutputFile string
	SortBy     SortColumn
	Ascending  bool
}

func parseFlags() (Config, error) {
	var cfg Config

	cfg.SortBy = ByHeroSum

	flag.StringVar(&cfg.APIToken, "token", "", "Supercell API bearer token")
	flag.StringVar(&cfg.ClanTag, "tag", "", "Clan tag")
	flag.StringVar(&cfg.OutputFile, "out", "clan_stats.xlsx", "Output Excel file")
	flag.Var(&cfg.SortBy, "sort", "Sort by column: tag, name, th, cw, capital, herosum, equipsum, lab, donations, clangames, cwl")
	flag.BoolVar(&cfg.Ascending, "asc", false, "Sorting in ascending order")

	flag.Parse()

	if cfg.APIToken == "" || cfg.ClanTag == "" {
		return Config{}, errors.New("-token and -tag flags are required")
	}

	return cfg, nil
}
