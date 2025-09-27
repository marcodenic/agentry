package tui

// Keybinds define keyboard shortcuts for the TUI actions.
type Keybinds struct {
	Quit        string `json:"quit"`
	ToggleTab   string `json:"toggleTab"`
	Submit      string `json:"submit"`
	NextPane    string `json:"nextPane"`
	PrevPane    string `json:"prevPane"`
	Pause       string `json:"pause"`
	Diagnostics string `json:"diagnostics"`
}

var defaultKeybinds = Keybinds{
	Quit:        "ctrl+c",
	ToggleTab:   "tab",
	Submit:      "enter",
	NextPane:    "ctrl+n",
	PrevPane:    "ctrl+p",
	Pause:       "ctrl+s",
	Diagnostics: "ctrl+d",
}

// DefaultKeybinds returns the static keybindings used by the TUI.
func DefaultKeybinds() Keybinds {
	return defaultKeybinds
}

// Hex colour constants used throughout the interface.
const (
	uiColorForegroundHex = "#FFFFFF"
	uiColorPanelTitleHex = "#9CA3AF"
	uiColorUserAccentHex = "#8B5CF6"
	uiColorRoleAccentHex = "#10B981"
	uiColorToolAccentHex = "#8B5CF6"
	uiColorAIAccentHex   = "#9CA3AF"

	uiColorIdleHex    = "#22C55E"
	uiColorRunningHex = "#FBBF24"
	uiColorErrorHex   = "#EF4444"
	uiColorStoppedHex = "#6B7280"
)
