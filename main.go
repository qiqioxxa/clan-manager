package main

import (
	"errors"
	"log"
)

func main() {
	parsedCfg, err := ParseFlags()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	switch cfg := parsedCfg.(type) {
	case RosterConfig:
		client, clientErr := newClient(cfg.APIToken)
		if clientErr != nil {
			err = clientErr
		} else {
			err = RunRosterReport(client, cfg)
		}
	case CWLGroupConfig:
		client, clientErr := newClient(cfg.APIToken)
		if clientErr != nil {
			err = clientErr
		} else {
			err = RunCWLGroupReport(client, cfg)
		}
	case CWLLayoutConfig:
		client, clientErr := newClient(cfg.APIToken)
		if clientErr != nil {
			err = clientErr
		} else {
			err = RunCWLLayoutReport(client, cfg)
		}
	default:
		err = errors.New("unknown configuration type")
	}

	if err != nil {
		log.Fatalf("Execution error: %v", err)
	}
}
