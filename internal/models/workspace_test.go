package models

import (
	"testing"

	"github.com/google/uuid"
)

func TestLayoutMode_PaneCount(t *testing.T) {
	cases := []struct {
		mode  LayoutMode
		panes int
	}{
		{LayoutModeSingle, 1},
		{LayoutModeSplitVertical, 2},
		{LayoutModeSplitHorizontal, 2},
		{LayoutModeThreePane, 3},
		{LayoutModeGridFourPane, 4},
	}
	for _, tc := range cases {
		if got := tc.mode.PaneCount(); got != tc.panes {
			t.Errorf("%s: got %d panes, want %d", tc.mode, got, tc.panes)
		}
	}
}

func TestWorstStatus_Priority(t *testing.T) {
	idle := &Agent{Status: AgentStatusIdle}
	running := &Agent{Status: AgentStatusRunning}
	input := &Agent{Status: AgentStatusInput}

	if got := WorstStatus([]*Agent{idle}); got != AgentStatusIdle {
		t.Errorf("single idle: got %s", got)
	}
	if got := WorstStatus([]*Agent{idle, running}); got != AgentStatusRunning {
		t.Errorf("idle+running: got %s", got)
	}
	if got := WorstStatus([]*Agent{running, input}); got != AgentStatusInput {
		t.Errorf("running+input: got %s", got)
	}
	if got := WorstStatus([]*Agent{idle, running, input}); got != AgentStatusInput {
		t.Errorf("all three: got %s", got)
	}
	if got := WorstStatus(nil); got != AgentStatusIdle {
		t.Errorf("empty: got %s", got)
	}
}

func TestWorstStatus_Error(t *testing.T) {
	err := &Agent{Status: AgentStatusError}
	running := &Agent{Status: AgentStatusRunning}
	// Error should surface like Input (both are "worst").
	result := WorstStatus([]*Agent{running, err})
	if result != AgentStatusError {
		t.Errorf("error status not surfaced: got %s", result)
	}
}

func TestWorkspaceColors_Count(t *testing.T) {
	if len(WorkspaceColors) == 0 {
		t.Error("WorkspaceColors must not be empty")
	}
}

func TestWorkspaceColors_ValidHex(t *testing.T) {
	for _, c := range WorkspaceColors {
		if len(c) != 7 || c[0] != '#' {
			t.Errorf("invalid color format: %q (want #RRGGBB)", c)
		}
	}
}

func TestWorkspace_LayoutMode_Default(t *testing.T) {
	ws := Workspace{
		ID:   uuid.New(),
		Name: "Test",
	}
	// Zero value of LayoutMode is "" — PaneCount should return 1 (single)
	if ws.LayoutMode.PaneCount() != 1 {
		t.Errorf("zero LayoutMode should behave as single (1 pane), got %d", ws.LayoutMode.PaneCount())
	}
}
