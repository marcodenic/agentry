package tui

import (
	"math/rand"
	"time"

	"github.com/charmbracelet/lipgloss"
)

// RobotState represents different emotional/activity states of the robot
type RobotState int

const (
	RobotIdle RobotState = iota
	RobotActive
	RobotThinking
	RobotError
	RobotSleeping
	RobotBlinking
)

// RobotFace represents our cute little robot companion
type RobotFace struct {
	state        RobotState
	lastBlink    time.Time
	blinkCounter int
	colorPhase   int
	lastUpdate   time.Time
}

// NewRobotFace creates a new robot face
func NewRobotFace() *RobotFace {
	return &RobotFace{
		state:      RobotIdle,
		lastBlink:  time.Now(),
		lastUpdate: time.Now(),
	}
}

// Update advances the robot's animation state
func (r *RobotFace) Update() {
	now := time.Now()
	
	// Blink every 3-5 seconds when not in special states
	if r.state == RobotIdle || r.state == RobotActive {
		timeSinceBlink := now.Sub(r.lastBlink)
		if timeSinceBlink > time.Duration(3+rand.Intn(3))*time.Second {
			r.state = RobotBlinking
			r.lastBlink = now
			r.blinkCounter = 3 // Blink for 3 frames
		}
	}
	
	// Handle blinking state
	if r.state == RobotBlinking {
		r.blinkCounter--
		if r.blinkCounter <= 0 {
			r.state = RobotIdle
		}
	}
	
	// Update color phase for thinking state
	if r.state == RobotThinking {
		timeSinceUpdate := now.Sub(r.lastUpdate)
		if timeSinceUpdate > 200*time.Millisecond {
			r.colorPhase = (r.colorPhase + 1) % 6
			r.lastUpdate = now
		}
	}
	
	// Add gentle breathing animation for idle state - flicker every 3 seconds for demo
	if r.state == RobotIdle {
		timeSinceUpdate := now.Sub(r.lastUpdate)
		if timeSinceUpdate > 3*time.Second {
			r.colorPhase = (r.colorPhase + 1) % 8
			r.lastUpdate = now
		}
	}
}

// SetState changes the robot's emotional state
func (r *RobotFace) SetState(state RobotState) {
	if r.state != state {
		r.state = state
		r.lastUpdate = time.Now()
		r.colorPhase = 0
	}
}

// GetFace returns the current face string based on state
func (r *RobotFace) GetFace() string {
	switch r.state {
	case RobotIdle:
		// Breathing animation: flicker to smaller squares briefly every 10 seconds
		if r.colorPhase%4 < 1 { // Show smaller squares for 1/4 of the breathing cycle (quick flicker)
			return "[▪_▪]" // Smaller squares for breathing effect
		}
		return "[■_■]" // Normal large squares
	case RobotActive:
		return "[●_●]"
	case RobotThinking:
		// Animated thinking eyes
		switch r.colorPhase % 4 {
		case 0:
			return "[~_~]"
		case 1:
			return "[¬_¬]"
		case 2:
			return "[°_°]"
		case 3:
			return "[•_•]"
		default:
			return "[~_~]"
		}
	case RobotError:
		return "[O_O]"
	case RobotSleeping:
		return "[--_--]"
	case RobotBlinking:
		// Cute blinking sequence
		switch r.blinkCounter % 3 {
		case 0:
			return "[^_^]"
		case 1:
			return "[#_#]"
		case 2:
			return "[＊_＊]"
		default:
			return "[^_^]"
		}
	default:
		return "[■_■]"
	}
}

// GetStyledFace returns the robot face with appropriate styling
func (r *RobotFace) GetStyledFace() string {
	face := r.GetFace()
	
	switch r.state {
	case RobotIdle:
		return r.renderMultiColorFace(face, "#8B5FBF", true) // Pinkish purple from agentry logo
			
	case RobotActive:
		return r.renderMultiColorFace(face, "#32CD32", true) // Toned down green with transparency
			
	case RobotThinking:
		// Cycle through rainbow colors with multi-color rendering
		colors := []string{"#FF6B6B", "#4ECDC4", "#45B7D1", "#96CEB4", "#FFEAA7", "#DDA0DD"}
		color := colors[r.colorPhase]
		return r.renderMultiColorFace(face, color, false) // No transparency for thinking state
			
	case RobotError:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4444")). // Red - error
			Bold(true).
			Background(lipgloss.Color("#2D1B1B")).
			Blink(true).
			Render(face)
			
	case RobotSleeping:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#696969")). // Dim gray - sleeping
			Faint(true).
			Render(face)
			
	case RobotBlinking:
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFD700")). // Gold - blinking
			Bold(true).
			Render(face)
			
	default:
		return r.renderMultiColorFace(face, "#8B5FBF", true)
	}
}

// renderMultiColorFace renders a robot face with a single consistent style
func (r *RobotFace) renderMultiColorFace(face, baseColor string, withTransparency bool) string {
	// Use single color for everything to ensure eyes are always identical
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color(baseColor)).
		Bold(true)
	
	// Render the entire face with the same style
	return style.Render(face)
}

// lightenColor lightens a hex color by the given factor (0.0 to 1.0)
func (r *RobotFace) lightenColor(hexColor string, factor float64) string {
	// Simple color lightening - predefined lighter shades for different colors
	switch hexColor {
	case "#8B5FBF": // Pinkish purple from agentry logo
		return "#B485D1" // Lighter pinkish purple for brackets/ears
	case "#32CD32": // Green
		return "#66E066" // Lighter green for brackets/ears
	case "#FF6B6B": // Red (thinking)
		return "#FF9999" // Lighter red
	case "#4ECDC4": // Teal (thinking) 
		return "#7FDBDA" // Lighter teal
	case "#45B7D1": // Blue (thinking)
		return "#78C8E0" // Lighter blue
	case "#96CEB4": // Light green (thinking)
		return "#B4DCC4" // Lighter light green
	case "#FFEAA7": // Yellow (thinking)
		return "#FFF2C7" // Lighter yellow
	case "#DDA0DD": // Plum (thinking)
		return "#E8B5E8" // Lighter plum
	default:
		return hexColor
	}
}

// darkenColor darkens a hex color by the given factor (0.0 to 1.0)  
func (r *RobotFace) darkenColor(hexColor string, factor float64) string {
	// Simple color darkening - predefined darker shades for mouth
	switch hexColor {
	case "#8B5FBF": // Pinkish purple from agentry logo
		return "#6B3F9F" // Darker pinkish purple for mouth
	case "#32CD32": // Green
		return "#228B22" // Darker green for mouth
	case "#FF6B6B": // Red (thinking)
		return "#CC4444" // Darker red
	case "#4ECDC4": // Teal (thinking)
		return "#3BA8A1" // Darker teal
	case "#45B7D1": // Blue (thinking)
		return "#3691B1" // Darker blue
	case "#96CEB4": // Light green (thinking)
		return "#7AB89A" // Darker light green
	case "#FFEAA7": // Yellow (thinking)
		return "#E6D197" // Darker yellow
	case "#DDA0DD": // Plum (thinking)
		return "#C480C4" // Darker plum
	default:
		return hexColor
	}
}

// GetMoodText returns a cute status text based on the robot's state
func (r *RobotFace) GetMoodText() string {
	switch r.state {
	case RobotIdle:
		return "ready"
	case RobotActive:
		return "working"
	case RobotThinking:
		return "thinking..."
	case RobotError:
		return "error"
	case RobotSleeping:
		return "sleeping"
	case RobotBlinking:
		return "blinking"
	default:
		return "idle"
	}
}

// GetStyledMoodText returns the mood text with appropriate styling
func (r *RobotFace) GetStyledMoodText() string {
	mood := r.GetMoodText()
	
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A8A8A8")).
		Italic(true).
		Faint(true).
		Render(mood)
}

// updateRobotState updates the robot's emotional state based on Agent 0's status
func (m *Model) updateRobotState() {
	if m.robot == nil {
		return
	}
	
	// Check if Agent 0 exists and update robot state accordingly
	if len(m.agents) > 0 && len(m.order) > 0 {
		agent0ID := m.order[0] // Agent 0 is always first
		if info, ok := m.infos[agent0ID]; ok {
			switch info.Status {
			case StatusRunning:
				if info.TokensStarted {
					m.robot.SetState(RobotThinking)
				} else {
					m.robot.SetState(RobotActive)
				}
			case StatusError:
				m.robot.SetState(RobotError)
			case StatusStopped:
				m.robot.SetState(RobotSleeping)
			default:
				m.robot.SetState(RobotIdle)
			}
		}
	}
}