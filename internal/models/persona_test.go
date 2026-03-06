package models

import (
	"testing"

	"github.com/google/uuid"
)

// Fixed UUIDs must never change — agents on different machines identify personas by ID.
func TestPersona_FixedUUIDs(t *testing.T) {
	cases := []struct {
		name string
		id   uuid.UUID
		want string
	}{
		{"KentBeck", PersonaIDKentBeck, "a1000001-0000-0000-0000-000000000001"},
		{"MartinFowler", PersonaIDMartinFowler, "a1000001-0000-0000-0000-000000000002"},
		{"LinusTorvalds", PersonaIDLinusTorvalds, "a1000001-0000-0000-0000-000000000003"},
		{"UncleBob", PersonaIDUncleBob, "a1000001-0000-0000-0000-000000000004"},
		{"JohnCarmack", PersonaIDJohnCarmack, "a1000001-0000-0000-0000-000000000005"},
		{"DaveFarley", PersonaIDDaveFarley, "a1000001-0000-0000-0000-000000000006"},
	}
	for _, tc := range cases {
		if tc.id.String() != tc.want {
			t.Errorf("%s: got UUID %s, want %s", tc.name, tc.id.String(), tc.want)
		}
	}
}

func TestDefaultPersonas_Count(t *testing.T) {
	personas := DefaultPersonas()
	if len(personas) != 6 {
		t.Errorf("expected 6 default personas, got %d", len(personas))
	}
}

func TestDefaultPersonas_AllSystem(t *testing.T) {
	for _, p := range DefaultPersonas() {
		if p.Type != PersonaTypeSystem {
			t.Errorf("persona %q should be type system, got %s", p.Name, p.Type)
		}
		if p.State != PersonaStateEnabled {
			t.Errorf("persona %q should be enabled by default, got %s", p.Name, p.State)
		}
	}
}

func TestDefaultPersonas_NoDuplicateIDs(t *testing.T) {
	seen := make(map[uuid.UUID]string)
	for _, p := range DefaultPersonas() {
		if prev, ok := seen[p.ID]; ok {
			t.Errorf("duplicate UUID %s shared by %q and %q", p.ID, prev, p.Name)
		}
		seen[p.ID] = p.Name
	}
}

func TestDefaultPersonas_NonEmptyInstructions(t *testing.T) {
	for _, p := range DefaultPersonas() {
		if p.Instructions == "" {
			t.Errorf("persona %q has empty instructions", p.Name)
		}
	}
}
