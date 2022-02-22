package chatcommand

import (
	"fmt"
	"strings"

	"awesome-dragon.science/go/irc/event"
	"awesome-dragon.science/go/irc/user"
)

// Argument represents an argument to a command
type Argument struct {
	CommandName string
	Arguments   []string

	Event  *event.Message
	Target string
	Reply  func(string)
}

// ArgString returns the arguments as a space joined string
func (a *Argument) ArgString() string {
	return strings.Join(a.Arguments, " ")
}

// Replyf is a thin wrapper around reply that allows for easy printf formatting of replies
func (a *Argument) Replyf(format string, args ...interface{}) {
	a.Reply(fmt.Sprintf(format, args...))
}

// SourceUser is a shortcut to event.SourceUser. It will be inlined
func (a *Argument) SourceUser() *user.EphemeralUser {
	return a.Event.SourceUser
}

// CurrentNick is a shortcut to event.CurrentNick. It will be inlined
func (a *Argument) CurrentNick() string {
	return a.Event.CurrentNick
}
