package main

import "github.com/jedib0t/go-pretty/v6/table"

type RosterFormatter struct {
	Header      table.Row
	GenerateRow func(id int, p Player) table.Row
}

var rosterFormatters = map[RosterView]RosterFormatter{
	ViewAll: {
		Header: table.Row{
			"ID",
			"Name",
			"Townhall",
			"Hero sum",
			"Equip sum",
			"Lab updates",
			"CW stars",
			"CWL stars",
			"Donations",
			"Clan Games points",
			"Capital contributions",
		},
		GenerateRow: func(id int, p Player) table.Row {
			return table.Row{
				id,
				p.Name,
				p.TownHall,
				p.HeroSum,
				p.EquipmentSum,
				p.LabUpdates,
				p.WarStars,
				p.WarLeagueStars,
				p.Donations,
				p.ClanGamesPoints,
				p.ClanCapitalContributions,
			}
		},
	},
	ViewProgression: {
		Header: table.Row{
			"ID",
			"Name",
			"Townhall",
			"Hero sum",
			"Equip sum",
			"Lab updates",
			"Progress score",
		},
		GenerateRow: func(id int, p Player) table.Row {
			return table.Row{
				id,
				p.Name,
				p.TownHall,
				p.HeroSum,
				p.EquipmentSum,
				p.LabUpdates,
				p.Score(ReportRoster),
			}
		},
	},
	ViewActivity: {
		Header: table.Row{
			"ID",
			"Name",
			"Townhall",
			"CW stars",
			"CWL stars",
			"Donations",
			"Clan Games points",
			"Capital contributions",
		},
		GenerateRow: func(id int, p Player) table.Row {
			return table.Row{
				id,
				p.Name,
				p.TownHall,
				p.WarStars,
				p.WarLeagueStars,
				p.Donations,
				p.ClanGamesPoints,
				p.ClanCapitalContributions,
			}
		},
	},
}
