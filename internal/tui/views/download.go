package views

import (
	"strings"

	"github.com/Amirali-Amirifar/gofetch.git/internal/controller"
	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	log "github.com/sirupsen/logrus"
)

// Simple button representation
type button struct {
	label  string
	action string
}

// Model for the download view
type model struct {
	startButton  button
	cancelButton button
	focusIndex   int
	inputs       []textinput.Model
	state        models.AppState
	err          error
	statusMsg    string
}

// Message when a button is pressed
type buttonPressedMsg struct {
	action string
}

func (m model) GetKeyBinds() []key.Binding {
	bindings := []key.Binding{
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select/confirm"),
		),
		key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "previous field"),
		),
	}
	return bindings
}

func (m model) GetName() string {
	return "Download Page"
}

func InitDownloads(state models.AppState) model {
	urlTextInput := textinput.New()
	urlTextInput.Placeholder = "https://..."
	urlTextInput.Focus()
	urlTextInput.Width = 40

	queueTextInput := textinput.New()
	queueTextInput.Placeholder = "..."
	queueTextInput.Width = 40
	queueTextInput.CharLimit = 100

	fileNameInput := textinput.New()
	fileNameInput.Placeholder = "Optional, relative or absolute path"
	fileNameInput.Width = 40

	// Store inputs in a slice for easier focus cycling
	inputs := []textinput.Model{urlTextInput, queueTextInput, fileNameInput}

	return model{
		startButton:  button{label: "Start Download", action: "start"},
		cancelButton: button{label: "Cancel", action: "cancel"},
		inputs:       inputs,
		focusIndex:   0, // Start with URL input focused
		state:        state,
		err:          nil,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// Cycle focus between inputs and buttons
			if m.focusIndex < len(m.inputs) {
				// Currently on an input, blur it
				m.inputs[m.focusIndex].Blur()
			}

			m.focusIndex = (m.focusIndex + 1) % (len(m.inputs) + 2) // +2 for the two buttons

			if m.focusIndex < len(m.inputs) {
				// Focus the next input if it's an input
				m.inputs[m.focusIndex].Focus()
			}

			return m, nil
		case "shift+tab":
			// Cycle focus between inputs and buttons backward
			if m.focusIndex < len(m.inputs) {
				// Currently on an input, blur it
				m.inputs[m.focusIndex].Blur()
			}

			// Go back one step (with wrapping)
			m.focusIndex = (m.focusIndex - 1 + (len(m.inputs) + 2)) % (len(m.inputs) + 2)

			if m.focusIndex < len(m.inputs) {
				// Focus the previous input if it's an input
				m.inputs[m.focusIndex].Focus()
			}

			return m, nil

		case "enter":
			// Handle button presses when focused on buttons
			if m.focusIndex == len(m.inputs) { // Start button
				return m, func() tea.Msg {
					// This is where you'd implement the actual download logic
					return buttonPressedMsg{action: m.startButton.action}
				}
			} else if m.focusIndex == len(m.inputs)+1 { // Cancel button
				return m, func() tea.Msg {
					return buttonPressedMsg{action: m.cancelButton.action}
				}
			}
		}

	case buttonPressedMsg:
		switch msg.action {
		case "start":
			// Handle start download action
			url := m.inputs[0].Value()
			queueName := m.inputs[1].Value()
			fileName := m.inputs[2].Value()

			if url == "" {
				m.statusMsg = "Error: URL cannot be empty"
				return m, nil
			}

			if queueName == "" {
				queueName = "Default"
			}

			download := models.Download{
				FileName:  fileName,
				URL:       url,
				QueueName: queueName,
				Status:    models.DownloadStatusQueued,
				Progress:  0,
			}

			// Add the download to the state
			m.state.Downloads = append(m.state.Downloads, download)

			// Create QueueManager for the selected queue
			var queueManager *controller.QueueManager
			for _, queue := range m.state.Queues {
				if queue.Name == queueName {
					queueManager = &controller.QueueManager{
						Queue:           queue,                                      // Set the Queue
						ActiveDownloads: 0,                                          // Initialize active downloads
						DownloadChannel: make(chan struct{}, queue.MaxSimultaneous), // Control simultaneous downloads
					}
					break
				}
			}

			if queueManager == nil {
				m.statusMsg = "Error: Queue not found"
				return m, nil
			}

			// Create the download in the controller, passing the queue manager
			c := &controller.Download{Download: download}
			c.Create(queueManager) // Pass the QueueManager to Create()

			log.Infof("Added download info %#v", download)

			m.statusMsg = "Download has been queued successfully!"
			m.clearInputs()
		case "cancel":
			m.clearInputs()
		}

	case error:
		m.err = msg
		return m, nil
	}

	// Handle updates to the currently focused input
	var cmd tea.Cmd

	// Update all inputs but only the focused one will actually respond to typing
	for i := range m.inputs {
		m.inputs[i], cmd = m.inputs[i].Update(msg)
	}

	return m, cmd
}

func (m model) View() string {
	doc := strings.Builder{}
	docStyles := lipgloss.NewStyle().
		Padding(4).
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#ffffff")).
		Foreground(lipgloss.Color("1"))

	// Style definitions
	normalButton := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#888888")).
		Padding(0, 3).
		Margin(0, 1)

	focusedButton := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#000000")).
		Background(lipgloss.Color("#00FF00")).
		Padding(0, 3).
		Margin(0, 1)

	// Build the view
	doc.WriteString("URL: ")
	doc.WriteString(m.inputs[0].View())
	doc.WriteString("\n\n")

	doc.WriteString("Queue: ")
	doc.WriteString(m.inputs[1].View())
	doc.WriteString("\n\n")

	doc.WriteString("File Name: ")
	doc.WriteString(m.inputs[2].View())
	doc.WriteString("\n\n")

	// Render buttons with appropriate styles
	startButtonStyle := normalButton
	if m.focusIndex == len(m.inputs) {
		startButtonStyle = focusedButton
	}

	cancelButtonStyle := normalButton
	if m.focusIndex == len(m.inputs)+1 {
		cancelButtonStyle = focusedButton
	}

	doc.WriteString(startButtonStyle.Render(m.startButton.label))
	doc.WriteString(cancelButtonStyle.Render(m.cancelButton.label))

	doc.WriteString("\n\n")

	// Display status message if present
	if m.statusMsg != "" {
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Margin(1, 0)

		if strings.HasPrefix(m.statusMsg, "Error") {
			statusStyle = statusStyle.Foreground(lipgloss.Color("#FF0000"))
		}
		doc.WriteString(statusStyle.Render(m.statusMsg))
		doc.WriteString("\n\n")
	}

	doc.WriteString("(Tab to switch fields, Enter to confirm)")

	return docStyles.Render(doc.String())
}

func (m model) clearInputs() {
	for i := range m.inputs {
		m.inputs[i].SetValue("")
	}
}
