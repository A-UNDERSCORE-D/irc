package irccommand

import (
	"awesome-dragon.science/go/irc/client/event"
	"github.com/ergochat/irc-go/ircmsg"
)

// SimpleHandler is a wrapper around Handler that works with the raw ircmsg.Message directly.
type SimpleHandler struct {
	*Handler
}

func (h *SimpleHandler) init() {
	if h.Handler == nil {
		h.Handler = new(Handler)
	}
}

// OnMessage calls Handler.OnMessage
func (h *SimpleHandler) OnMessage(msg *event.Message) error {
	h.init()

	return h.Handler.OnMessage(msg)
}

// AddCallback adds a callback to the SimpleHandler instance
func (h *SimpleHandler) AddCallback(command string, callback func(*ircmsg.Message) error) int {
	h.init()

	return h.Handler.AddCallback(command, func(m *event.Message) error {
		return callback(m.Raw)
	})
}

// RemoveCallback calls Handler.RemoveCallback
func (h *SimpleHandler) RemoveCallback(id int) {
	h.init()
	h.Handler.RemoveCallback(id)
}

// WaitFor waits for the specified IRC command, and sends the
func (h *SimpleHandler) WaitFor(command string) <-chan *ircmsg.Message {
	h.init()

	outChan := make(chan *ircmsg.Message)
	c := h.Handler.WaitFor(command)

	go func() {
		outChan <- (<-c).Raw
		close(outChan)
	}()

	return outChan
}
