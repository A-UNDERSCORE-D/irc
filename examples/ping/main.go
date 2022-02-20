package main

import (
	"context"

	"awesome-dragon.science/go/irc/client"
	"awesome-dragon.science/go/irc/connection"
	"awesome-dragon.science/go/irc/event/chatcommand"
)

// A simple IRC bot with a single command that returns PONG! followed by anything it was given

func main() {
	// Instantiate bot
	b := client.New(&client.Config{
		Connection: connection.Config{
			Host:   "irc.libera.chat",
			Port:   "6697",
			TLS:    true,
			RawLog: true,
		},
		Nick:     "goIRCTest",
		Username: "GIT",
		Realname: "Example awesome-dragon.science/go/irc bot",
	})

	// Create a command handler
	h := &chatcommand.Handler{
		Prefix:      "~",
		MessageFunc: b.SendMessage,
	}

	// Set it as the message handler for the bot (see multi.Handler for how to use multiple)
	b.SetMessageHandler(h)

	// Add the simple command.
	h.AddCommand("ping", "Example command that returns PONG", nil, -1, func(a *chatcommand.Argument) error {
		a.Replyf("PONG! %s", a.ArgString())
		return nil
	})

	// Run forever
	b.Run(context.Background())
}
