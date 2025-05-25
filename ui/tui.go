package ui

import (
	"fmt"
	"gotopia-rpg/api"
	"gotopia-rpg/model"
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

type Model struct {
	Game     *model.Game
	Quitting bool
	Loading  bool
	Spinner  spinner.Model
	Err      error
	Turn     Turn

	width  int // <─ current terminal width
	height int // <─ current terminal height
}

type monsterFetchedMsg struct {
	Monster model.Monster
	Err     error
}

type monsterAttackMsg struct{}

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
				m.Turn = PlayerTurn
				cmds = append(cmds, fetchMonsterCmd(), m.Spinner.Tick)
				return m, tea.Batch(cmds...)
			}
		case "1":
			if m.Game.Scene == model.SceneBattle && m.Turn == PlayerTurn && m.Game.CurrentMon.HP > 0 && m.Game.Player.HP > 0 {
				// Player attacks monster
				m.Game.CurrentMon.HP -= 3
				if m.Game.CurrentMon.HP <= 0 {
					m.Game.Scene = model.SceneSpawn
				} else {
					m.Turn = MonsterTurn
					return m, monsterAttackCmd() // Waits 0.5s before monster attacks
				}
			}
		}
	case tea.WindowSizeMsg: // <─ remember the new size
		m.width, m.height = msg.Width, msg.Height
		return m, nil
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
			m.Turn = PlayerTurn
		}
	case monsterAttackMsg:
		if m.Game.Scene == model.SceneBattle && m.Turn == MonsterTurn && m.Game.CurrentMon.HP > 0 && m.Game.Player.HP > 0 {
			// Monster attacks player
			m.Game.Player.HP -= 3
			if m.Game.Player.HP <= 0 {
				m.Game.Player.HP = 0
			}
			m.Turn = PlayerTurn
		}
	case spinner.TickMsg:
		if m.Loading {
			var cmd tea.Cmd
			m.Spinner, cmd = m.Spinner.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	// Monster's turn: process only if it's monster's turn, battle is ongoing, and both are alive
	if m.Game.Scene == model.SceneBattle && m.Turn == MonsterTurn && m.Game.CurrentMon.HP > 0 && m.Game.Player.HP > 0 {
		// Monster attacks player
		m.Game.Player.HP -= 3 // You can make this more advanced
		if m.Game.Player.HP <= 0 {
			m.Game.Player.HP = 0 // Prevent negative HP
		}
		m.Turn = PlayerTurn
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
	if m.Game.Player.HP <= 0 {
		return "-- Gotopia RPG -- \nYou died!\nPress q to quit.\n"
	}

	// fallback width if we never received WindowSizeMsg yet
	if m.width == 0 {
		m.width = 80
	}

	s := "-- Gotopia RPG -- \n"
	if m.Loading {
		s += fmt.Sprintf("%s Fetching a monster...\n", m.Spinner.View())
		return s
	}
	if m.Err != nil {
		s += fmt.Sprintf("Error: %v\n", m.Err)
	}
	switch m.Game.Scene {
	case model.SceneSpawn:
		s += "Press [n] to start battle\n"
	case model.SceneBattle:
		desc := m.Game.CurrentMon.Description
		if desc == "" || desc == "False" {
			desc = "No description available."
		}
		descWrapped := wordwrap.String(desc, max(10, m.width-2))
		s += fmt.Sprintf("Description:\n%s\n\n", descWrapped)

		s += fmt.Sprintf("Enemy: %s\n", m.Game.CurrentMon.Name)
		s += fmt.Sprintf("Enemy HP:%d AC:%d\n\n", m.Game.CurrentMon.HP, m.Game.CurrentMon.AC)
		s += fmt.Sprintf("Turn: %s\n\n", map[Turn]string{PlayerTurn: "Player", MonsterTurn: "Monster"}[m.Turn])
		s += fmt.Sprintf("Player race: %s\n", m.Game.Player.Race)
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
				Description:     string(mon.Description),
				HP:              mon.HitPoints,
				AC:              mon.ArmorClass,
				ChallengeRating: mon.ChallengeRating,
			}
		}
		return monsterFetchedMsg{Monster: m, Err: err}
	}
}

func monsterAttackCmd() tea.Cmd {
	return tea.Tick(time.Second/2, func(t time.Time) tea.Msg {
		return monsterAttackMsg{}
	})
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
