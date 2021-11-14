package capab

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"awesome-dragon.science/go/irc/numerics"
	"awesome-dragon.science/go/irc/util"
	"github.com/ergochat/irc-go/ircmsg"
)

func (n *Negotiator) doSasl() error {
	saslCap := n.capByName("sasl")

	if !n.config.SASL || saslCap == nil || !saslCap.acknowledged {
		// either we're not supposed to, or its not available
		return ErrSASLNotSupported
	}

	mech := ""
	if n.config.SASLUsername != "" && n.config.SASLPassword != "" {
		mech = "PLAIN"
	}

	supportedMechs := strings.Split(saslCap.value, ",")

	if !util.StringSliceContains(mech, supportedMechs) {
		return ErrSASLMechNotSupported
	}

	switch mech {
	case "PLAIN":
		return n.doSaslPLAIN(n.config.SASLUsername, n.config.SASLPassword)

	default:
		return ErrSASLMechNotSupported
	}
}

func makePlainAuth(username, password string) string {
	toEncode := fmt.Sprintf("\x00%s\x00%s", username, password)

	return base64.RawStdEncoding.EncodeToString([]byte(toEncode))
}

// Various SASL error messages
var (
	ErrSASLFailed           = errors.New("SASL Failed")
	ErrSASLNotSupported     = errors.New("SASL not supported by server")
	ErrSASLMechNotSupported = errors.New("SASL mechanism not supported by server")
)

func (n *Negotiator) doSaslPLAIN(username, password string) error {
	if username == "" || password == "" {
		return fmt.Errorf("cannot authenticate with empty username or password: %w", ErrSASLFailed)
	}

	authChan := make(chan string)

	n.client.AddOneShotCallback(
		"AUTHENTICATE",
		func(msg *ircmsg.Message) { authChan <- msg.Params[len(msg.Params)-1] },
		true,
	)

	authGood := n.client.AddOneShotCallback(
		numerics.RPL_SASLSUCCESS, func(*ircmsg.Message) { authChan <- "GOOD" }, true,
	)
	authBad := n.client.AddOneShotCallback(numerics.ERR_SASLFAIL, func(*ircmsg.Message) { authChan <- "BAD" }, true)

	defer n.client.RemoveCallback(authGood)
	defer n.client.RemoveCallback(authBad)

	_ = n.client.Write("AUTHENTICATE", "PLAIN")

	if res := <-authChan; res != "+" {
		return fmt.Errorf("server returned unexpected data %q: %w", res, ErrSASLFailed)
	}

	_ = n.client.Write("AUTHENTICATE", makePlainAuth(username, password))

	switch res := <-authChan; res {
	case "GOOD":
		return nil
	case "BAD":
		return ErrSASLFailed

	default:
		return ErrSASLFailed
	}
}
