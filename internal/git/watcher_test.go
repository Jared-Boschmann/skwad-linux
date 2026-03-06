package git

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestWatcher_FiresOnChange(t *testing.T) {
	dir := t.TempDir()

	var (
		mu    sync.Mutex
		fired bool
	)

	w, err := NewWatcher(dir, func() {
		mu.Lock()
		fired = true
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	defer w.Stop()

	// Write a file to trigger the watcher.
	if err := os.WriteFile(filepath.Join(dir, "trigger.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}

	deadline := time.After(2 * time.Second)
	for {
		select {
		case <-deadline:
			t.Fatal("watcher did not fire within 2s")
		case <-time.After(50 * time.Millisecond):
			mu.Lock()
			done := fired
			mu.Unlock()
			if done {
				return
			}
		}
	}
}

func TestWatcher_Debounce(t *testing.T) {
	dir := t.TempDir()

	var (
		mu    sync.Mutex
		count int
	)

	w, err := NewWatcher(dir, func() {
		mu.Lock()
		count++
		mu.Unlock()
	})
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	defer w.Stop()

	// Write several files in quick succession.
	for i := 0; i < 5; i++ {
		name := filepath.Join(dir, "file"+string(rune('a'+i))+".txt")
		_ = os.WriteFile(name, []byte("x"), 0o644)
	}

	// Wait longer than the debounce window.
	time.Sleep(debounceDelay + 300*time.Millisecond)

	mu.Lock()
	got := count
	mu.Unlock()

	// With debounce, multiple rapid writes should produce only 1-2 callbacks.
	if got > 3 {
		t.Errorf("expected debounced callbacks (≤3), got %d", got)
	}
	if got == 0 {
		t.Error("expected at least one callback")
	}
}

func TestWatcher_Stop(t *testing.T) {
	dir := t.TempDir()
	w, err := NewWatcher(dir, func() {})
	if err != nil {
		t.Fatalf("NewWatcher: %v", err)
	}
	// Should not panic on double stop.
	w.Stop()
	w.Stop()
}
