package main

const (
	maxTownHall  = 18
	maxHeroes    = 480
	maxEquipment = 324
	maxLab       = 570
)

func (p *Player) calculateSums() {
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
func (p *Player) processAchievements() {
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

func (p Player) Score(forCWL bool) int {
	if forCWL {
		return p.clanWarLeagueScore()
	} else {
		return p.progressionScore()
	}
}
func (p Player) progressionScore() int {
	weightedSum := p.HeroSum*10 + p.EquipmentSum*12 + p.LabUpdates*3
	maxSum := maxHeroes*10 + maxEquipment*12 + maxLab*3
	return int(float64(weightedSum) / float64(maxSum) * 1000)
}
func (p Player) clanWarLeagueScore() int {
	weightedSum := p.progressionScore() + p.TownHall*100 + min(1000, p.WarLeagueStars)
	maxSum := 1000 + (maxTownHall * 100) + 1000
	return int(float64(weightedSum) / float64(maxSum) * 1000)
}
