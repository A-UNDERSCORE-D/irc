// Package user provides some simple IRC user tracking
package user

import (
	"net"
	"strings"

	"awesome-dragon.science/go/irc/capab"
	"github.com/ergochat/irc-go/ircmsg"
)

func capAvailable(key string, s []capab.Capability) bool {
	for _, v := range s {
		if strings.EqualFold(v.Name, key) {
			return true
		}
	}

	return false
}

// User represents an IRC user, with some optional bits of info, if known
type User struct {
	ircmsg.NUH
	RealIP   net.IP
	RealHost string
	RealName string
	Account  string
}

// Mask returns a n!u@h mask for the given User instance
func (u *User) Mask() string {
	return u.NUH.Canonical()
}

// EphemeralUser represents an IRC user. It is intended ephemeral use on messages.
// as in, it should be created from a message and optionally augmented with
// stored data. it should *NOT* be used as a normal user store, as some things
// that are stored on it may change at any time server side (such as oper status)
//
// There is no promise that anything other than nick/user/host exists.
type EphemeralUser struct {
	User
	AuthedForNick bool
	Oper          bool
	OperName      string
}

// FromMessage creates an EphemeralUser instance from a message.
// It will make use of various tags, if offered, to add more information to
// the struct
func FromMessage(msg *ircmsg.Message, availableCaps []capab.Capability) *EphemeralUser {
	nuh, _ := ircmsg.ParseNUH(msg.Source)
	out := &EphemeralUser{
		User: User{NUH: nuh},
	}

	for tagname, value := range msg.AllTags() {
		switch tagname {
		case "account-tag":
			out.Account = value

		case "solanum.chat/identified":
			out.AuthedForNick = true

		case "solanum.chat/oper":
			out.Oper = true
			out.OperName = value

		case "solanum.chat/realhost":
			out.RealHost = value

		case "solanum.chat/ip":
			out.RealIP = net.ParseIP(value)
		}
	}

	switch { //nolint:gocritic // It will have others eventually
	case msg.Command == "JOIN" && capAvailable("extended-join", availableCaps):
		// ASSUMING this is extended-join
		out.RealName = msg.Params[len(msg.Params)-1]
	}

	return out
}
