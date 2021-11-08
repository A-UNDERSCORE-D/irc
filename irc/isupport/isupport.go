// Package isupport contains an implementation of an ISUPPORT message handler
package isupport

// spell-checker: words isupport
import (
	"strings"
	"sync"

	"awesome-dragon.science/go/irc/irc/mode"
	"github.com/ergochat/irc-go/ircmsg"
)

// ISupport represents ISUPPORT tokens from the server
type ISupport struct {
	mu           sync.Mutex
	tokens       map[string]string
	channelModes mode.ModeSet
}

// New creates a new instance of ISupport ready for use
func New() *ISupport {
	return &ISupport{tokens: make(map[string]string)}
}

// Parse parses an RPL_ISUPPORT line into its constituent tokens, and adds them
// to the ISupport struct
func (i *ISupport) Parse(msg *ircmsg.Message) {
	i.mu.Lock()
	defer i.mu.Unlock()

	if i.tokens == nil {
		i.tokens = make(map[string]string)
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
			delete(i.tokens, name[1:])

			continue
		}

		i.tokens[strings.ToLower(name)] = arg
	}

	if res, ok := i.tokens["chanmodes"]; ok {
		i.channelModes = mode.ModesFromISupportToken(res)
	}
}

// Modes returns the channel modes available on this ISupport struct. the returned data is a copy
func (i *ISupport) Modes() mode.ModeSet {
	i.mu.Lock()
	defer i.mu.Unlock()

	return append(mode.ModeSet(nil), i.channelModes...)
}
