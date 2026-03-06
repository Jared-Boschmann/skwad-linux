package models

import "github.com/google/uuid"

// LayoutMode determines how many terminal panes are shown and their arrangement.
type LayoutMode string

const (
	LayoutModeSingle          LayoutMode = "single"
	LayoutModeSplitVertical   LayoutMode = "splitVertical"
	LayoutModeSplitHorizontal LayoutMode = "splitHorizontal"
	LayoutModeThreePane       LayoutMode = "threePane"
	LayoutModeGridFourPane    LayoutMode = "gridFourPane"
)

// PaneCount returns the number of panes for a given layout mode.
func (l LayoutMode) PaneCount() int {
	switch l {
	case LayoutModeSplitVertical, LayoutModeSplitHorizontal:
		return 2
	case LayoutModeThreePane:
		return 3
	case LayoutModeGridFourPane:
		return 4
	default:
		return 1
	}
}

// WorkspaceColors is the predefined color palette for workspaces.
var WorkspaceColors = []string{
	"#4A90D9", // blue
	"#7B4FD4", // purple
	"#C83FB5", // magenta
	"#8B7FD4", // lavender
	"#D44F8B", // pink
	"#E05B7A", // rosePink
	"#4FC8E0", // skyBlue
	"#00BCD4", // cyan
	"#00BFA5", // aqua
	"#8BC34A", // lime
	"#4CAF50", // green
	"#009688", // teal
	"#8D6E63", // mauve
	"#FF7043", // coral
	"#F44336", // red
	"#FF9800", // orange
	"#FFC107", // amber
	"#A1887F", // tan
}

// Workspace groups agents with a shared layout state.
type Workspace struct {
	ID                  uuid.UUID   `json:"id"`
	Name                string      `json:"name"`
	ColorHex            string      `json:"colorHex"`
	AgentIDs            []uuid.UUID `json:"agentIds"`
	LayoutMode          LayoutMode  `json:"layoutMode"`
	ActiveAgentIDs      []uuid.UUID `json:"activeAgentIds"`
	FocusedPaneIndex    int         `json:"focusedPaneIndex"`
	SplitRatio          float64     `json:"splitRatio"`
	SplitRatioSecondary float64     `json:"splitRatioSecondary"`
}

// WorstStatus returns the most severe status across the given agents.
// Order: input > running > idle (nil treated as idle).
func WorstStatus(agents []*Agent) AgentStatus {
	worst := AgentStatusIdle
	for _, a := range agents {
		switch a.Status {
		case AgentStatusInput, AgentStatusError:
			return a.Status
		case AgentStatusRunning:
			worst = AgentStatusRunning
		}
	}
	return worst
}
