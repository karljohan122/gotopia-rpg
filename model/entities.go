package model

//Stats holds basic ability scores
// Game holds state, reset logic, scene progression

type Stats struct {
	Str, Dex, Con, Int, Wis, Cha int
}

type Player struct {
	Race       string
	Stats      Stats
	ArmorClass int
	HP, MP     int
	Inventory  []Item
	Equipped   map[string]Item
}

type Monster struct {
	Name            string
	Description     string
	HP, AC          int
	ChallengeRating string
	// Attacks, spells can be added
}

type Item struct {
	Name string
}

// Scenes
type Scene int

const (
	SceneSpawn Scene = iota
	SceneBattle
)

type Game struct {
	Player     Player
	CurrentMon Monster
	Scene      Scene
}

func NewGame() *Game {
	return &Game{
		Player: Player{
			Stats:     Stats{},
			Inventory: []Item{},
			Equipped:  make(map[string]Item),
		},
		CurrentMon: Monster{},
		Scene:      SceneSpawn,
	}
}
