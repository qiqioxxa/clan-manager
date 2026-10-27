package main

const (
	maxTownHall  = 18
	maxHeroes    = 480
	maxEquipment = 324
	maxLab       = 570
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
	weightedSum := p.HeroSum*10 + p.EquipmentSum*12 + p.LabUpdates*3
	maxSum := maxHeroes*10 + maxEquipment*12 + maxLab*3
	return float64(weightedSum) / float64(maxSum) * 1000
}
func (p Player) clanWarLeagueScore() float64 {
	weightedSum := p.progressionScore() + float64(p.TownHall*100) + float64(min(1000, p.WarLeagueStars))
	maxSum := 1000 + (maxTownHall * 100) + 1000
	return weightedSum / float64(maxSum) * 1000
}
