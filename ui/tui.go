package ui

import (
	"fmt"
	"gotopia-rpg/api"
	"gotopia-rpg/model"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	Game     *model.Game
	Quitting bool
	Loading  bool
	Spinner  spinner.Model
	Err      error
}

type monsterFetchedMsg struct {
	Monster model.Monster
	Err     error
}

func NewModel(g *model.Game) Model {
	s := spinner.New()
	s.Spinner = spinner.Dot

	return Model{
		Game:    g,
		Spinner: s,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit
		case "n":
			if m.Game.Scene == model.SceneSpawn && !m.Loading {
				m.Loading = true
				m.Err = nil
				cmds = append(cmds, fetchMonsterCmd(), m.Spinner.Tick)
				return m, tea.Batch(cmds...)
			}
		case "1":
			if m.Game.Scene == model.SceneBattle && m.Game.CurrentMon.HP > 0 {
				m.Game.CurrentMon.HP -= 3
				if m.Game.CurrentMon.HP <= 0 {
					m.Game.Scene = model.SceneSpawn
				}
			}
		}
	case monsterFetchedMsg:
		m.Loading = false
		if msg.Err != nil {
			m.Err = msg.Err
			m.Game.CurrentMon = model.Monster{
				Name:            "Goblin",
				Description:     "Just some forest goblin",
				HP:              10,
				AC:              12,
				ChallengeRating: "1/4",
			}
		} else {
			m.Game.CurrentMon = msg.Monster
			m.Game.Scene = model.SceneBattle
			m.Err = nil
		}
	case spinner.TickMsg:
		if m.Loading {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	if len(cmds) > 0 {
		return m, tea.Batch(cmds...)
	}
	return m, nil
}

func (m Model) View() string {
	if m.Quitting {
		return "Goodbye\n"
	}
	s := "-- Gotopia RPG -- \n"
	if m.Loading {
		s += fmt.Sprintf("%s Loading...\n", m.Spinner.View())
		return s
	}
	if m.Err != nil {
		s += fmt.Sprintf("Error: %v\n", m.Err)
	}
	switch m.Game.Scene {
	case model.SceneSpawn:
		s += "Press [n] to start battle\n"
	case model.SceneBattle:
		s += fmt.Sprintf("Enemy: %s\n", m.Game.CurrentMon.Name)
		if m.Game.CurrentMon.Description != "" && m.Game.CurrentMon.Description != "False" {
			s += fmt.Sprintf("Description: %s\n", m.Game.CurrentMon.Description)
		} else {
			s += "Description: No description available.\n"
		}
		s += fmt.Sprintf("HP:%d AC:%d\n", m.Game.CurrentMon.HP, m.Game.CurrentMon.AC)
		s += "[1] Attack  [q] Quit"
	}
	return s
}

func fetchMonsterCmd() tea.Cmd {
	return func() tea.Msg {
		mon, err := api.FetchRandomMonster()
		var m model.Monster
		if err == nil && mon != nil {
			m = model.Monster{
				Name:            mon.Name,
				Description:     string(mon.Description),
				HP:              mon.HitPoints,
				AC:              mon.ArmorClass,
				ChallengeRating: mon.ChallengeRating,
			}
		}
		return monsterFetchedMsg{Monster: m, Err: err}
	}
}
