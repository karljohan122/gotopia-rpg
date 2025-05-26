package ui

import (
	"fmt"
	"gotopia-rpg/api"
	"gotopia-rpg/model"
	"math/rand"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/reflow/wordwrap"
)

type Turn int

const (
	PlayerTurn Turn = iota
	MonsterTurn
)

const maxHP = 100

type Model struct {
	Game          *model.Game
	Quitting      bool
	Loading       bool
	Spinner       spinner.Model
	Err           error
	Turn          Turn
	Message       string // general message (resting, etc.)
	PlayerAction  string // latest player attack message
	MonsterAction string // latest monster attack message
	restUsed      bool
	Level         int
	width, height int
}

type monsterFetchedMsg struct {
	Monster model.Monster
	Err     error
}

type monsterAttackMsg struct{}

func NewModel(g *model.Game) Model {
	rand.Seed(time.Now().UnixNano())
	s := spinner.New()
	s.Spinner = spinner.Dot
	return Model{Game: g, Spinner: s}
}

func (m Model) Init() tea.Cmd { return nil }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch v := msg.(type) {

	case tea.KeyMsg:
		switch v.String() {

		case "q", "ctrl+c":
			m.Quitting = true
			return m, tea.Quit

		case "n":
			if m.Game.Scene == model.SceneSpawn && !m.Loading {
				if m.Game.Player.HP <= 0 {
					m.Game.Player.HP = maxHP
				}
				m.Loading = true
				m.Err = nil
				m.Message = ""
				m.PlayerAction = ""
				m.MonsterAction = ""
				m.Turn = PlayerTurn
				cmds = append(cmds, fetchMonsterCmd(), m.Spinner.Tick)
				return m, tea.Batch(cmds...)
			}

		case "r":
			if m.Game.Scene == model.SceneSpawn && !m.Loading &&
				!m.restUsed && m.Game.Player.HP > 0 && m.Game.Player.HP < maxHP {

				heal := 50
				if m.Game.Player.HP+heal > maxHP {
					heal = maxHP - m.Game.Player.HP
				}
				m.Game.Player.HP += heal
				m.restUsed = true
				m.Message = fmt.Sprintf("You rest and regain %d HP.", heal)
				return m, nil
			}

		case "1":
			if m.Game.Scene == model.SceneBattle &&
				m.Turn == PlayerTurn &&
				m.Game.CurrentMon.HP > 0 &&
				m.Game.Player.HP > 0 {

				damage := rand.Intn(9) // 0–8 (0 means miss)

				if damage == 0 {
					m.PlayerAction = fmt.Sprintf("You miss %s!", m.Game.CurrentMon.Name)
				} else {
					m.Game.CurrentMon.HP -= damage
					m.PlayerAction = fmt.Sprintf("You hit %s for %d damage!", m.Game.CurrentMon.Name, damage)
				}

				if m.Game.CurrentMon.HP <= 0 {
					m.Game.CurrentMon.HP = 0
					m.Level++
					m.PlayerAction = fmt.Sprintf("You have slain %s!", m.Game.CurrentMon.Name)
					m.Game.Scene = model.SceneSpawn
					m.Turn = PlayerTurn
					m.restUsed = false
					return m, nil
				}

				m.Turn = MonsterTurn
				cmds = append(cmds, monsterAttackCmd())
				return m, tea.Batch(cmds...)
			}
		}

	case tea.WindowSizeMsg:
		m.width, m.height = v.Width, v.Height
		return m, nil

	case monsterFetchedMsg:
		m.Loading = false
		m.Message = ""
		if v.Err != nil {
			m.Err = v.Err
			m.Game.Scene = model.SceneSpawn
		} else {
			m.Game.CurrentMon = v.Monster
			m.Game.Scene = model.SceneBattle
			m.Err = nil
			m.Turn = PlayerTurn
		}

	case monsterAttackMsg:
		if m.Game.Scene == model.SceneBattle &&
			m.Turn == MonsterTurn &&
			m.Game.CurrentMon.HP > 0 &&
			m.Game.Player.HP > 0 {

			damage := rand.Intn(9) // 0–8

			if damage == 0 {
				m.MonsterAction = fmt.Sprintf("%s misses you!", m.Game.CurrentMon.Name)
			} else {
				m.Game.Player.HP -= damage
				m.MonsterAction = fmt.Sprintf("%s hits you for %d damage!", m.Game.CurrentMon.Name, damage)
			}

			if m.Game.Player.HP <= 0 {
				m.Game.Player.HP = 0
				m.Game.Scene = model.SceneSpawn
				m.MonsterAction = "You died!"
				m.restUsed = false
				m.Level = 0
			}
			m.Turn = PlayerTurn
		}

	case spinner.TickMsg:
		if m.Loading {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(v)
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

	if m.width == 0 {
		m.width = 80
	}

	header := "-- Gotopia RPG"
	if m.Level > 0 {
		header += fmt.Sprintf(" - Level %d", m.Level)
	}
	header += " -- \n"

	s := header

	// Spawn scene messages (resting, errors, etc.)
	if m.Game.Scene == model.SceneSpawn && m.Message != "" {
		s += wordwrap.String(m.Message, max(10, m.width-2)) + "\n\n"
	}

	if m.Loading {
		s += fmt.Sprintf("%s Fetching a monster...\n", m.Spinner.View())
		return s
	}
	if m.Err != nil {
		s += fmt.Sprintf("Error: %v\n", m.Err)
	}

	switch m.Game.Scene {

	case model.SceneSpawn:
		s += fmt.Sprintf("Player HP:%d AC:%d\n\n", m.Game.Player.HP, m.Game.Player.ArmorClass)
		s += "[n] Next battle\n"
		if !m.restUsed && m.Game.Player.HP > 0 && m.Game.Player.HP < maxHP {
			s += "[r] Rest\n"
		}
		s += "[q] Quit\n"

	case model.SceneBattle:
		desc := m.Game.CurrentMon.Description
		if desc == "" || desc == "False" {
			desc = "No description available."
		}
		s += fmt.Sprintf("Description:\n%s\n\n", wordwrap.String(desc, max(10, m.width-2)))

		s += fmt.Sprintf("Enemy: %s\n", m.Game.CurrentMon.Name)
		s += fmt.Sprintf("Enemy HP:%d AC:%d\n\n", m.Game.CurrentMon.HP, m.Game.CurrentMon.AC)
		s += fmt.Sprintf("Turn: %s\n\n", map[Turn]string{PlayerTurn: "Player", MonsterTurn: "Monster"}[m.Turn])

		if m.MonsterAction != "" {
			s += wordwrap.String(m.MonsterAction, max(10, m.width-2)) + "\n"
		}
		if m.PlayerAction != "" {
			s += wordwrap.String(m.PlayerAction, max(10, m.width-2)) + "\n"
		}
		if m.MonsterAction != "" || m.PlayerAction != "" {
			s += "\n"
		}

		s += fmt.Sprintf("Player HP:%d AC:%d\n", m.Game.Player.HP, m.Game.Player.ArmorClass)

		if m.Turn == PlayerTurn {
			s += "[1] Attack  [q] Quit"
		} else {
			if len(m.Game.CurrentMon.Attacks) > 0 {
				s += fmt.Sprintf("Monster uses %s!\n", m.Game.CurrentMon.Attacks[0].Name)
			} else {
				s += "Monster attacks!\n"
			}
		}
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
				Description:     mon.Description,
				HP:              mon.HitPoints,
				AC:              mon.ArmorClass,
				ChallengeRating: mon.ChallengeRating,
			}
		}
		return monsterFetchedMsg{Monster: m, Err: err}
	}
}

func monsterAttackCmd() tea.Cmd {
	return tea.Tick(200*time.Millisecond, func(t time.Time) tea.Msg {
		return monsterAttackMsg{}
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
