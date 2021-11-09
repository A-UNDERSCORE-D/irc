package isupport

import (
	"strconv"
	"strings"
)

// This file contains all the named getters for various ISUPPORT tokens
// The following methods are direct methods to access "standard" ISUPPORT tokens.
// They are in alphabetical order of the tokens, and are in the same order as
// the definitions shown at https://modern.ircdocs.horse/#rplisupport-parameters

// MaxAwayLen returns the maximum size of an away message, or -1 if unlimited or unset
func (i *ISupport) MaxAwayLen() int { return i.NumericToken("AWAYLEN") }

// CaseMapping returns the casemapping of the server, or "" if unset
func (i *ISupport) CaseMapping() string { return i.getTokenDontCare("CASEMAPPING") }

// ChanLimit returns the maximum number of channels a client may be in, per channel type
// https://modern.ircdocs.horse/#chanlimit-parameter
func (i *ISupport) ChanLimit() map[string]int {
	res, ok := i.GetToken("CHANLIMIT")
	if !ok {
		return nil
	}

	out := make(map[string]int)

	limits := strings.Split(res, ",")
	for _, v := range limits {
		split := strings.Split(v, ":")
		chanTypes := split[0]
		limit := -1

		if len(split) > 1 {
			l, err := strconv.Atoi(split[1])
			if err != nil {
				limit = -1
			} else {
				limit = l
			}
		}

		for _, t := range chanTypes {
			out[string(t)] = limit
		}
	}

	return out
}

// ChanModes Returns the raw CHANMODES isupport param. You likely want Modes
// https://modern.ircdocs.horse/#chanmodes-parameter
func (i *ISupport) ChanModes() string { return i.getTokenDontCare("CHANMODES") }

// MaxChanLen returns the maximum length of a channel the client may join
// https://modern.ircdocs.horse/#channellen-parameter
func (i *ISupport) MaxChanLen() int { return i.NumericToken("CHANNELLEN") }

// ChanTypes returns the list of channel types supported by the network
// https://modern.ircdocs.horse/#chantypes-parameter
func (i *ISupport) ChanTypes() []string { return i.listToken("CHANTYPES") }

// EList returns the supported extensions to the LIST command
// https://modern.ircdocs.horse/#elist-parameter
func (i *ISupport) EList() []string { return i.listToken("ELIST") }

// Excepts returns the mode to set ban exemptions, assuming the server supports it.
func (i *ISupport) Excepts() []string { return i.listTokenDefault("EXCEPTS", "e") }

// Extban returns the prefix and available extban characters
// https://modern.ircdocs.horse/#rplisupport-parameters
func (i *ISupport) Extban() (prefix string, types []string) {
	res, ok := i.GetToken("EXTBAN")
	if !ok {
		return "", nil
	}

	split := strings.Split(res, ",")
	prefix = split[0]
	types = strings.Split(split[1], "")

	return prefix, types
}

// MaxHostLen returns the maximum hostname length the server SHOULD use
// Its possible this will be outright ignored
// https://modern.ircdocs.horse/#hostlen-parameter
func (i *ISupport) MaxHostLen() int { return i.NumericToken("HOSTLEN") }

// InviteExemption returns the invite exemption mode, if supported
// https://modern.ircdocs.horse/#invex-parameter
func (i *ISupport) InviteExemption() string { return i.GetTokenDefault("INVEX", "I") }

// MaxKickLen returns the maximum length a kick message should have when sent by the client
// https://modern.ircdocs.horse/#kicklen-parameter
func (i *ISupport) MaxKickLen() int { return i.NumericToken("KICKLEN") }

// MaxListModes returns an array of mode char -> int, mapping the maximum
// https://modern.ircdocs.horse/#maxlist-parameter
func (i *ISupport) MaxListModes() map[rune]int {
	res, ok := i.GetToken("MAXLIST")
	if !ok {
		return nil
	}

	out := make(map[rune]int)

	for _, pair := range strings.Split(res, ",") {
		split := strings.Split(pair, ":")
		types := split[0]
		max := -1

		if len(split) > 1 {
			num, err := strconv.Atoi(split[1])
			if err != nil {
				max = num
			}
		}

		for _, s := range types {
			out[s] = max
		}
	}

	return out
}

// MaxTargets returns the maximum targets for all commands. i.MaxCommandTargets should be preferred to this if
// available
// https://modern.ircdocs.horse/#maxtargets-parameter
func (i *ISupport) MaxTargets() int { return i.NumericToken("MAXTARGETS") }

// MaxModes returns the maximum number of type A, B, and C modes that should
// be sent in a single MODE command.
// -1 is returned when either the token wasn't present, or the number is unlimited
// https://modern.ircdocs.horse/#modes-parameter
func (i *ISupport) MaxModes() int { return i.NumericToken("MODES") }

// Network returns the content of the NETWORK token, if it exists
// https://modern.ircdocs.horse/#network-parameter
func (i *ISupport) Network() string { return i.getTokenDontCare("NETWORK") }

// MaxNickLen returns the maximum nick length the client should send to the server
// https://modern.ircdocs.horse/#nicklen-parameter
func (i *ISupport) MaxNickLen() int { return i.NumericToken("NICKLEN") }

// Prefix returns a map of modeChar -> prefix
// https://modern.ircdocs.horse/#prefix-parameter
func (i *ISupport) Prefix() map[rune]rune {
	res, exists := i.GetToken("PREFIX")
	if !exists {
		return nil
	}

	out := make(map[rune]rune)

	split := strings.Split(res[1:], ")")
	// Just to be sure
	if len(split) < 3 || len(split[0]) != len(split[1]) {
		return nil
	}

	for idx, m := range split[0] {
		out[m] = rune(split[1][idx])
	}

	return out
}

// SafeList returns whether or not LIST usage promises to not RECVQ (disconnect
// due to large buffer of sent data server side)
// https://modern.ircdocs.horse/#safelist-parameter
func (i *ISupport) SafeList() bool { return i.HasToken("SAFELIST") }

// SilenceMax returns the maximum number of entries on the SILENCE list.
// a return of -1 indicates silence is *NOT* supported
// https://modern.ircdocs.horse/#silence-parameter
func (i *ISupport) SilenceMax() int { return i.NumericToken("SILENCE") }

// StatusMsg returns the different prefixes for status messages, or nil
// https://modern.ircdocs.horse/#statusmsg-parameter
func (i *ISupport) StatusMsg() []string { return i.listToken("STATUSMSG") }

// MaxCommandTargets returns a map of commandname (ToLower()'d) to int, where
// the int is the maximum number of targets allowed, -1 meaning no limit
// https://modern.ircdocs.horse/#targmax-parameter
func (i *ISupport) MaxCommandTargets() map[string]int {
	params, exists := i.GetToken("TARGMAX")
	if !exists {
		return nil
	}

	out := map[string]int{}

	for _, pair := range strings.Split(params, ",") {
		split := strings.Split(pair, ":")
		num := -1

		if len(split) > 1 {
			realNum, err := strconv.Atoi(split[1])
			if err != nil {
				num = realNum
			}
		}

		out[strings.ToLower(split[0])] = num
	}

	return out
}

// MaxTopicLen returns the maximum topic size the client should send to the server
// https://modern.ircdocs.horse/#topiclen-parameter
func (i *ISupport) MaxTopicLen() int { return i.NumericToken("TOPICLEN") }
