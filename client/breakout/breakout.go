// Package breakout provides functions to split known IRC messages up into their "important" bits
//
// Note that these functions are generally NOT tag aware
package breakout

import (
	"awesome-dragon.science/go/irc/client/event"
	"awesome-dragon.science/go/irc/user"
)

// Join breaks up an IRC JOIN message into its parts.
func Join(msg *event.Message) (source *user.EphemeralUser, channel string) {
	return nil, ""
}
