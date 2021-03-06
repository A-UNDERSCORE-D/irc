// Package irccommand provides an event.Handler implementation scoped to IRC commands
package irccommand

import (
	"sort"
	"strings"
	"sync"

	"awesome-dragon.science/go/irc/event"
)

var _ event.MessageHandler = (*Handler)(nil)

// Handler implements a IRC command based event system (and is an implementation of event.MessageHandler)
type Handler struct {
	mu sync.Mutex
	// command_name: id: func
	hooks  map[string]map[int]event.CallbackFunc
	lastID int
}

func keys(m map[int]event.CallbackFunc) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}

	sort.Ints(out)

	return out
}

func (h *Handler) collectCallbacks(name string, includeStar bool) []event.CallbackFunc {
	h.mu.Lock()
	defer h.mu.Unlock()

	out := []event.CallbackFunc{}

	for _, idx := range keys(h.hooks[name]) {
		out = append(out, h.hooks[name][idx])
	}

	if includeStar {
		for _, idx := range keys(h.hooks["*"]) {
			out = append(out, h.hooks[name][idx])
		}
	}

	return out
}

// OnMessage implements the MessageHandler interface
func (h *Handler) OnMessage(msg *event.Message) error {
	// this way callbacks can remove themselves or others
	toCall := h.collectCallbacks(strings.ToUpper(msg.Raw.Command), true)
	errors := []error{}

	for _, cb := range toCall {
		if err := cb(msg); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return &event.MultiError{Errors: errors}
	}

	return nil
}

// AddCallback adds a callback to the given command name. The returned ID can be used to remove the callback
// the special command name * will collect calls for any command. Similar to the raw handler
func (h *Handler) AddCallback(command string, callback event.CallbackFunc) int {
	h.mu.Lock()
	defer h.mu.Unlock()

	if h.hooks == nil {
		h.hooks = make(map[string]map[int]event.CallbackFunc)
	}

	command = strings.ToUpper(command)

	if h.hooks[command] == nil {
		h.hooks[command] = make(map[int]event.CallbackFunc)
	}

	h.lastID++

	h.hooks[command][h.lastID] = callback

	return h.lastID
}

// RemoveCallback remove a callback from the Handler instance
func (h *Handler) RemoveCallback(id int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, callbackMap := range h.hooks {
		delete(callbackMap, id) // there will never be a duplicated ID so this is fine
	}

	commandNames := make([]string, 0, len(h.hooks))
	for k := range h.hooks {
		commandNames = append(commandNames, k)
	}

	for _, k := range commandNames {
		if len(h.hooks[k]) == 0 {
			delete(h.hooks, k)
		}
	}
}

// WaitFor waits until the specified command is returned by the server
func (h *Handler) WaitFor(command string) <-chan *event.Message {
	outChan := make(chan *event.Message, 1)

	var id int
	id = h.AddCallback(command, func(m *event.Message) error {
		if !strings.EqualFold(m.Raw.Command, command) {
			return nil
		}

		h.RemoveCallback(id)
		outChan <- m
		close(outChan)

		return nil
	})

	return outChan
}
