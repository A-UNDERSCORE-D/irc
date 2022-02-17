package event

import (
	"awesome-dragon.science/go/irc/capab"
	"awesome-dragon.science/go/irc/user"
	"github.com/ergochat/irc-go/ircmsg"
)

// Message represents an incoming IRC message and associated metadata
type Message struct {
	Raw           *ircmsg.Message
	SourceUser    *user.EphemeralUser
	CurrentNick   string
	AvailableCaps []capab.Capability
}

// MessageHandler represents anything that can deal with an IRC message
type MessageHandler interface {
	OnMessage(message *Message) error
}

// CallbackFunc is a base name for a function that could be used as a callback.
// implementations of MessageHandler dont have to use this internally, but
// some of the base ones do so its helpful.
type CallbackFunc func(*Message) error
