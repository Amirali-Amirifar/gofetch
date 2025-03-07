package main

import (
	"fmt"
	"github.com/Amirali-Amirifar/gofetch.git/internal/tui"
)

func main() {
	if _, err := tui.GetTui().Run(); err != nil {
		fmt.Println(err)
	}
}
