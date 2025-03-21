package views

import (
	"fmt"

	"github.com/Amirali-Amirifar/gofetch.git/internal/config"
	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/sirupsen/logrus"
)

type downloadListModel struct {
	table table.Model
	state models.AppState
}

func (m downloadListModel) GetKeyBinds() []key.Binding {
	return []key.Binding{}
}

func (m downloadListModel) GetName() string {
	return "Download List"
}

func InitDownloadList(state models.AppState) downloadListModel {
	columns := []table.Column{
		{Title: "URL", Width: 50},
		{Title: "Queue", Width: 15},
		{Title: "Status", Width: 15},
		{Title: "Progress", Width: 10},
	}

	var rows []table.Row
	db := config.GetDB()
	downloads, err := db.GetDownloads()
	if err != nil {
		log.Errorf("Failed to fetch downloads: %v", err)
	}

	for _, download := range downloads {
		rows = append(rows, table.Row{
			download.URL,
			download.Queue,
			string(download.Status),
			fmt.Sprintf("%d%%", download.Progress),
		})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(10),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	return downloadListModel{table: t, state: state}
}

func (m downloadListModel) Init() tea.Cmd {
	return nil
}

func (m downloadListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Printf("Selected: %s", m.table.SelectedRow()[0])
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m downloadListModel) View() string {
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	renderedTable := baseStyle.Render(m.table.View())

	return lipgloss.JoinVertical(lipgloss.Center, []string{"Download List Info", renderedTable}...) + "\n"
}
