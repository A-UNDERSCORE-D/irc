# IRC

[![Go Reference](https://pkg.go.dev/badge/awesome-dragon.science/go/irc.svg)](https://pkg.go.dev/awesome-dragon.science/go/irc)
[![Go Report Card](https://goreportcard.com/badge/awesome-dragon.science/go/irc)](https://goreportcard.com/report/awesome-dragon.science/go/irc)

IRC is a IRC bot core (or simple connection broker) written to take the place
of all the other ones that I found to be somehow lacking.

## Levels

IRC has two levels that it can operate at, the highest is client (and associated
`bot` types).

### Client

At the client level, most normal IRC procedures are taken care for you. Messages
are parsed out into useful objects, and optionally can be sent over callbacks
designed particularly for those message types.

This is probably what you want to use. It provides a simple and easy frontend to building complex IRC bots.

### Connection

Connection provides only a connection to IRC, and a channel on which parsed
IRC lines will be sent. It will do nothing nice for you, but is a nice place
to start when looking to do things very manually.

This is what Client uses internally for its implementation

## Other bits

There are a whole bunch of other bits implemented for use in bots, such as a mode parser, an ISUPPORT parser, and
multiple different event systems (all of which work together, see how you add handlers on client)
