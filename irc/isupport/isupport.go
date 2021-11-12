// Package isupport contains an implementation of an ISUPPORT message handler
package isupport

// spell-checker: words isupport
import (
	"strconv"
	"strings"
	"sync"

	"awesome-dragon.science/go/irc/irc/mode"
	"github.com/ergochat/irc-go/ircmsg"
)

// ISupport represents ISUPPORT tokens from the server
type ISupport struct {
	mu           sync.Mutex
	tokens       map[string]string
	channelModes mode.Set
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

	for _, param := range msg.Params[1 : len(msg.Params)-1] {
		split := strings.SplitN(param, "=", 2)
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

	// Extra work for modes
	if res, ok := i.tokens["chanmodes"]; ok {
		i.channelModes = mode.ModesFromISupportToken(res)
	}

	if prefix := i.prefix(true); prefix != nil {
		i.setupPrefixModes(prefix)
	}
}

func (i *ISupport) setupPrefixModes(prefixModes map[rune]rune) {
outer:
	for p, m := range prefixModes {
		for idx, real := range i.channelModes {
			// just in case
			if real.Char == m {
				i.channelModes[idx].Prefix = string(p)

				continue outer
			}
		}

		// wasn't found
		i.channelModes = append(i.channelModes, mode.Mode{Type: mode.TypeD, Char: m, Prefix: string(p)})
	}
}

// Modes returns the channel modes available on this ISupport struct. the returned data is a copy
func (i *ISupport) Modes() mode.Set {
	i.mu.Lock()
	defer i.mu.Unlock()

	return append(mode.Set(nil), i.channelModes...)
}

// fetches a token from the internal map, this does *NOT* interact with the mutex at all
func (i *ISupport) getTokenUnsafe(name string) (value string, exists bool) {
	res, ok := i.tokens[strings.ToLower(name)]

	return res, ok
}

// GetToken gets a token by name if it exists.
// You will likely want to use named methods for the token you want instead,
// assuming they exist.
func (i *ISupport) GetToken(name string) (value string, exists bool) {
	i.mu.Lock()
	defer i.mu.Unlock()

	return i.getTokenUnsafe(name)
}

// GetTokenDefault gets the given token, and returns a default if the *content*
// of the token is "". You should verify that the token at least existed before
// using this.
func (i *ISupport) GetTokenDefault(name, dflt string) string {
	res, ok := i.GetToken(name)
	if !ok {
		return ""
	}

	if res == "" {
		return dflt
	}

	return res
}

func (i *ISupport) getTokenDontCare(name string) string {
	res, _ := i.GetToken(name)

	return res
}

func (i *ISupport) listToken(name string) []string {
	return strings.Split(i.getTokenDontCare(name), "")
}

func (i *ISupport) listTokenDefault(name, dflt string) []string { //nolint:unused // but it might be!
	return strings.Split(i.GetTokenDefault(name, dflt), "")
}

// NumericToken returns the value of a token as a number, or -1 if it either
// doesn't exist, or fails to be parsed as a base 10 number
func (i *ISupport) NumericToken(name string) int {
	res, ok := i.GetToken(name)
	if !ok {
		return -1
	}

	num, err := strconv.Atoi(res)
	if err != nil {
		return -1
	}

	return num
}

// HasToken returns whether or not the token exists
func (i *ISupport) HasToken(name string) bool {
	_, ok := i.GetToken(name)

	return ok
}
