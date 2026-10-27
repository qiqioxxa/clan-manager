package main

type CWLGroup struct {
	Clans  []CWLClan `json:"clans"`
	Rounds []Round   `json:"rounds"`
}
type Clan struct {
	Tag  string `json:"tag"`
	Name string `json:"name"`
}
type CWLClan struct {
	Clan
	ListedMembers []ClanMember `json:"members"`
}
type Round struct {
	WarTags [4]string `json:"warTags"`
}
type War struct {
	TeamSize int     `json:"teamSize"`
	Clan     WarClan `json:"clan"`
	Opponent WarClan `json:"opponent"`
}
type WarClan struct {
	Clan
	Attacks int         `json:"attacks"`
	Stars   int         `json:"stars"`
	Members []WarMember `json:"members"`
}
type WarMember struct {
	Tag      string `json:"tag"`
	TownHall int    `json:"townHallLevel"`
}
type ScoredCWLClan struct {
	CWLClan
	Rank           int
	Score          float64
	Stars          int
	Attacks        int
	WinStars       int
	PresentMembers []ClanMember
}
type WarStats struct {
	TeamSize int
	Stars    int
	Attacks  int
	WinStars int
}
type ClanMembersResponse struct {
	Items []ClanMember `json:"items"`
}
type ClanMember struct {
	Tag string `json:"tag"`
}

type Player struct {
	Tag                      string        `json:"tag"`
	Name                     string        `json:"name"`
	TownHall                 int           `json:"townHallLevel"`
	WarStars                 int           `json:"warStars"`
	ClanCapitalContributions int           `json:"clanCapitalContributions"`
	Achievements             []Achievement `json:"achievements"`
	Heroes                   []Hero        `json:"heroes"`
	Troops                   []Troop       `json:"troops"`
	Spells                   []Spell       `json:"spells"`

	HeroSum         int `json:"-"`
	EquipmentSum    int `json:"-"`
	LabUpdates      int `json:"-"`
	Donations       int `json:"-"`
	ClanGamesPoints int `json:"-"`
	WarLeagueStars  int `json:"-"`
}
type Achievement struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}
type Hero struct {
	Name      string      `json:"name"`
	Level     int         `json:"level"`
	Village   string      `json:"village"`
	Equipment []Equipment `json:"equipment"`
}
type Equipment struct {
	Name    string `json:"name"`
	Level   int    `json:"level"`
	Village string `json:"village"`
}
type Troop struct {
	Name    string `json:"name"`
	Level   int    `json:"level"`
	Village string `json:"village"`
}
type Spell struct {
	Name    string `json:"name"`
	Level   int    `json:"level"`
	Village string `json:"village"`
}
