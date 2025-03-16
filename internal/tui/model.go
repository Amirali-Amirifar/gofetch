package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type ChildModel interface {
	tea.Model
	GetName() string
	GetKeyBinds() []key.Binding
}
