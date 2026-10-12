package main

import (
	"cmp"
	"slices"
	"strings"
)

func sortPlayers(players []Player, sortColumn SortColumn, report ReportType, reversed bool) {
	slices.SortFunc(players, func(a, b Player) int {
		var result int

		switch sortColumn {
		case ByName:
			result = cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		case ByTownHall:
			if a.TownHall != b.TownHall {
				result = cmp.Compare(b.TownHall, a.TownHall)
			} else {
				result = cmp.Compare(b.Score(ReportRoster), a.Score(ReportRoster))
			}
		case ByWarStars:
			result = cmp.Compare(b.WarStars, a.WarStars)
		case ByCapitalContributions:
			result = cmp.Compare(b.ClanCapitalContributions, a.ClanCapitalContributions)
		case ByHeroSum:
			result = cmp.Compare(b.HeroSum, a.HeroSum)
		case ByEquipmentSum:
			result = cmp.Compare(b.EquipmentSum, a.EquipmentSum)
		case ByLabUpdates:
			result = cmp.Compare(b.LabUpdates, a.LabUpdates)
		case ByDonations:
			result = cmp.Compare(b.Donations, a.Donations)
		case ByClanGamesPoints:
			result = cmp.Compare(b.ClanGamesPoints, a.ClanGamesPoints)
		case ByWarLeagueStars:
			result = cmp.Compare(b.WarLeagueStars, a.WarLeagueStars)
		case ByScore:
			result = cmp.Compare(b.Score(report), a.Score(report))
		default:
			result = cmp.Compare(b.TownHall, a.TownHall)
		}

		if reversed {
			return -result
		}
		return result
	})
}
func sortClans(clans []ScoredCWLClan, sortColumn SortColumn, reversed bool) {
	slices.SortFunc(clans, func(a, b ScoredCWLClan) int {
		var result int

		switch sortColumn {
		case ByName:
			result = cmp.Compare(strings.ToLower(a.Name), strings.ToLower(b.Name))
		case ByListed:
			result = cmp.Compare(len(b.ListedMembers), len(a.ListedMembers))
		case ByPresent:
			result = cmp.Compare(len(b.PresentMembers), len(a.PresentMembers))
		case ByScore:
			result = cmp.Compare(b.Score, a.Score)
		case ByAttacks:
			result = cmp.Compare(b.Attacks, a.Attacks)
		case ByStars:
			fallthrough
		default:
			if a.Stars != b.Stars {
				result = cmp.Compare(b.Stars, a.Stars)
			} else {
				result = cmp.Compare(b.Score, a.Score)
			}
		}

		if reversed {
			return -result
		}
		return result
	})
}
