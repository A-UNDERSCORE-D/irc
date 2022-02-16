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

	if !n.config.SASL {
		return nil // Not supposed to, dont return an error
	}

	if saslCap == nil || !saslCap.Acknowledged {
		// either we're not supposed to, or its not available
		return ErrSASLNotSupported
	}

	mech := ""
	if n.config.SASLUsername != "" && n.config.SASLPassword != "" {
		mech = "PLAIN"
	}

	supportedMechs := strings.Split(saslCap.Value, ",")

	if !util.StringSliceContains(mech, supportedMechs) {
		return ErrSASLMechNotSupported
	}

	switch mech {
	case "PLAIN":
		c := make(chan error)

		go func() { c <- n.doSaslPLAIN(n.config.SASLUsername, n.config.SASLPassword) }()

		return <-c

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

	var authenticateID, authGoodID, authBadID int

	authenticateID = n.eventManager.AddCallback(
		"AUTHENTICATE",
		func(msg *ircmsg.Message) error {
			authChan <- msg.Params[len(msg.Params)-1]
			n.eventManager.RemoveCallback(authenticateID)

			return nil
		},
	)

	authGoodID = n.eventManager.AddCallback(
		numerics.RPL_SASLSUCCESS,
		func(*ircmsg.Message) error {
			authChan <- "GOOD"
			n.eventManager.RemoveCallback(authGoodID)

			return nil
		},
	)

	authBadID = n.eventManager.AddCallback(
		numerics.ERR_SASLFAIL,
		func(*ircmsg.Message) error {
			authChan <- "BAD"
			n.eventManager.RemoveCallback(authBadID)

			return nil
		},
	)

	defer n.eventManager.RemoveCallback(authGoodID)
	defer n.eventManager.RemoveCallback(authBadID)

	_ = n.writeIRC("AUTHENTICATE", "PLAIN")

	if res := <-authChan; res != "+" {
		return fmt.Errorf("server returned unexpected data %q: %w", res, ErrSASLFailed)
	}

	_ = n.writeIRC("AUTHENTICATE", makePlainAuth(username, password))

	switch res := <-authChan; res {
	case "GOOD":
		return nil
	case "BAD":
		return ErrSASLFailed

	default:
		return ErrSASLFailed
	}
}
