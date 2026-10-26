package main

import (
	"fmt"
	"log"
)

func main() {
	parsedCfg, err := ParseFlags()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	switch cfg := parsedCfg.(type) {
	case RosterConfig:
		client := NewCoCClient(cfg.APIToken)
		err = RunRosterReport(client, cfg)
	case CWLGroupConfig:
		client := NewCoCClient(cfg.APIToken)
		err = RunCWLGroupReport(client, cfg)
	case CWLLayoutConfig:
		client := NewCoCClient(cfg.APIToken)
		err = RunCWLLayoutReport(client, cfg)
	default:
		err = fmt.Errorf("Unknown configuration type")
	}

	if err != nil {
		log.Fatalf("Execution error: %v", err)
	}
}
