package main

import (
	"fmt"
	"gotopia-rpg/model"
	"gotopia-rpg/ui"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	game := model.NewGame()
	m := ui.NewModel(game)
	p := tea.NewProgram(m)

	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
	}
}
