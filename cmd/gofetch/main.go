package main

import (
	"fmt"
	"github.com/Amirali-Amirifar/gofetch.git/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func main() {
	tabs := []string{"New Download", "Downloads List", "Queues List"}
	tabContent := []string{"Lip Gloss Tab", "Blush Tab", "Eye Shadow Tab", "Mascara Tab", "Foundation Tab"}
	m := tui.Model{Tabs: tabs, TabContent: tabContent}
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
