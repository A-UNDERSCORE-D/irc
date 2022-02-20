// Package capab implements IRCv3 Capability negotiation
package capab

import (
	"strings"
	"sync"

	"awesome-dragon.science/go/irc/numerics"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("irc-capab") //nolint:gochecknoglobals // logger

type eventManager interface {
	AddCallback(string, func(*ircmsg.Message) error) int
	RemoveCallback(int)
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

// Capability represents a single IRCv3 capability
type Capability struct {
	Name         string
	Value        string
	Available    bool
	Request      bool
	Acknowledged bool
}

func (c *Capability) String() string {
	return c.Name
}

// Negotiator negotiates IRCv3 capabilities over a Client instance
type Negotiator struct {
	mu sync.Mutex

	eventManager eventManager
	writeIRC     func(string, ...string) error
	config       *Config

	capabilities []*Capability
	incomingCaps []string

	doingNegotiation bool
	requestsSent     int
}

// New creates a new Negotiator instance
func New(conf *Config, writeToIRC func(string, ...string) error, eventManager eventManager) *Negotiator {
	out := &Negotiator{
		config:       conf,
		writeIRC:     writeToIRC,
		eventManager: eventManager,
	}

	for _, c := range conf.ToRequest {
		out.capabilities = append(out.capabilities, &Capability{Name: c, Request: true})
	}

	return out
}

// Negotiate negotiates IRCv3 capabilities with a server, and optionally performs
// sasl authentication
func (n *Negotiator) Negotiate() {
	if len(n.capabilities) == 0 {
		// None to request, dont do anything
		return
	}
	n.doNegotiation()

	if err := n.doSasl(); err != nil {
		log.Errorf("Failed SASL: %s", err)
	}

	// Add NEW/DEL
	n.eventManager.AddCallback("CAP", func(msg *ircmsg.Message) error {
		split := strings.Split(msg.Params[len(msg.Params)-1], " ")

		switch cmd := msg.Params[1]; cmd {
		case "NEW":
			n.onCapNEW(split)
		case "DEL":
			n.onCapDEL(split)
		}

		return nil
	})

	_ = n.writeIRC("CAP", "END")
}

// AvailableCaps returns a list of capabilities that have been requested and acknowledged by the server
func (n *Negotiator) AvailableCaps() []Capability {
	n.mu.Lock()
	defer n.mu.Unlock()

	out := []Capability{} // explicitly not a pointer, we want copies in here

	for _, c := range n.capabilities {
		if c.Acknowledged {
			out = append(out, *c)
		}
	}

	return out
}

func (n *Negotiator) doNegotiation() {
	msgChan := make(chan *ircmsg.Message)

	capCallback := n.eventManager.AddCallback("CAP", func(msg *ircmsg.Message) error {
		msgChan <- msg

		return nil
	})
	welcomeCallback := n.eventManager.AddCallback(
		numerics.RPL_WELCOME,
		func(msg *ircmsg.Message) error {
			msgChan <- msg

			return nil
		},
	)

	defer n.eventManager.RemoveCallback(capCallback)
	defer n.eventManager.RemoveCallback(welcomeCallback)
	n.doingNegotiation = true
	_ = n.writeIRC("CAP", "LS", "302")

	for n.doingNegotiation {
		msg := <-msgChan

		if msg.Command == numerics.RPL_WELCOME {
			log.Warning("Got unexpected 001. Assuming the server does not support capabilities")

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
			log.Infof("Unknown CAP command %q. ignoring", cmd)
		}
	}
}

func (n *Negotiator) onCapLS(caps []string, moreComing bool) {
	n.incomingCaps = append(n.incomingCaps, caps...)

	if moreComing {
		return
	}

	// No more coming
	log.Infof("Server offered caps %v", n.incomingCaps)
	n.parseCaps()
	n.incomingCaps = nil // clear this for use in ACK later
	n.requestCaps()
}

func (n *Negotiator) requestCaps() {
	var toRequest []*Capability

	for _, c := range n.capabilities {
		if c.Request && c.Available {
			toRequest = append(toRequest, c)
		}
	}

	var (
		lines   []string
		builder strings.Builder
	)

	for _, c := range toRequest {
		if builder.Len()+len(c.Name) >= 450 {
			lines = append(lines, strings.TrimSpace(builder.String()))
			builder.Reset()
		}

		builder.WriteString(c.Name)
		builder.WriteRune(' ')
	}

	lines = append(lines, strings.TrimSpace(builder.String()))

	log.Infof("Requesting capabilities %v", toRequest)

	n.requestsSent += len(lines)

	for _, l := range lines {
		_ = n.writeIRC("CAP", "REQ", l)
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
			c.Available = true
			c.Value = value

			continue
		}

		n.capabilities = append(n.capabilities, &Capability{
			Name:         name,
			Value:        value,
			Available:    true,
			Request:      false,
			Acknowledged: false,
		})
	}
}

func (n *Negotiator) capByName(name string) *Capability {
	for _, capab := range n.capabilities {
		if capab.Name == name {
			return capab
		}
	}

	return nil
}

func (n *Negotiator) onCapACK(caps []string) {
	// process ack into things and stuff
	n.incomingCaps = append(n.incomingCaps, caps...)
	n.requestsSent--
	// more coming?
	if n.requestsSent > 0 {
		return
	}

	// no more coming
	ackedCaps := make([]*Capability, 0, len(n.incomingCaps))

	for _, cName := range n.incomingCaps {
		c := n.capByName(cName)
		ackedCaps = append(ackedCaps, c)

		if c != nil {
			c.Acknowledged = true
		} else {
			log.Warningf("Got an ACK for a CAP %q we dont know about! ignoring!", cName)
		}
	}

	log.Infof("Server ack'd caps: %v", ackedCaps)

	n.doingNegotiation = false
}

func (n *Negotiator) onCapNAK(split []string) {
	n.mu.Lock()
	defer n.mu.Unlock()
	// this shouldn't be possible, but if it is, do nothing, the outer client can handle this
	n.requestsSent--
	for _, v := range split {
		if c := n.capByName(v); c != nil {
			c.Acknowledged = false
		}
	}
}

func (n *Negotiator) onCapDEL(caps []string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	for _, v := range caps {
		if c := n.capByName(v); c != nil {
			c.Available = false
			c.Acknowledged = false
		} else {
			log.Warningf("unknown cap %q DELeted", v)
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
			c.Available = true
			c.Value = value
		} else {
			n.capabilities = append(n.capabilities, &Capability{
				Name:      name,
				Value:     value,
				Available: true,
			})
		}
	}
}
