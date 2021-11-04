package irc

import (
	"context"
)

// spell-checker: words sasl

// ClientConfig contains all configuration options for Client
type ClientConfig struct {
	Connection ConnectionConfig

	ServerPassword string

	Nickname string
	Username string
	Realname string

	CTCPResponses map[string]string

	SASL               bool
	SASLMech           string
	NickServAuthUser   string
	NickServAuthPasswd string

	JoinChannels []string
}

func NewClient() *Client {
	return nil
}

// Client is a full fledged IRC "client". It handles a bit of bookkeeping itself
// and provides a frontend for an event system that you can run your own code on
type Client struct {
	connection *Connection
}

func (c *Client) Connect() {
	ctx := context.Background()
	c.connection.Connect(ctx)
}
