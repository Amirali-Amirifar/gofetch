package components

import (
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TabKeyMap holds key bindings for a specific tab
type TabKeyMap struct {
	Bindings []key.Binding
	Name     string
}

// keyMap defines a set of keybindings. To work for help it must satisfy
// key.Map.
type keyMap struct {
	TabBindings map[string]TabKeyMap
	Help        key.Binding
	Quit        key.Binding
	ActiveTab   string
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k keyMap) ShortHelp() []key.Binding {
	if tab, ok := k.TabBindings[k.ActiveTab]; ok {
		return append(tab.Bindings, k.Help, k.Quit)
	}
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
func (k keyMap) FullHelp() [][]key.Binding {
	if tab, ok := k.TabBindings[k.ActiveTab]; ok {
		return [][]key.Binding{
			tab.Bindings,
			{k.Help, k.Quit},
		}
	}
	return [][]key.Binding{
		{k.Help, k.Quit},
	}
}

var defaultKeys = keyMap{
	TabBindings: make(map[string]TabKeyMap),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "toggle help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	ActiveTab: "",
}

type HelpModel struct {
	keys       keyMap
	help       help.Model
	inputStyle lipgloss.Style
	quitting   bool
}

func (m HelpModel) Init() tea.Cmd {
	return nil
}

func (m HelpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// If we set a width on the help menu it can gracefully truncate
		// its view as needed.
		m.help.Width = msg.Width
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		}
	}

	return m, nil
}

func (m HelpModel) View() string {
	var status string
	helpView := m.help.View(m.keys)
	return "\n" + status + helpView
}

func (m HelpModel) SetKeyMap(keyMap map[string]TabKeyMap) HelpModel {
	m.keys.TabBindings = keyMap
	return m
}

func (m HelpModel) SetActiveTab(activeTab string) HelpModel {
	m.keys.ActiveTab = activeTab
	return m
}
func InitHelp() HelpModel {
	return HelpModel{
		keys:       defaultKeys,
		help:       help.New(),
		inputStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#FF75B7")),
	}
}
