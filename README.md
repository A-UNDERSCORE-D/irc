# IRC

IRC is a IRC bot core (or simple connection broker) written to take the place
of all the other ones that I found to be somehow lacking.

# Levels

IRC has two levels that it can operate at, the highest is client (and associated
`bot` types).

## Client

At the client level, most normal IRC procedures are taken care for you. Messages
are parsed out into useful objects, and optionally can be sent over callbacks
designed particularly for those message types.

## Connection

Connection provides only a connection to IRC, and a channel on which parsed
IRC lines will be sent. It will do nothing nice for you, but is a nice place
to start when looking to do things very manually.


