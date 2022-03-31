package client

import (
	"fmt"

	"awesome-dragon.science/go/irc/util"
)

// WaitForExit blocks until the connection is closed
func (c *Client) WaitForExit() {
	<-c.DoneChan()
}

// DoneChan returns a channel that will be closed when the connection is closed
func (c *Client) DoneChan() <-chan struct{} {
	return c.connection.Done()
}

// Stop stops the bot, quitting with the given message if possible
func (c *Client) Stop(message string) {
	c.connection.Stop(message)
}

const maxMessageLength = 400

// SendMessage sends a PRIVMSG to the given target with the given message.
func (c *Client) SendMessage(target, message string) error {
	return c.WriteIRC("PRIVMSG", target, message)
}

// SendMessageChunked will use SendMessage to send your message, but if the message is over the max for an IRC
// message, it will split it into chunks, and send each one individually.
//
// This is mostly intended for use with things like chatcommand.Handler that dont know how long their messages
// will be. If you can, you likely want to do this manually, as the chunking this uses is not intelligent at all.
func (c *Client) SendMessageChunked(target, message string) error {
	for _, l := range util.ChunkMessage(message, maxMessageLength) {
		if err := c.SendMessage(target, l); err != nil {
			return err
		}
	}

	return nil
}

// SendMessagef is like SendMessage but with printf formatting
func (c *Client) SendMessagef(target, format string, args ...interface{}) error {
	return c.SendMessage(target, fmt.Sprintf(format, args...))
}

// SendNotice sends a NOTICE to the given target with the given message
func (c *Client) SendNotice(target, message string) error {
	return c.WriteIRC("NOTICE", target, message)
}

// SendNoticef is like SendNotice but with printf formatting
func (c *Client) SendNoticef(target, format string, args ...interface{}) error {
	return c.SendNotice(target, fmt.Sprintf(format, args...))
}

// SendNoticeChunked will use SendNotice to send your message, but if the message is over the max for an IRC
// message, it will split it into chunks, and send each one individually.
//
// This is mostly intended for use with things like chatcommand.Handler that dont know how long their messages
// will be. If you can, you likely want to do this manually, as the chunking this uses is not intelligent at all.
func (c *Client) SendNoticeChunked(message, target string) error {
	for _, l := range util.ChunkMessage(message, maxMessageLength) {
		if err := c.SendNotice(target, l); err != nil {
			return err
		}
	}

	return nil
}

// CurrentNick returns what the Client believes its current nick is. It is safe for concurrent use.
// A client created with New() will internally handle tracking nick changes.
func (c *Client) CurrentNick() string {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.currentNick
}

// ToggleRawLog enables or disables raw IRC line logging
func (c *Client) ToggleRawLog() {
	c.config.Connection.RawLog = !c.config.Connection.RawLog
}

/*
Util functions
Join
Part
Quit
Kick
Nick
Privmsg
Notice

these should be ISUPPORT aware where possible, and so on

*/

// func (c *Client) Join(channels ...string) {}

// func (c *Client) Part(channels ...string) {}

// func (c *Client) Nick(newNickname string) {}
