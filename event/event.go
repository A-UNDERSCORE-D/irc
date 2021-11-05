// Package event defines a simple event system for use with irc
package event

// spell-checker: words ircmsg

import (
	"strings"
	"sync"

	"github.com/ergochat/irc-go/ircmsg"
)

// CallbackFunc is a function used in a callback (yes, really! this comment exists to make revive stop complaining)
type CallbackFunc func(msg *ircmsg.Message)

// Callback is a single event callback
type Callback struct {
	id         int
	concurrent bool
	callback   CallbackFunc
}

// Fire calls the callback, in a goroutine if requested
func (c *Callback) Fire(msg *ircmsg.Message) {
	if c.concurrent {
		go c.callback(msg)
	} else {
		c.callback(msg)
	}
}

// Manager implements a simple event system
type Manager struct {
	lastID int
	mu     sync.Mutex
	events map[string][]*Callback
}

// NewManager creates a new instance of Manager ready for use
func NewManager() *Manager {
	return &Manager{
		events: make(map[string][]*Callback),
	}
}

// GetEvent returns the event with the given ID, if it exists
func (m *Manager) GetEvent(cbID int) *Callback {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, events := range m.events {
		for _, e := range events {
			if e.id == cbID {
				return e
			}
		}
	}

	return nil
}

// Fire fires the given line across the event bus
func (m *Manager) Fire(name string, line *ircmsg.Message) {
	// Copy to allow modification from elsewhere
	m.mu.Lock()
	events := append([]*Callback(nil), m.events[strings.ToLower(name)]...) // <3 sandcat
	m.mu.Unlock()

	for _, event := range events {
		event.Fire(line)
	}
}

// AddCallback adds a callback to the event system
func (m *Manager) AddCallback(name string, callbackFunc CallbackFunc, concurrent bool) int {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastID++

	callback := &Callback{
		id:         m.lastID,
		callback:   callbackFunc,
		concurrent: concurrent,
	}

	loweredName := strings.ToLower(name)

	m.events[loweredName] = append(m.events[loweredName], callback)

	return callback.id
}

// RemoveCallback removes a callback from the event manager
func (m *Manager) RemoveCallback(cbID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for name, callbacks := range m.events {
		var newSlice []*Callback

		for _, callback := range callbacks {
			if callback.id == cbID {
				continue
			}

			newSlice = append(newSlice, callback)
		}

		m.events[name] = newSlice
	}
}

// AddOneShotCallback adds a callback that will be removed after its called once.
func (m *Manager) AddOneShotCallback(name string, callback CallbackFunc, concurrent bool) {
	id := 0
	wrapper := func(msg *ircmsg.Message) {
		callback(msg)
		m.RemoveCallback(id)
	}

	id = m.AddCallback(name, wrapper, concurrent)
}

// WaitFor returns a channel that will have a single ircmsg.Message sent over it
// it is *always* called synchronously internally. the returned channel contains space for one IRC message
func (m *Manager) WaitFor(name string) <-chan *ircmsg.Message {
	out := make(chan *ircmsg.Message, 1)

	m.AddOneShotCallback(name, func(msg *ircmsg.Message) {
		out <- msg
		close(out)
	}, false)

	return out
}
