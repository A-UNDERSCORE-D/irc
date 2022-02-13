package servernotice

import (
	"strings"
	"sync"

	"awesome-dragon.science/go/irc/client/event"
)

var _ event.MessageHandler = &Handler{}

// Handler implements a simple IRC
type Handler struct {
	mu        sync.Mutex
	callbacks map[int]event.CallbackFunc
	lastID    int
}

// OnMessage implements event.MessageHandler
func (h *Handler) OnMessage(msg *event.Message) error {
	if msg.Raw.Command != "NOTICE" {
		return nil
	}

	if !strings.Contains(msg.SourceUser.Nick, ".") {
		// could be a user or otherwise
		return nil
	}

	return nil
}

// RegisterCallback registers a callback to be used when a server notice is received from the server
func (h *Handler) RegisterCallback(f event.CallbackFunc) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.callbacks == nil {
		h.callbacks = make(map[int]event.CallbackFunc)
	}

	id := h.lastID
	h.lastID++

	h.callbacks[id] = f

	return id
}
