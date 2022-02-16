package chatcommand

import (
	"strings"

	"awesome-dragon.science/go/irc/client/event"
	"awesome-dragon.science/go/irc/user"
)

// Argument represents an argument to a command
type Argument struct {
	CommandName string
	Arguments   []string

	Event       *event.Message
	SourceUser  *user.EphemeralUser
	CurrentNick string
	Target      string
	Reply       func(string)
}

// ArgString returns the arguments as a space joined string
func (a *Argument) ArgString() string {
	return strings.Join(a.Arguments, " ")
}
