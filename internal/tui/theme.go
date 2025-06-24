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
// Palette defines a set of base colours for the UI panels.
type Palette struct {
	Background string `json:"background"`
	Foreground string `json:"foreground"`
}

// Theme holds colour settings and keybinds.
// Mode selects which built-in palette to use ("light" or "dark").
type Theme struct {
	Mode         string  `json:"mode"`
	Palette      Palette `json:"palette"`
	UserBarColor string  `json:"userBarColor"`
	AIBarColor   string  `json:"aiBarColor"`

	IdleColor    string `json:"idleColor"`
	RunningColor string `json:"runningColor"`
	ErrorColor   string `json:"errorColor"`
	StoppedColor string `json:"stoppedColor"`

	Keybinds Keybinds `json:"keybinds"`
}

// Pre-defined palettes for light and dark modes.
var (
	LightPalette = Palette{Background: "#FFFFFF", Foreground: "#000000"}
	DarkPalette  = Palette{Background: "#000000", Foreground: "#FFFFFF"}
)

// DefaultTheme returns built‑in colours and keybindings.
func DefaultTheme() Theme {
	return Theme{
		Mode:         "dark",
		Palette:      DarkPalette,
		UserBarColor: "#8B5CF6",
		AIBarColor:   "#9CA3AF",
		IdleColor:    "#22C55E",
		RunningColor: "#FBBF24",
		ErrorColor:   "#EF4444",
		StoppedColor: "#6B7280",
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
	if mode := os.Getenv("AGENTRY_THEME"); mode != "" {
		t.Mode = mode
	}
	if t.Mode == "light" {
		t.Palette = LightPalette
	} else {
		t.Palette = DarkPalette
	}
	return t
}
