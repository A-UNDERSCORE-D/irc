package capab

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"

	"awesome-dragon.science/go/irc/irc/numerics"
	"awesome-dragon.science/go/irc/irc/util"
	"github.com/ergochat/irc-go/ircmsg"
)

// minClient is the smallest interface required to make Negotiator work
type minClient interface {
	AddCallback(string, func(*ircmsg.Message), bool) int
	RemoveCallback(int)
	AddOneShotCallback(string, func(*ircmsg.Message), bool) int

	Write(string, ...string) error
}

// Config contains the configuration options for the Negotiator struct
type Config struct {
	ToRequest []string
	SASL      bool

	SASLUsername string
	SASLPassword string
	SASLMech     string

	// TODO: keys
}

type capability struct {
	name         string
	value        string
	available    bool
	request      bool
	acknowledged bool
}

func (c *capability) String() string {
	return c.name
}

// Negotiator negotiates IRCv3 capabilities over a Client instance
type Negotiator struct {
	mu sync.Mutex

	client minClient
	config *Config

	capabilities []*capability
	incomingCaps []string

	doingNegotiation bool
	reqsSent         int
}

// New creates a new Negotiator instance
func New(conf *Config, client minClient) *Negotiator {
	out := &Negotiator{config: conf}
	for _, c := range conf.ToRequest {
		out.capabilities = append(out.capabilities, &capability{name: c, request: true})
	}

	out.client = client

	return out
}

// Negotiate negotiates IRCv3 capabilities with a server, and optionally performs
// sasl authentication
func (n *Negotiator) Negotiate() {
	n.doNegotiation()

	if err := n.doSasl(); err != nil {
		log.Printf("Failed SASL: %s", err)
	}

	// Add NEW/DEL
	n.client.AddCallback("CAP", func(msg *ircmsg.Message) {
		split := strings.Split(msg.Params[len(msg.Params)-1], " ")

		switch cmd := msg.Params[1]; cmd {
		case "NEW":
			n.onCapNEW(split)
		case "DEL":
			n.onCapDEL(split)
		}
	}, true)

	_ = n.client.Write("CAP", "END")
}

func (n *Negotiator) doNegotiation() {
	msgChan := make(chan *ircmsg.Message)

	capCallback := n.client.AddCallback("CAP", func(msg *ircmsg.Message) { msgChan <- msg }, false)
	welcomeCallback := n.client.AddCallback(numerics.RPL_WELCOME, nil, false)

	defer n.client.RemoveCallback(capCallback)
	defer n.client.RemoveCallback(welcomeCallback)
	n.doingNegotiation = true
	_ = n.client.Write("CAP", "LS", "302")

	for n.doingNegotiation {
		msg := <-msgChan

		if msg.Command == numerics.RPL_WELCOME {
			log.Printf("Got unexpected 001. Assuming the server does not support capabilities")

			break
		}

		split := strings.Split(msg.Params[len(msg.Params)-1], " ")
		moreComing := false

		if len(msg.Params) >= 3 {
			moreComing = msg.Params[2] == "*"
		}

		switch cmd := msg.Params[1]; cmd {
		case "LS":
			n.onCapLS(split, moreComing)
		case "ACK":
			n.onCapACK(split)
		case "NAK":
			n.onCapNAK(split)
		case "DEL":
			n.onCapDEL(split)
		case "NEW":
			n.onCapNEW(split)
		default:
			log.Printf("Unknown CAP command %q. ignoring", cmd)
		}
	}
}

func (n *Negotiator) onCapLS(caps []string, moreComing bool) {
	n.incomingCaps = append(n.incomingCaps, caps...)

	if moreComing {
		return
	}

	// No more coming
	log.Printf("Server offered caps %v", n.incomingCaps)
	n.parseCaps()
	n.incomingCaps = nil // clear this for use in ACK later
	n.requestCaps()
}

func (n *Negotiator) requestCaps() {
	var toRequest []*capability

	for _, c := range n.capabilities {
		if c.request && c.available {
			toRequest = append(toRequest, c)
		}
	}

	var (
		lines   []string
		builder strings.Builder
	)

	for _, c := range toRequest {
		if builder.Len()+len(c.name) >= 450 {
			lines = append(lines, strings.TrimSpace(builder.String()))
			builder.Reset()
		}

		builder.WriteString(c.name)
		builder.WriteRune(' ')
	}

	lines = append(lines, strings.TrimSpace(builder.String()))

	log.Printf("Requesting capabilities %v", toRequest)

	n.reqsSent += len(lines)

	for _, l := range lines {
		_ = n.client.Write("CAP", "REQ", l)
	}
}

func (n *Negotiator) parseCaps() {
	for _, capab := range n.incomingCaps {
		name := capab
		value := ""

		if strings.Contains(name, "=") {
			split := strings.SplitN(capab, "=", 2)
			name = split[0]
			value = split[1]
		}

		if c := n.capByName(name); c != nil {
			c.available = true
			c.value = value

			continue
		}

		n.capabilities = append(n.capabilities, &capability{
			name:         name,
			value:        value,
			available:    true,
			request:      false,
			acknowledged: false,
		})
	}
}

func (n *Negotiator) capByName(name string) *capability {
	for _, capab := range n.capabilities {
		if capab.name == name {
			return capab
		}
	}

	return nil
}

func (n *Negotiator) onCapACK(caps []string) {
	// process ack into things and stuff
	n.incomingCaps = append(n.incomingCaps, caps...)
	n.reqsSent--
	// more coming?
	if n.reqsSent > 0 {
		return
	}

	// no more coming
	ackedCaps := make([]*capability, 0, len(n.incomingCaps))

	for _, cName := range n.incomingCaps {
		c := n.capByName(cName)
		ackedCaps = append(ackedCaps, c)

		if c != nil {
			c.acknowledged = true
		} else {
			log.Printf("Got an ACK for a CAP %q we dont know about! ignoring!", cName)
		}
	}

	log.Printf("Server ack'd caps: %v", ackedCaps)

	n.doingNegotiation = false
}

func (n *Negotiator) onCapNAK(split []string) {
	// this shouldnt be possible, but if it is, do nothing, the outer client can handle this
	n.reqsSent--
	for _, v := range split {
		if c := n.capByName(v); c != nil {
			c.acknowledged = false
		}
	}
}

func (n *Negotiator) onCapDEL(caps []string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, v := range caps {
		if c := n.capByName(v); c != nil {
			c.available = false
			c.acknowledged = false
		} else {
			log.Printf("unknown cap %q DELeted", v)
		}
	}
}

func (n *Negotiator) onCapNEW(caps []string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, v := range caps {
		name := v
		value := ""

		if strings.Contains(name, "=") {
			split := strings.SplitN(name, "=", 2)

			name = split[0]
			value = split[1]
		}

		if c := n.capByName(v); c != nil {
			c.available = true
			c.value = value
		} else {
			n.capabilities = append(n.capabilities, &capability{
				name:      name,
				value:     value,
				available: true,
			})
		}
	}
}

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
