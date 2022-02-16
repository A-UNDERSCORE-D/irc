package function

import "awesome-dragon.science/go/irc/client/event"

// FuncHandler is a thin wrapper around a function that implements event.Handler
type FuncHandler func(msg *event.Message) error

// OnMessage redirects the incoming message to the func on FuncHandler
func (f FuncHandler) OnMessage(msg *event.Message) error { return f(msg) }
