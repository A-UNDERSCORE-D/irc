package chatcommand

import (
	"fmt"

	"github.com/ergochat/irc-go/ircmsg"
)

// CommandPanicedError is returned from OnMessage if a command panics
type CommandPanicedError struct {
	CommandName string
	line        *ircmsg.Message
	PanicData   interface{}
}

func (e *CommandPanicedError) Error() string {
	l, _ := e.line.Line()

	return fmt.Sprintf("Command %q paniced on line %q with data %#v", e.CommandName, l, e.PanicData)
}
