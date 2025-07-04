package main

import (
	"fmt"
	"strings"

	"github.com/marcodenic/agentry/internal/tui"
)

func main() {
	// Create a progress bar to test initial state
	prog := tui.CreateTokenProgressBar()

	fmt.Println("Testing token progress bar...")
	fmt.Printf("Initial progress bar view: %s\n", prog.View())

	// Test with 0% (should be green)
	prog.SetPercent(0.0)
	view0 := prog.View()
	fmt.Printf("0%% progress bar view: %s\n", view0)

	// Test with 50% (should be orange/yellow)
	prog.SetPercent(0.5)
	view50 := prog.View()
	fmt.Printf("50%% progress bar view: %s\n", view50)

	// Test with 100% (should be red)
	prog.SetPercent(1.0)
	view100 := prog.View()
	fmt.Printf("100%% progress bar view: %s\n", view100)

	// Check if green color is present in 0% view
	if strings.Contains(view0, "22C55E") || strings.Contains(view0, "green") {
		fmt.Println("✓ 0% progress bar appears to be green")
	} else {
		fmt.Println("! 0% progress bar color check needs visual verification")
	}

	fmt.Println("Test completed. Visual verification may be needed for colors.")
}
