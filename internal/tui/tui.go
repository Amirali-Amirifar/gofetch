package tui

import (
	"strings"

	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
	"github.com/Amirali-Amirifar/gofetch.git/internal/tui/views"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	Tabs      []string
	children  []tea.Model
	activeTab int
	width     int
	height    int
	state     models.AppState
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "right", "tab":
			m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
			return m, nil
		case "left", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		case "1":
			m.activeTab = 0
			return m, nil
		case "2":
			m.activeTab = 1
			return m, nil
		case "3":
			m.activeTab = 2
			return m, nil
		}
	}

	// Delegate update to the active view
	m.children[m.activeTab], cmd = m.children[m.activeTab].Update(msg)
	return m, cmd
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
		} else if isFirst && !isActive {
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

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		docStyle.Render(doc.String()))
}

func (m model) initializeChildren() tea.Model {
	// Initialize child models with the loaded state
	m.children = []tea.Model{nil, nil, nil}
	m.children[0] = views.InitDownloads(m.state)    // New Download tab
	m.children[1] = views.InitDownloadList(m.state) // Downloads List tab
	m.children[2] = views.InitQueueList(m.state)    // Queues List tab

	// Initialize all child models
	for i := range m.children {
		m.children[i].Init()
	}

	return m
}

func GetTui(state models.AppState) *tea.Program {
	tabs := []string{"New Download", "Downloads List", "Queues List"}
	m := model{
		Tabs:      tabs,
		activeTab: 1, // Default to Downloads List tab
		state:     state,
	}.initializeChildren()

	return tea.NewProgram(m)
}
