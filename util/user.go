package util

import (
	"strings"

	"github.com/ergochat/irc-go/ircutils"
)

// UserHostCanonical returns the mask form (nick!user@host) of a given UserHost
func UserHostCanonical(u ircutils.UserHost) string {
	out := strings.Builder{}
	out.Grow(len(u.Nick))
	out.WriteString(u.Nick)

	if u.User != "" {
		out.WriteRune('!')
		out.WriteString(u.User)
	}

	if u.Host != "" {
		out.WriteRune('@')
		out.WriteString(u.Host)
	}

	return out.String()
}
