package views

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type queueListModel struct {
	table   table.Model
	state   models.AppState
	focused bool
}

func InitQueueList(state models.AppState) queueListModel {
	columns := []table.Column{
		{Title: "Name", Width: 15},
		{Title: "Folder", Width: 20},
		{Title: "Max DL", Width: 7},
		{Title: "Speed", Width: 10},
		{Title: "Time Range", Width: 15},
	}

	var rows []table.Row
	for _, q := range state.Queues {
		rows = append(rows, table.Row{q.Name, q.Folder, fmt.Sprintf("%d", q.MaxDL), q.Speed, q.TimeRange})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	return queueListModel{table: t, state: state, focused: true}
}

func (m queueListModel) Init() tea.Cmd {
	return nil
}

func (m queueListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			m.focused = !m.focused
			if m.focused {
				m.table.Focus()
			} else {
				m.table.Blur()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "n": // Add new queue
			newQueue := models.Queue{Name: "New Queue", Folder: "~/Downloads", MaxDL: 3, Speed: "1MB/s", TimeRange: "Anytime"}
			m.state.Queues = append(m.state.Queues, newQueue)
			m.updateTableRows()
			if err := m.saveQueuesToFile(); err != nil {
				return m, tea.Printf("Error saving queues: %v", err)
			}
		case "e": // Edit selected queue
			if len(m.state.Queues) == 0 {
				return m, tea.Printf("No queues available to edit.")
			}
			idx := m.table.Cursor()
			if idx >= 0 && idx < len(m.state.Queues) {
				m.state.Queues[idx].Name = "Edited Queue" // Replace with actual edit logic
				m.updateTableRows()
				if err := m.saveQueuesToFile(); err != nil {
					return m, tea.Printf("Error saving queues: %v", err)
				}
			} else {
				return m, tea.Printf("Invalid selection.")
			}
		case "d": // Delete selected queue
			if len(m.state.Queues) == 0 {
				return m, tea.Printf("No queues available to delete.")
			}
			idx := m.table.Cursor()
			if idx >= 0 && idx < len(m.state.Queues) {
				m.state.Queues = append(m.state.Queues[:idx], m.state.Queues[idx+1:]...)
				m.updateTableRows()
				if err := m.saveQueuesToFile(); err != nil {
					return m, tea.Printf("Error saving queues: %v", err)
				}
			} else {
				return m, tea.Printf("Invalid selection.")
			}
		case "left", "right": // Tab navigation
			return m, func() tea.Msg {
				return models.SwitchTabMsg{Direction: msg.String()}
			}
		}
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m queueListModel) View() string {
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	renderedTable := baseStyle.Render(m.table.View())

	footer := lipgloss.NewStyle().
		Padding(1, 2).
		Align(lipgloss.Center).
		Render("N: New Queue | E: Edit | D: Delete | Q: Quit")

	return lipgloss.JoinVertical(lipgloss.Left, renderedTable, footer)
}

func (m *queueListModel) updateTableRows() {
	var rows []table.Row
	for _, q := range m.state.Queues {
		rows = append(rows, table.Row{q.Name, q.Folder, fmt.Sprintf("%d", q.MaxDL), q.Speed, q.TimeRange})
	}
	m.table.SetRows(rows)
}

func (m *queueListModel) saveQueuesToFile() error {
	data, err := json.MarshalIndent(m.state.Queues, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling queues: %w", err)
	}
	if err := os.WriteFile("queues.json", data, 0644); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
}
