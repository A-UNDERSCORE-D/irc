// Package chatmessage provides a simple frontend for message handlers
package chatmessage

import (
	"strings"
	"sync"

	"awesome-dragon.science/go/irc/event"
)

// MessageFunc is a callback func for a PRIVMSG
type MessageFunc func(message, target string, isPM bool, event *event.Message)

// Handler implements MessageHandler and powers this event system
type Handler struct {
	mu        sync.Mutex
	nextID    int
	callbacks map[int]MessageFunc
}

var _ event.MessageHandler = (*Handler)(nil)

// OnMessage implements a message handler for chat (PRIVMSG) messages
func (h *Handler) OnMessage(message *event.Message) error {
	if message.Raw.Command != "PRIVMSG" {
		return nil
	}

	isPM := message.CurrentNick == message.Raw.Params[0]

	h.mu.Lock()
	defer h.mu.Unlock()

	values := make([]MessageFunc, 0, len(h.callbacks))
	for _, v := range h.callbacks {
		values = append(values, v)
	}

	for _, v := range values {
		v(message.Raw.Params[1], message.Raw.Params[0], isPM, message)
	}

	return nil
}

// AddHandler adds a function to handle all incoming messages
func (h *Handler) AddHandler(f MessageFunc) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.setupIfNeeded()

	curID := h.nextID
	h.nextID++

	h.callbacks[curID] = f

	return curID
}

func isCTCP(m string) bool {
	return strings.HasPrefix(m, "\x01") && strings.HasSuffix(m, "\x01")
}

func (h *Handler) AddCTCPHandler(ctcp string, f MessageFunc) int {
	ctcp = strings.TrimSpace(ctcp)

	return h.AddHandler(func(message, target string, isPM bool, event *event.Message) {
		if !isCTCP(message) {
			return
		}

		message = strings.Trim(message, "\x01")
		cmd, _, _ := strings.Cut(message, " ")

		if ctcp != "" && !strings.EqualFold(cmd, ctcp) {
			return
		}

		f(message, target, isPM, event)
	})
}

// DelHandler removes a handler function, if it exists
func (h *Handler) DelHandler(id int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.setupIfNeeded()

	delete(h.callbacks, id)
}

func (h *Handler) setupIfNeeded() {
	if h.callbacks == nil {
		h.callbacks = make(map[int]MessageFunc)
	}
}
