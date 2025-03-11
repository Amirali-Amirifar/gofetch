package views

import (
	"fmt"

	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type TextInputModel struct {
	textInput textinput.Model
	state     models.AppState
	err       error
}

func InitDownloads(state models.AppState) TextInputModel {
	ti := textinput.New()
	ti.Placeholder = "https://..."
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 20

	return TextInputModel{
		textInput: ti,
		state:     state,
		err:       nil,
	}
}

func (m TextInputModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m TextInputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			// Add the URL to the downloads list in the state
			url := m.textInput.Value()
			m.state.Downloads = append(m.state.Downloads, models.Download{
				URL:    url,
				Queue:  "Default", // Default queue for now
				Status: "Pending",
			})
			return m, tea.Quit
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}

	case error:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m TextInputModel) View() string {
	return fmt.Sprintf(
		"Enter a URL to start download\n\n%s\n\n%s",
		m.textInput.View(),
		"(esc to quit)",
	) + "\n"
}
