package tui

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Keybinds define keyboard shortcuts for the TUI actions.
type Keybinds struct {
	Quit      string `json:"quit"`
	ToggleTab string `json:"toggleTab"`
	Submit    string `json:"submit"`
	NextPane  string `json:"nextPane"`
	PrevPane  string `json:"prevPane"`
	Pause     string `json:"pause"`
}

// Theme holds colour settings and keybinds.
type Theme struct {
	UserBarColor string   `json:"userBarColor"`
	AIBarColor   string   `json:"aiBarColor"`
	Keybinds     Keybinds `json:"keybinds"`
}

// DefaultTheme returns built‑in colours and keybindings.
func DefaultTheme() Theme {
	return Theme{
		UserBarColor: "#8B5CF6",
		AIBarColor:   "#9CA3AF",
		Keybinds: Keybinds{
			Quit:      "ctrl+c",
			ToggleTab: "tab",
			Submit:    "enter",
			NextPane:  "ctrl+n",
			PrevPane:  "ctrl+p",
			Pause:     "ctrl+s",
		},
	}
}

// LoadTheme loads theme.json from the current directory hierarchy or
// "$HOME/.config/agentry". Local files override global ones.
func LoadTheme() Theme {
	t := DefaultTheme()

	if home, err := os.UserHomeDir(); err == nil {
		if b, err := os.ReadFile(filepath.Join(home, ".config", "agentry", "theme.json")); err == nil {
			_ = json.Unmarshal(b, &t)
		}
	}

	dir, err := os.Getwd()
	if err != nil {
		return t
	}
	for {
		if b, err := os.ReadFile(filepath.Join(dir, "theme.json")); err == nil {
			_ = json.Unmarshal(b, &t)
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return t
}
