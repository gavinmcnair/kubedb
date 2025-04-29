package watch

import (
	"sync"
)

// WatchManager manages watch registrations and notifications for key changes.
type WatchManager struct {
	watchers map[string][]chan []byte
	mu       sync.RWMutex 
}

// NewWatchManager initializes a new watch manager.
func NewWatchManager() *WatchManager {
	return &WatchManager{
		watchers: make(map[string][]chan []byte),
	}
}

// Watch registers a new watcher for a specific key, returning the channel.
// Prefixed watches have to be handled like Watch("prefix/")
func (wm *WatchManager) Watch(key string) chan []byte {
	ch := make(chan []byte)
	wm.mu.Lock()
	wm.watchers[key] = append(wm.watchers[key], ch)
	wm.mu.Unlock()
	return ch
}

// CancelWatch deregisters a watcher channel for a specific key.
func (wm *WatchManager) CancelWatch(key string, ch chan []byte) {
	wm.mu.Lock()
	defer wm.mu.Unlock()

	if chans, exists := wm.watchers[key]; exists {
		for i, c := range chans {
			if c == ch {
				close(ch)
				chans = append(chans[:i], chans[i+1:]...)
				break
			}
		}
		if len(chans) == 0 {
			delete(wm.watchers, key)
		} else {
			wm.watchers[key] = chans
		}
	}
}

// Notify sends notifications to all registered watchers for a specific key.
// If the key is watched, watchers will be sending the new value.
func (wm *WatchManager) Notify(key string, value []byte) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()

	if chans, exists := wm.watchers[key]; exists {
		for _, ch := range chans {
			// Non-blocking send to avoid blocking the loop
			select {
			case ch <- value:
			default:
			}
		}
	}
}

