package irccommand

import (
	"awesome-dragon.science/go/irc/client/event"
	"github.com/ergochat/irc-go/ircmsg"
)

// SimpleHandler is a wrapper around Handler that works with the raw ircmsg.Message directly
type SimpleHandler struct {
	*Handler
}

// AddCallback adds a callback to the SimpleHandler instance
func (h *SimpleHandler) AddCallback(command string, callback func(*ircmsg.Message) error) int {
	return h.Handler.AddCallback(command, func(m *event.Message) error {
		return callback(m.Raw)
	})
}

// WaitFor waits for the specified IRC command, and sends the
func (h *SimpleHandler) WaitFor(command string) <-chan *ircmsg.Message {
	outChan := make(chan *ircmsg.Message)
	c := h.Handler.WaitFor(command)

	go func() {
		outChan <- (<-c).Raw
		close(outChan)
	}()

	return outChan
}
