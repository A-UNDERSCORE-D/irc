package client

import (
	"context"
	"fmt"
	"sync"

	"awesome-dragon.science/go/irc/capab"
	"awesome-dragon.science/go/irc/connection"
	"awesome-dragon.science/go/irc/event"
	"awesome-dragon.science/go/irc/event/irccommand"
	"awesome-dragon.science/go/irc/numerics"
	"awesome-dragon.science/go/irc/user"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("irc-c") //nolint:gochecknoglobals // logger

// Config is a startup configuration for a client instance
type Config struct {
	Connection connection.Config
	Nick       string
	Username   string
	Realname   string

	doSASL       bool
	SASLUsername string
	SASLPassword string

	RequestedCapabilities []string
}

// Client implements a full IRC client for use in bots. It does most of the work
// in connecting and otherwise handling the protocol
type Client struct {
	mu             sync.Mutex
	connection     *connection.Connection
	internalEvents *irccommand.Handler
	clientEvents   event.MessageHandler

	currentNick string

	capabilities *capab.Negotiator
	config       *Config
	// outgoingEvents MessageHandler
}

// New creates a new instance of Client.
func New(config *Config) *Client {
	conn := connection.NewConnection(&config.Connection)
	out := &Client{
		internalEvents: &irccommand.Handler{},
		connection:     conn,
		config:         config,
	}

	out.capabilities = capab.New(&capab.Config{
		ToRequest:    config.RequestedCapabilities,
		SASL:         config.doSASL,
		SASLUsername: config.SASLUsername,
		SASLPassword: config.SASLPassword,
		SASLMech:     "PLAIN",
	}, out.WriteIRC, &irccommand.SimpleHandler{Handler: out.internalEvents})

	out.internalEvents.AddCallback("PING", func(m *event.Message) error {
		return out.WriteIRC("PONG", m.Raw.Params...)
	})

	out.internalEvents.AddCallback(numerics.ERR_NICKNAMEINUSE, func(m *event.Message) error {
		return out.WriteIRC("NICK", m.Raw.Params[1]+"_")
	})

	out.internalEvents.AddCallback(numerics.NICK, func(m *event.Message) error {
		if m.SourceUser.Nick != out.currentNick {
			return nil
		}

		out.mu.Lock()
		out.currentNick = m.Raw.Params[len(m.Raw.Params)-1]
		out.mu.Unlock()

		return nil
	})

	return out
}

// SetMessageHandler sets the callback handler for incoming IRC Messages
func (c *Client) SetMessageHandler(handler event.MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.clientEvents = handler
}

// Run connects to IRC and handles messages until a disconnection occurs
func (c *Client) Run(ctx context.Context) error {
	if err := c.connection.Connect(ctx); err != nil {
		return fmt.Errorf("could not connect to IRC: %w", err)
	}

	// Connection complete, attach line handlers etc
	go c.listenLoop(ctx)

	c.capabilities.Negotiate()

	c.currentNick = c.config.Nick

	if err := c.WriteIRC("NICK", c.config.Nick); err != nil {
		return err
	}

	if err := c.WriteIRC("USER", c.config.Username, "*", "*", c.config.Realname); err != nil {
		return err
	}

	<-c.connection.Done()

	return nil
}

func (c *Client) listenLoop(ctx context.Context) {
	lineChan := c.connection.LineChan()
loop:
	for {
		select {
		case line, ok := <-lineChan:
			if !ok {
				break loop
			}

			c.mu.Lock()
			clientHandler := c.clientEvents
			c.mu.Unlock()

			sourceUser := user.FromMessage(line, c.capabilities.AvailableCaps())

			ev := &event.Message{
				Raw:           line,
				SourceUser:    &sourceUser,
				AvailableCaps: c.capabilities.AvailableCaps(),
			}

			if err := c.internalEvents.OnMessage(ev); err != nil {
				log.Criticalf("Error during internal handling of %q: %s", ev.Raw, err)
			}

			pubEv := &event.Message{
				Raw:           line,
				SourceUser:    &sourceUser,
				CurrentNick:   c.CurrentNick(),
				AvailableCaps: c.capabilities.AvailableCaps(),
			}

			if clientHandler != nil {
				if err := clientHandler.OnMessage(pubEv); err != nil {
					log.Warningf("Error during client handling of %q: %s", ev.Raw, err)
				}
			}

		case <-ctx.Done():
			break loop
		}
	}
}

// WriteIRC constructs an IRC line and sends it to the server
func (c *Client) WriteIRC(command string, params ...string) error {
	if err := c.connection.WriteLine(command, params...); err != nil {
		return fmt.Errorf("client.writeirc: %w", err)
	}

	return nil
}

// SendMessage sends a PRIVMSG to the given target with the given message
func (c *Client) SendMessage(target, message string) error {
	return c.WriteIRC("PRIVMSG", target, message)
}

// SendNotice sends a NOTICE to the given target with the given message
func (c *Client) SendNotice(target, message string) error {
	return c.WriteIRC("NOTICE", target, message)
}

// CurrentNick returns what the Client believes its current nick is. It is safe for concurrent use.
// A client created with New() will internally handle tracking nick changes.
func (c *Client) CurrentNick() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.currentNick
}
