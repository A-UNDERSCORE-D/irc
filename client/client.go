package client

import (
	"context"
	"fmt"
	"log"
	"sync"

	"awesome-dragon.science/go/irc/capab"
	"awesome-dragon.science/go/irc/connection"
	"awesome-dragon.science/go/irc/event"
	"awesome-dragon.science/go/irc/mode"
	"awesome-dragon.science/go/irc/numerics"
	"awesome-dragon.science/go/irc/oper"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/ergochat/irc-go/ircutils"
)

// spell-checker: words sasl

// ClientConfig contains all configuration options for Client
type ClientConfig struct {
	Connection connection.Config

	ServerPassword string

	Nickname string
	Username string
	Realname string

	CTCPResponses map[string]string

	SASL               bool
	NickServAuthUser   string
	NickServAuthPasswd string
	SASLCertPath       string
	SASLCertKeyPath    string
	SASLCertPasswd     string

	JoinChannels []string
	Capabilities []string
}

// NewClient creates a new client that is ready to use
func NewClient(config *ClientConfig) *Client {
	if config.SASL {
		config.Capabilities = append(config.Capabilities, "sasl")
	}

	c := &Client{
		connection:   connection.NewConnection(&config.Connection),
		config:       config,
		log:          log.Default(),
		EventManager: event.NewManager(),
	}

	c.EventManager.AddCallback("NICK", c.onNick, false)
	c.EventManager.AddCallback("JOIN", c.onJoin, false)
	c.EventManager.AddCallback("PART", c.onPart, false)

	c.capabilityNegotiator = capab.New(&capab.Config{
		ToRequest:    config.Capabilities,
		SASL:         config.NickServAuthPasswd != "" && config.NickServAuthUser != "",
		SASLUsername: config.NickServAuthUser,
		SASLPassword: config.NickServAuthUser,
	}, c)

	go func() {
		for line := range c.connection.LineChan() {
			c.onMessage(line)
		}
	}()

	return c
}

// Client is a full fledged IRC "client". It handles a bit of bookkeeping itself
// and provides a frontend for an event system that you can run your own code on
type Client struct {
	connection *connection.Connection
	config     *ClientConfig

	log                  *log.Logger
	EventManager         *event.Manager
	capabilityNegotiator *capab.Negotiator

	mu       sync.Mutex
	nickname string   // the *current* nickname
	channels []string // channels we're in
	oper     bool
	operMech oper.Mechanism
	umodes   []mode.Mode
}

// Connect connects the Client to IRC.
// You probably want Run.
func (c *Client) Connect() error {
	ctx := context.Background()
	if err := c.connection.Connect(ctx); err != nil {
		return err
	}

	return nil
}

// Run starts the connection to IRC and blocks until the connection is closed
func (c *Client) Run() error {
	if err := c.Connect(); err != nil {
		return fmt.Errorf("could not connect: %w", err)
	}

	c.capabilityNegotiator.Negotiate()

	if err := c.Write("NICK", c.config.Nickname); err != nil {
		c.connection.Stop("")

		return fmt.Errorf("could not write NICK command: %w", err)
	}

	if err := c.Write("USER", c.config.Username, "*", "*", c.config.Realname); err != nil {
		c.connection.Stop("")

		return fmt.Errorf("could not write USER command: %w", err)
	}

	<-c.EventManager.WaitFor("001")
	c.log.Print("CONNECTED!!!!!!!!!!!!!!")
	<-c.EventManager.WaitFor(numerics.RPL_ENDOFMOTD)
	c.log.Print(c.connection.ISupport.Modes())

	<-c.connection.Done()

	return nil
}

func (c *Client) onMessage(msg *ircmsg.Message) {
	switch msg.Command {
	case "PING":
		if err := c.Write("PONG", msg.Params...); err != nil {
			log.Printf("Failed to write PONG message! %s", err)
		}

	case "ERROR":
		c.log.Printf("ERROR from server: %v", msg)
	}

	c.EventManager.Fire(msg.Command, msg)
}

func (c *Client) fromMe(msg *ircmsg.Message) bool {
	source := ircutils.ParseUserhost(msg.Prefix)

	return source.Nick == c.nickname
}

func (c *Client) onNick(msg *ircmsg.Message) {
	// :oldnick!*@* NICK newnick
	if !c.fromMe(msg) {
		// Wasn't from us
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.nickname = msg.Params[0]
}

func (c *Client) onJoin(msg *ircmsg.Message) {
	if !c.fromMe(msg) {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.channels = append(c.channels, msg.Params[0])
}

func (c *Client) onPart(msg *ircmsg.Message) {
	if !c.fromMe(msg) {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	newChans := []string{}

	for _, c := range c.channels {
		if c == msg.Params[0] {
			continue
		}

		newChans = append(newChans, c)
	}

	c.channels = newChans
}

// TODO:
// - RPL_YOUREOPER
// - umodes

// Write writes an IRC command to the Server
func (c *Client) Write(command string, args ...string) error {
	return c.connection.WriteLine(command, args...)
}

// Nick returns the client's current nickname
func (c *Client) Nick() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.nickname
}

func (c *Client) AddCallback(name string, f func(*ircmsg.Message), concurrent bool) int {
	return c.EventManager.AddCallback(name, f, concurrent)
}

func (c *Client) RemoveCallback(id int) {
	c.EventManager.RemoveCallback(id)
}

func (c *Client) AddOneShotCallback(name string, f func(*ircmsg.Message), concurrent bool) int {
	return c.EventManager.AddOneShotCallback(name, f, concurrent)
}
