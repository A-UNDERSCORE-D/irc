// Package isupport contains an implementation of an ISUPPORT message handler
package isupport

import (
	"strings"
	"sync"

	"awesome-dragon.science/go/irc/irc/mode"
	"github.com/ergochat/irc-go/ircmsg"
)

// ISupport represents ISUPPORT tokens from the server
// Note that this is protected by a mutex, and it is suggested that you lock it
// when querying, or make use of convenience methods
type ISupport struct {
	sync.Mutex
	Tokens       map[string]string
	ChannelModes []mode.ChannelMode
}

func New() *ISupport {
	return &ISupport{Tokens: make(map[string]string)}
}

// Parse parses an RPL_ISUPPORT line into its constituent tokens, and adds them
// to the ISupport struct
func (i *ISupport) Parse(msg *ircmsg.Message) {
	i.Lock()
	defer i.Unlock()

	if i.Tokens == nil {
		i.Tokens = make(map[string]string)
	}

	for idx := 1; idx < len(msg.Params)-1; idx++ {
		split := strings.SplitN(msg.Params[idx], "=", 2)
		arg := ""
		name := split[0]

		if len(split) > 1 {
			arg = split[1]
		}

		if name[0] == '-' {
			// token being removed
			delete(i.Tokens, name[1:])

			continue
		}

		i.Tokens[strings.ToLower(name)] = arg
	}

	if res, ok := i.Tokens["chanmodes"]; ok {
		i.ChannelModes = mode.ModesFromISupportToken(res)
	}
}
