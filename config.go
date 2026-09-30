package main

import (
	"errors"
	"flag"
)

type Config struct {
	APIToken   string
	ClanTag    string
	OutputFile string
}

func parseFlags() (Config, error) {
	var cfg Config

	flag.StringVar(&cfg.APIToken, "token", "", "Supercell API bearer token")
	flag.StringVar(&cfg.ClanTag, "tag", "", "Clan tag")
	flag.StringVar(&cfg.OutputFile, "out", "clan_stats.xlsx", "Output Excel file")

	flag.Parse()

	if cfg.APIToken == "" || cfg.ClanTag == "" {
		return Config{}, errors.New("-token and -tag flags are required")
	}

	return cfg, nil
}
