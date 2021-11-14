package capab

import (
	"log"
	"strings"
	"sync"

	"awesome-dragon.science/go/irc/numerics"
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
