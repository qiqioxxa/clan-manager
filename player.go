package main

const (
	maxTownHall  = 19
	maxHeroes    = 510
	maxEquipment = 396
	maxLab       = 638
)

func (p *Player) CalculateSums() {
	p.HeroSum = 0
	p.EquipmentSum = 0
	p.LabUpdates = 0

	for _, hero := range p.Heroes {
		if hero.Village == "home" {
			p.HeroSum += hero.Level
			for _, equipment := range hero.Equipment {
				p.EquipmentSum += equipment.Level
			}
		}
	}

	for _, troop := range p.Troops {
		if troop.Village == "home" {
			p.LabUpdates += troop.Level - 1
		}
	}

	for _, spell := range p.Spells {
		if spell.Village == "home" {
			p.LabUpdates += spell.Level - 1
		}
	}
}
func (p *Player) ProcessAchievements() {
	for _, a := range p.Achievements {
		switch a.Name {
		case "Friend in Need":
			p.Donations = a.Value
		case "Games Champion":
			p.ClanGamesPoints = a.Value
		case "War League Legend":
			p.WarLeagueStars = a.Value
		}
	}
	p.Achievements = nil
}

func (p Player) Score(forCWL bool) float64 {
	if forCWL {
		return p.clanWarLeagueScore()
	} else {
		return p.progressionScore()
	}
}
func (p Player) progressionScore() float64 {
	return (0.50*float64(p.HeroSum)/maxHeroes + 0.35*float64(p.EquipmentSum)/maxEquipment + 0.15*float64(p.LabUpdates)/maxLab) * 1000
}
func (p Player) clanWarLeagueScore() float64 {
	return (0.20*p.progressionScore()/1000 + 0.50*float64(p.TownHall)/maxTownHall + 0.30*float64(min(1000, p.WarLeagueStars))/1000) * 1000
}
