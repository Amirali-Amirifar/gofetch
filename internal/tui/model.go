package tui

import tea "github.com/charmbracelet/bubbletea"

type ChildModel interface {
	tea.Model
	GetName() string
}
