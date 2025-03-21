package views

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Amirali-Amirifar/gofetch.git/internal/controller"
	"github.com/Amirali-Amirifar/gofetch.git/internal/models"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// button represents a simple clickable button.
type button struct {
	label  string
	action string
}

// progressMsg delivers progress updates to the view.
type progressMsg struct {
	progress float64 // between 0.0 and 1.0.
	speed    float64 // bytes per second.
}

// buttonPressedMsg is sent when a button is clicked.
type buttonPressedMsg struct {
	action string
}

// model represents the download view.
type model struct {
	// Input widgets.
	inputs     []textinput.Model
	focusIndex int

	// Action buttons.
	startButton  button
	cancelButton button

	// State and progress info.
	state           models.AppState
	err             error
	activeDownload  bool
	progressBar     progress.Model
	progressVal     float64 // A value between 0.0 and 1.0.
	speed           float64 // bytes per second.
	downloadControl *controller.Download
	startedAt       time.Time
}

// GetKeyBinds returns key bindings.
func (m model) GetKeyBinds() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select/confirm")),
		key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
		key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "previous field")),
	}
}

// GetName returns the view's name.
func (m model) GetName() string {
	return "Download Page"
}

// InitDownloads initializes the view.
func InitDownloads(state models.AppState) model {
	// Create input fields.
	urlInput := textinput.New()
	urlInput.Placeholder = "https://..."
	urlInput.Focus()
	urlInput.Width = 40

	queueInput := textinput.New()
	queueInput.Placeholder = "Queue"
	queueInput.Width = 40
	queueInput.CharLimit = 100

	fileNameInput := textinput.New()
	fileNameInput.Placeholder = "Optional, relative or absolute path"
	fileNameInput.Width = 40

	inputs := []textinput.Model{urlInput, queueInput, fileNameInput}
	prog := progress.New(progress.WithDefaultGradient())

	return model{
		startButton:    button{label: "Start Download", action: "start"},
		cancelButton:   button{label: "Cancel", action: "cancel"},
		inputs:         inputs,
		focusIndex:     0,
		state:          state,
		err:            nil,
		activeDownload: false,
		progressBar:    prog,
		progressVal:    0,
		speed:          0,
	}
}

// Init is the Bubble Tea initialization.
func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink)
}

// pollDownloadProgressCmd polls the active download for progress.
func pollDownloadProgressCmd(c *controller.Download) tea.Cmd {
	return tea.Tick(500*time.Millisecond, func(t time.Time) tea.Msg {
		progFloat := 0.0
		if c.ContentLength > 0 {
			progFloat = float64(c.CurrentProgress) / float64(c.ContentLength)
			if progFloat > 1.0 {
				progFloat = 1.0
			}
		}
		elapsed := time.Since(c.StartTime)
		speed := 0.0
		if elapsed.Seconds() > 0 {
			speed = float64(c.CurrentProgress) / elapsed.Seconds()
		}
		return progressMsg{progress: progFloat, speed: speed}
	})
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			if !m.activeDownload {
				if m.focusIndex < len(m.inputs) {
					m.inputs[m.focusIndex].Blur()
				}
				m.focusIndex = (m.focusIndex + 1) % (len(m.inputs) + 2) // +2 for buttons.
				if m.focusIndex < len(m.inputs) {
					m.inputs[m.focusIndex].Focus()
				}
			}
			return m, tea.ClearScreen
		case "shift+tab":
			if !m.activeDownload {
				if m.focusIndex < len(m.inputs) {
					m.inputs[m.focusIndex].Blur()
				}
				m.focusIndex = (m.focusIndex - 1 + (len(m.inputs) + 2)) % (len(m.inputs) + 2)
				if m.focusIndex < len(m.inputs) {
					m.inputs[m.focusIndex].Focus()
				}
			}
			return m, tea.ClearScreen
		case "enter":
			cmds := []tea.Cmd{tea.ClearScreen}
			if !m.activeDownload {
				if m.focusIndex == len(m.inputs) {
					cmds = append(cmds, func() tea.Msg {
						return buttonPressedMsg{action: m.startButton.action}
					})
				} else if m.focusIndex == len(m.inputs)+1 {
					cmds = append(cmds, func() tea.Msg {
						return buttonPressedMsg{action: m.cancelButton.action}
					})
				}
			}
			return m, tea.Batch(cmds...)
		case "q":
			if m.activeDownload {
				m.activeDownload = false
				m.progressVal = 0
				return m, tea.ClearScreen
			}
		}
	case buttonPressedMsg:
		switch msg.action {
		case "start":
			m.activeDownload = true
			m.progressVal = 0
			m.speed = 0

			// Gather input values.
			url := m.inputs[0].Value()
			queue := m.inputs[1].Value()
			fileName := m.inputs[2].Value()
			if queue == "" {
				queue = "Default"
			}

			download := models.Download{
				FileName: fileName,
				URL:      url,
				Queue:    queue,
				Status:   models.DownloadStatusQueued,
			}

			// Create and start the download.
			ctrl := &controller.Download{Download: download}
			m.downloadControl = ctrl
			// Run in a separate goroutine so the UI isn’t blocked.
			go ctrl.Create()
			log.Printf("Started download: %#v", download)

			// Begin polling for progress.
			return m, pollDownloadProgressCmd(ctrl)

		case "cancel":
			return m, tea.ClearScreen
		}
	case progressMsg:
		m.progressVal = msg.progress
		m.speed = msg.speed
		// Keep polling until complete.
		if m.progressVal < 1.0 {
			if m.downloadControl != nil {
				return m, pollDownloadProgressCmd(m.downloadControl)
			}
		}
		m.activeDownload = false
		m.progressVal = 1.0
		return m, nil
	case error:
		m.err = msg
		return m, nil
	}

	var cmds []tea.Cmd
	if !m.activeDownload {
		for i := range m.inputs {
			var cmd tea.Cmd
			m.inputs[i], cmd = m.inputs[i].Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	newModel, progCmd := m.progressBar.Update(msg)
	if pb, ok := newModel.(progress.Model); ok {
		m.progressBar = pb
	}
	cmds = append(cmds, progCmd)
	return m, tea.Batch(cmds...)
}

// View renders the UI.
func (m model) View() string {
	var b strings.Builder
	docStyles := lipgloss.NewStyle().
		Padding(2).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("205")).
		Foreground(lipgloss.Color("229"))

	if m.activeDownload {
		b.WriteString("Download in progress...\n\n")
		b.WriteString(m.progressBar.ViewAs(m.progressVal))
		b.WriteString(fmt.Sprintf("\nSpeed: %.2f bytes/s\n", m.speed))
		b.WriteString("\n(Press 'q' to quit progress view)")
		return docStyles.Render(b.String())
	}

	b.WriteString("URL: " + m.inputs[0].View() + "\n\n")
	b.WriteString("Queue: " + m.inputs[1].View() + "\n\n")
	b.WriteString("File Name: " + m.inputs[2].View() + "\n\n")

	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("7")).Background(lipgloss.Color("240")).Padding(0, 2)
	focusedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("0")).Background(lipgloss.Color("229")).Padding(0, 2)

	btns := []string{
		func() string {
			if m.focusIndex == len(m.inputs) {
				return focusedStyle.Render(m.startButton.label)
			}
			return normalStyle.Render(m.startButton.label)
		}(),
		func() string {
			if m.focusIndex == len(m.inputs)+1 {
				return focusedStyle.Render(m.cancelButton.label)
			}
			return normalStyle.Render(m.cancelButton.label)
		}(),
	}
	b.WriteString(strings.Join(btns, "  "))
	return docStyles.Render(b.String())
}
