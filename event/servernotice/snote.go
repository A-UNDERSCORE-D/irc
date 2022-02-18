// Package servernotice implements event.Handler for IRC server notices (not regular NOTICEs)
package servernotice

import (
	"strings"
	"sync"

	"awesome-dragon.science/go/irc/event"
)

var _ event.MessageHandler = &Handler{}

// SnoteData provides easy access to the message in a server notice
type SnoteData struct {
	Event   *event.Message
	Message string
}

// Callback is the func type that servernotice.Handler
type Callback func(*SnoteData) error

// Handler implements a simple event.Handler that
type Handler struct {
	mu        sync.Mutex
	callbacks map[int]Callback
	lastID    int
}

// OnMessage implements event.MessageHandler
func (h *Handler) OnMessage(msg *event.Message) error {
	if msg.Raw.Command != "NOTICE" {
		return nil
	}

	if !strings.Contains(msg.SourceUser.Name, ".") {
		// could be a user or otherwise
		return nil
	}

	if msg.Raw.Params[0] != "*" { // solanum uses NOTICE *
		return nil
	}

	toSend := &SnoteData{
		Event:   msg,
		Message: msg.Raw.Params[len(msg.Raw.Params)-1],
	}

	outErrs := []error{}

	for _, f := range h.collectCallbacks() {
		if err := f(toSend); err != nil {
			outErrs = append(outErrs, err)
		}
	}

	if len(outErrs) > 0 {
		return &event.MultiError{Errors: outErrs}
	}

	return nil
}

func (h *Handler) collectCallbacks() []Callback {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]Callback, 0, len(h.callbacks))

	for _, v := range h.callbacks {
		out = append(out, v)
	}

	return out
}

// RegisterCallback registers a callback to be used when a server notice is received from the server
func (h *Handler) RegisterCallback(f Callback) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.callbacks == nil {
		h.callbacks = make(map[int]Callback)
	}

	id := h.lastID
	h.lastID++

	h.callbacks[id] = f

	return id
}
