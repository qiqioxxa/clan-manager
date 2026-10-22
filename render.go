package main

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

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
				p.Score(false),
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

func renderRoster(players []Player, view RosterView) {
	formatter := rosterFormatters[view]

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(formatter.Header)

	totalScore := 0
	for i, player := range players {
		t.AppendRow(formatter.GenerateRow(i+1, player))
		totalScore += player.Score(false)
	}

	t.Render()

	if view == ViewProgression && len(players) != 0 {
		fmt.Printf("Avg clan progression score: %d\n", totalScore/len(players))
	}
}
func renderCWLGroup(clans []ScoredCWLClan) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{
		"Rank",
		"Name",
		"Stars",
		"Attacks",
		"Stars/atk",
		"Score", "Listed",
		"Present", "Tag",
	})
	t.SetColumnConfigs([]table.ColumnConfig{
		{
			Name:  "Stars/atk",
			Align: text.AlignRight,
		},
	})

	for _, clan := range clans {
		avgStars := 0.0
		if clan.Attacks > 0 {
			avgStars = float64(clan.Stars) / float64(clan.Attacks)
		}
		t.AppendRow(table.Row{
			clan.Rank,
			clan.Name,
			clan.Stars + clan.WinStars,
			clan.Attacks,
			fmt.Sprintf("%.2f", avgStars),
			clan.Score,
			len(clan.ListedMembers),
			len(clan.PresentMembers),
			clan.Tag,
		})
	}

	t.Render()
}
func renderCWLLayout(clans []WarClanLayout, day int, ownTag, enemyTag string) {
	fmt.Printf("=== DAY %d ===\n", day)

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{
		"Clan",
		"<11",
		"11",
		"12",
		"13",
		"14",
		"15",
		"16",
		"17",
		"18",
		"Sum",
	})

	for _, clan := range clans {
		row := make(table.Row, 0, len(clan.THs)+2)
		row = append(row, clan.Name)
		for _, count := range clan.THs {
			row = append(row, count)
		}
		row = append(row, clan.THSum)

		t.AppendRow(row)
	}

	tagByName := make(map[string]string, len(clans))
	for _, clan := range clans {
		tagByName[clan.Name] = clan.Tag
	}

	t.SetRowPainter(func(row table.Row) text.Colors {
		clanName, ok := row[0].(string)
		if !ok {
			return nil
		}

		switch tagByName[clanName] {
		case ownTag:
			return text.Colors{text.BgGreen, text.Bold}
		case enemyTag:
			return text.Colors{text.BgRed, text.Bold}
		default:
			return nil
		}
	})

	t.Render()
}
