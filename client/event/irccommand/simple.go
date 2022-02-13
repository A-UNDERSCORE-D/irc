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
func (c *SimpleHandler) AddCallback(command string, callback func(*ircmsg.Message) error) int {
	return c.Handler.AddCallback(command, func(m *event.Message) error {
		return callback(m.Raw)
	})
}
