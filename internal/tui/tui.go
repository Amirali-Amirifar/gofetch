package tui

import (
	"github.com/Amirali-Amirifar/gofetch.git/internal/tui/views"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"strings"
)

type model struct {
	Tabs      []string
	children  []tea.Model
	activeTab int
	isFocused bool

	// Terminal size
	width  int
	height int
}

func (m model) Init() tea.Cmd {
	return m.children[0].Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	m.children[m.activeTab], _ = m.children[m.activeTab].Update(msg)

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "right", "l", "n", "tab":
			m.activeTab = min(m.activeTab+1, len(m.Tabs)-1)
			return m, nil
		case "left", "h", "p", "shift+tab":
			m.activeTab = max(m.activeTab-1, 0)
			return m, nil
		}
	}

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
	m.children = []tea.Model{nil, nil, nil}
	m.children[0] = views.InitialModel()
	m.children[0].Init()

	return m
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func GetTui() *tea.Program {
	tabs := []string{"New Download", "Downloads List", "Queues List"}
	m := model{Tabs: tabs}.initializeChildren()

	return tea.NewProgram(m)
}
