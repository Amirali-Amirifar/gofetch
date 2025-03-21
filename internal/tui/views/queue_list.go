package views

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type queueListModel struct {
	table      table.Model
	state      models.AppState
	focused    bool
	editing    bool
	editInputs []textinput.Model
}

func (m queueListModel) GetKeyBinds() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("N"), key.WithHelp("N", "New Queue")),
		key.NewBinding(key.WithKeys("E"), key.WithHelp("E", "Edit")),
		key.NewBinding(key.WithKeys("D"), key.WithHelp("D", "Delete")),
	}
}

func (m queueListModel) GetName() string {
	return "Queue List"
}

func InitQueueList(state models.AppState) queueListModel {
	columns := []table.Column{
		{Title: "Name", Width: 15},
		{Title: "Folder", Width: 20},
		{Title: "Max DL", Width: 7},
		{Title: "Speed", Width: 10},
		{Title: "Time Start", Width: 15},
		{Title: "Time End", Width: 15},
	}

	var rows []table.Row
	for _, q := range state.Queues {
		rows = append(rows, table.Row{q.Name, q.StorageFolder, fmt.Sprintf("%d", q.MaxSimultaneous), fmt.Sprintf("%d", q.MaxDownloadSpeed), q.ActiveTimeStart, q.ActiveTimeEnd})
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
		case "up", "down": // Handle Up and Down keys
			if !m.editing { // Only navigate queues if not in edit mode
				m.table, cmd = m.table.Update(msg)
				return m, cmd
			}
		case "tab", "shift+tab": // Handle Tab and Shift+Tab
			if m.editing { // Only navigate fields if in edit mode
				// Get the currently focused input field index
				currentIndex := -1
				for i, input := range m.editInputs {
					if input.Focused() {
						currentIndex = i
						break
					}
				}

				// Determine the next input field index
				if msg.String() == "tab" {
					currentIndex = (currentIndex + 1) % len(m.editInputs) // Move to the next field
				} else if msg.String() == "shift+tab" {
					currentIndex = (currentIndex - 1 + len(m.editInputs)) % len(m.editInputs) // Move to the previous field
				}

				// Update focus
				for i := range m.editInputs {
					if i == currentIndex {
						m.editInputs[i].Focus()
					} else {
						m.editInputs[i].Blur()
					}
				}
				return m, nil
			}
		case "enter":
			if m.editing {
				idx := m.table.Cursor()
				if idx >= 0 && idx < len(m.state.Queues) {
					m.applyEditInputs(idx)
					m.editing = false
					m.updateTableRows()
					if err := m.saveQueuesToFile(); err != nil {
						return m, tea.Printf("Error saving queues: %v", err)
					}
				}
			}
		case "e": // Enter edit mode
			if !m.editing && len(m.state.Queues) > 0 {
				m.editing = true
				m.initEditInputs(m.table.Cursor())
			}
		case "esc": // Exit edit mode
			if m.editing {
				m.editing = false
			}
			// Other key bindings...
		}
	}

	if m.editing {
		var cmds []tea.Cmd
		for i := range m.editInputs {
			m.editInputs[i], cmd = m.editInputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
		return m, tea.Batch(cmds...)
	}

	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m queueListModel) View() string {
	baseStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240"))

	renderedTable := baseStyle.Render(m.table.View())

	if m.editing {
		editView := lipgloss.JoinVertical(lipgloss.Left,
			"Edit Queue:",
			m.editInputs[0].View(),
			m.editInputs[1].View(),
			m.editInputs[2].View(),
			m.editInputs[3].View(),
			m.editInputs[4].View(),
			m.editInputs[5].View(),
		)
		return lipgloss.JoinVertical(lipgloss.Left, renderedTable, editView)
	}

	return lipgloss.JoinVertical(lipgloss.Left, renderedTable)
}

func (m *queueListModel) updateTableRows() {
	var rows []table.Row
	for _, q := range m.state.Queues {
		rows = append(rows, table.Row{q.Name, q.StorageFolder, fmt.Sprintf("%d", q.MaxSimultaneous), fmt.Sprintf("%d", q.MaxDownloadSpeed), q.ActiveTimeStart, q.ActiveTimeEnd})
	}
	m.table.SetRows(rows)
}

func (m *queueListModel) saveQueuesToFile() error {
	data, err := json.MarshalIndent(m.state, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling state: %w", err)
	}
	if err := os.WriteFile("state.json", data, 0644); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}
	return nil
}

func (m *queueListModel) initEditInputs(idx int) {
	queue := m.state.Queues[idx]
	m.editInputs = make([]textinput.Model, 6)

	m.editInputs[0] = textinput.New()
	m.editInputs[0].Placeholder = "Name"
	m.editInputs[0].SetValue(queue.Name)

	m.editInputs[1] = textinput.New()
	m.editInputs[1].Placeholder = "Folder"
	m.editInputs[1].SetValue(queue.StorageFolder)

	m.editInputs[2] = textinput.New()
	m.editInputs[2].Placeholder = "Max DL"
	m.editInputs[2].SetValue(fmt.Sprintf("%d", queue.MaxSimultaneous))

	m.editInputs[3] = textinput.New()
	m.editInputs[3].Placeholder = "Speed"
	m.editInputs[3].SetValue(fmt.Sprintf("%d", queue.MaxDownloadSpeed))

	m.editInputs[4] = textinput.New()
	m.editInputs[4].Placeholder = "Time Start"
	m.editInputs[4].SetValue(queue.ActiveTimeStart)

	m.editInputs[5] = textinput.New()
	m.editInputs[5].Placeholder = "Time End"
	m.editInputs[5].SetValue(queue.ActiveTimeEnd)
}

func (m *queueListModel) applyEditInputs(idx int) {
	queue := &m.state.Queues[idx]
	queue.Name = m.editInputs[0].Value()
	queue.StorageFolder = m.editInputs[1].Value()

	maxSimultaneous, err := strconv.Atoi(m.editInputs[2].Value())
	if err != nil {
		maxSimultaneous = 3
	}
	queue.MaxSimultaneous = maxSimultaneous

	maxDownloadSpeed, err := strconv.ParseInt(m.editInputs[3].Value(), 10, 64)
	if err != nil {
		maxDownloadSpeed = 1
	}
	queue.MaxDownloadSpeed = maxDownloadSpeed

	queue.ActiveTimeStart = m.editInputs[4].Value()
	queue.ActiveTimeEnd = m.editInputs[5].Value()
}
