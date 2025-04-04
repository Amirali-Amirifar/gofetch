package tui

import (
	"github.com/Amirali-Amirifar/gofetch.git/internal/tui/components"
	"strings"

	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
	"github.com/Amirali-Amirifar/gofetch.git/internal/tui/views"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	Tabs          []string
	activeTab     int
	isFocusedTab  bool
	width         int
	height        int
	state         models.AppState
	children      []ChildModel
	HelpComponent components.HelpModel
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// Handle cases where tab focus doesn't matter
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.updateSizeMsg(msg)
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.isFocusedTab = !m.isFocusedTab
			m.HelpComponent = m.HelpComponent.SetIsFocusedTab(m.isFocusedTab)
			return m, nil

		}
	}

	// Delegate update to the active view
	if m.isFocusedTab {
		updatedChild, cmd := m.children[m.activeTab].Update(msg)
		if child, ok := updatedChild.(ChildModel); ok {
			m.children[m.activeTab] = child
		} else {
			panic(`invalid child model`)
		}
		return m, cmd
	} else {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "right", "tab":
				return m.handleTabChange(min(m.activeTab+1, len(m.Tabs)-1)), nil
			case "left", "shift+tab":
				return m.handleTabChange(max(m.activeTab-1, 0)), nil
			case "1":
				return m.handleTabChange(0), nil
			case "2":
				return m.handleTabChange(1), nil
			case "3":
				return m.handleTabChange(2), nil
			case "?":
				updatedHelp, helpCmd := m.HelpComponent.Update(msg)
				if help, ok := updatedHelp.(components.HelpModel); ok {
					m.HelpComponent = help
				}
				return m, helpCmd
			}
		}
	}

	return m, cmd
}

func (m model) handleTabChange(newTab int) model {
	m.activeTab = newTab
	m.HelpComponent = m.HelpComponent.SetActiveTab(m.children[m.activeTab].GetName())
	return m
}

func tabBorderWithBottom(left, middle, right string) lipgloss.Border {
	border := lipgloss.RoundedBorder()
	border.BottomLeft = left
	border.Bottom = middle
	border.BottomRight = right
	return border
}

var (
	inactiveTabBorder = tabBorderWithBottom("┴", "─", "┴")
	activeTabBorder   = tabBorderWithBottom("┘", " ", "└")
	docStyle          = lipgloss.NewStyle().Padding(1, 2, 1, 2)
	highlightColor    = lipgloss.Color("#7D56F4")
	inactiveTabStyle  = lipgloss.NewStyle().Border(inactiveTabBorder, true).BorderForeground(highlightColor).Padding(0, 1)
	activeTabStyle    = inactiveTabStyle.Border(activeTabBorder, true)
	windowStyle       = lipgloss.NewStyle().
				BorderForeground(highlightColor).
				Padding(3, 3).
				Align(lipgloss.Center).
				Border(lipgloss.NormalBorder()).
				UnsetBorderTop()
)

func (m model) View() string {
	doc := strings.Builder{}

	var renderedTabs []string

	tabBarWidth := m.width
	for i, t := range m.Tabs {
		var style lipgloss.Style
		isFirst, isLast, isActive := i == 0, i == len(m.Tabs)-1, i == m.activeTab
		if isActive {
			style = activeTabStyle
		} else {
			style = inactiveTabStyle
		}

		border, _, _, _, _ := style.GetBorder()
		if isFirst && isActive {
			border.BottomLeft = "│"
		} else if isFirst {
			border.BottomLeft = "├"
		} else if isLast && isActive {
			border.BottomRight = "└"
		}

		style = style.Width(20).Border(border)
		renderedText := style.Render(t)

		renderedTabs = append(renderedTabs, renderedText)
		tabBarWidth = tabBarWidth - lipgloss.Width(renderedText)
	}

	blankBorder := lipgloss.HiddenBorder()
	blankBorder.Bottom = "─"
	blankBorder.BottomLeft = "─"
	blankBorder.BottomRight = "┐"
	blankTab := lipgloss.NewStyle().
		Width(tabBarWidth - windowStyle.GetHorizontalFrameSize()).
		Border(blankBorder).
		BorderForeground(highlightColor).
		Render("")

	renderedTabs = append(renderedTabs, blankTab)
	row := lipgloss.JoinHorizontal(lipgloss.Top, renderedTabs...)

	tabContents := windowStyle.Width(
		m.width - windowStyle.GetHorizontalFrameSize()).Height(m.height - windowStyle.GetVerticalFrameSize()).
		// Background is set for debugging
		Background(lipgloss.Color("#111")).
		Render(m.children[m.activeTab].View())

	doc.WriteString(row)
	doc.WriteString("\n")
	doc.WriteString(tabContents)
	doc.WriteString(m.HelpComponent.View())

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		docStyle.Render(doc.String()))
}

func (m model) initializeChildren() model {
	// Initialize child models with the loaded state
	m.children = []ChildModel{
		views.InitDownloads(m.state),    // New Download tab
		views.InitDownloadList(m.state), // Downloads List tab
		views.InitQueueList(m.state),    // Queues List tab
	}

	// Populate Tabs dynamically from children's GetName()
	m.Tabs = make([]string, len(m.children))
	keysMap := make(map[string]components.TabKeyMap)
	for i, child := range m.children {
		m.Tabs[i] = child.GetName()
		keysMap[child.GetName()] = components.TabKeyMap{Bindings: child.GetKeyBinds(), Name: child.GetName()}
	}

	m.HelpComponent = components.InitHelp()
	m.HelpComponent = m.HelpComponent.SetKeyMap(keysMap)
	m.handleTabChange(1)
	return m
}

func (m model) updateSizeMsg(msg tea.WindowSizeMsg) (model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height
	updatedHelp, _ := m.HelpComponent.Update(msg)
	if help, ok := updatedHelp.(components.HelpModel); ok {
		m.HelpComponent = help
	}

	for i, _ := range m.children {
		updatedChild, _ := m.children[i].Update(msg)
		if child, ok := updatedChild.(ChildModel); ok {
			m.children[i] = child
		} else {
			panic(`invalid child model`)
		}
	}
	return m, nil
}

func GetTui(state models.AppState) *tea.Program {
	m := model{
		state: state,
	}.initializeChildren()
	m = m.handleTabChange(1)

	return tea.NewProgram(m)
}
