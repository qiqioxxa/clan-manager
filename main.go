package main

import (
	"log"
)

func main() {
	cfg, err := parseFlags()
	if err != nil {
		log.Fatalf("Configuration error: %v", err)
	}

	client := NewCoCClient(cfg.APIToken)

	switch cfg.Report {
	case ReportRoster:
		err = runRosterReport(client, cfg)
	case ReportCWLGroup:
		err = runCWLGroupReport(client, cfg)
	case ReportCWLLayout:
		err = runCWLLayoutReport(client, cfg)
	}

	if err != nil {
		log.Fatalf("Execution error: %v", err)
	}
}
