package git

import (
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const debounceDelay = 500 * time.Millisecond

// Watcher monitors a directory for file changes and calls OnChange after debounce.
type Watcher struct {
	path     string
	watcher  *fsnotify.Watcher
	OnChange func()

	mu      sync.Mutex
	timer   *time.Timer
	stopped bool
}

// NewWatcher creates and starts a file watcher for path.
func NewWatcher(path string, onChange func()) (*Watcher, error) {
	fw, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	if err := fw.Add(path); err != nil {
		fw.Close()
		return nil, err
	}

	w := &Watcher{
		path:     path,
		watcher:  fw,
		OnChange: onChange,
	}

	go w.run()
	return w, nil
}

func (w *Watcher) run() {
	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return
			}
			_ = event
			w.schedule()
		case err, ok := <-w.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("file watcher error for %s: %v", w.path, err)
		}
	}
}

func (w *Watcher) schedule() {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.stopped {
		return
	}
	if w.timer != nil {
		w.timer.Reset(debounceDelay)
		return
	}
	w.timer = time.AfterFunc(debounceDelay, func() {
		w.mu.Lock()
		w.timer = nil
		w.mu.Unlock()
		if w.OnChange != nil {
			w.OnChange()
		}
	})
}

// Stop shuts down the watcher.
func (w *Watcher) Stop() {
	w.mu.Lock()
	w.stopped = true
	if w.timer != nil {
		w.timer.Stop()
		w.timer = nil
	}
	w.mu.Unlock()
	_ = w.watcher.Close()
}
