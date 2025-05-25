package ui

import (
	"fmt"
	"gotopia-rpg/api"
	"gotopia-rpg/model"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Game     *model.Game
	Quitting bool
}

func NewModel(g *model.Game) Model {
	return Model{Game: g}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		case "n":
			if m.Game.Scene == model.SceneSpawn {
				mon, err := api.FetchRandomMonster()
				if err != nil {
					fmt.Println("Error fetching monster:", err)
					// fallback to default monster if API fails
					m.Game.CurrentMon = model.Monster{
						Name:            "Goblin",
						Description:     "Just some forest goblin",
						HP:              10,
						AC:              12,
						ChallengeRating: "1/4",
					}
				} else {
					m.Game.CurrentMon = model.Monster{
						Name:            mon.Name,
						Description:     string(mon.Description),
						HP:              mon.HitPoints,
						AC:              mon.ArmorClass,
						ChallengeRating: mon.ChallengeRating,
					}
				}
				m.Game.Scene = model.SceneBattle
			}
		case "1":
			if m.Game.Scene == model.SceneBattle && m.Game.CurrentMon.HP > 0 {
				// Simple attack: subtract 3 HP per attack
				m.Game.CurrentMon.HP -= 3
				if m.Game.CurrentMon.HP <= 0 {
					// Monster defeated, return to spawn scene
					m.Game.Scene = model.SceneSpawn
				}
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.Quitting {
		return "Goodbye\n"
	}
	s := "-- Gotopia RPG -- \n"
	switch m.Game.Scene {
	case model.SceneSpawn:
		s += "Press [n] to start battle\n"
	case model.SceneBattle:
		s += fmt.Sprintf("Enemy: %s\n", m.Game.CurrentMon.Name)
		if m.Game.CurrentMon.Description != "" {
			s += fmt.Sprintf("Description: %s\n", m.Game.CurrentMon.Description)
		} else {
			s += "Description: No description available.\n"
		}
		s += fmt.Sprintf("HP:%d AC:%d\n", m.Game.CurrentMon.HP, m.Game.CurrentMon.AC)
		s += "[1] Attack  [q] Quit"
	}
	return s
}
