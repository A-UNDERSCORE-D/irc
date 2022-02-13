package client

import (
	"context"
	"fmt"
	"sync"

	"awesome-dragon.science/go/irc/capab"
	"awesome-dragon.science/go/irc/client/event"
	"awesome-dragon.science/go/irc/client/event/irccommand"
	"awesome-dragon.science/go/irc/connection"
	"github.com/ergochat/irc-go/ircmsg"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("ircClient")

type MessageSource interface {
	LineChan() <-chan *ircmsg.Message
}

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

	capabilities *capab.Negotiator
	// outgoingEvents MessageHandler
}

// New creates a new instance of Client.
func New(config *Config) *Client {
	conn := connection.NewConnection(&config.Connection)
	out := &Client{
		internalEvents: &irccommand.Handler{},
		connection:     conn,
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

	c.WriteIRC("NICK", "test")
	c.WriteIRC("USER", "dergtest", "*", "*", "asdf test")

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

			ev := &event.Message{
				Raw:           line,
				AvailableCaps: c.capabilities.AvailableCaps(),
			}

			if err := c.internalEvents.OnMessage(ev); err != nil {
				log.Criticalf("Error during internal handling of %q: %s", ev.Raw, err)
			}

			if clientHandler != nil {
				if err := clientHandler.OnMessage(ev); err != nil {
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
	return fmt.Errorf("WriteIRC: %w", c.connection.WriteLine(command, params...))
}

func (c *Client) SendMessage(target, message string) error {
	return c.WriteIRC("PRIVMSG", target, message)
}

func (c *Client) SendNotice(target, message string) error {
	return c.WriteIRC("NOTICE", target, message)
}
