// Package ircmsg provides a dispatch system for IRC message events
package ircmsg

import (
	"strings"
	"sync"

	"github.com/ergochat/irc-go/ircmsg"
)

// Hooks is an event engine for IRC messages, unlike the engine found in the parent
// directory, this engine expands IRC messages out to the important bits, rather than
// providing them all
type Hooks struct {
	mu    sync.Mutex
	hooks map[string][]Hook
}

func (h *Hooks) OnMessage(msg *ircmsg.Message) {
	h.mu.Lock()
	// make a copy, that way if something adds a hook it doesn't get funny
	hooks := append([]Hook(nil), h.hooks[strings.ToLower(msg.Command)]...)
	h.mu.Unlock()

	for _, h := range hooks {
		h.Fire(msg)
	}

	// TODO: snotes
}

// Hook implementations extract ircmsg.Message instances into distinct message types
type Hook interface {
	Fire(raw *ircmsg.Message)
}

// New creates a New hooks instance ready for use
func New() *Hooks {
	return &Hooks{
		hooks: make(map[string][]Hook),
	}
}

// Hook hooks a hook onto a name (say that four times fast)
func (h *Hooks) Hook(name string, hook Hook) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.hooks[strings.ToLower(name)] = append(h.hooks[strings.ToLower(name)], hook)
}

// HookJOIN hooks a method onto a JOIN command
func (h *Hooks) HookJOIN(cb UserHookFunc) {
	h.Hook("join", &UserHook{callback: cb})
}

// HookPART a method onto a PART command
func (h *Hooks) HookPART(cb UserMessageHookFunc) { h.Hook("part", &UserMessageHook{callback: cb}) }

// HookQUIT hooks a method onto a QUIT command
func (h *Hooks) HookQUIT(cb UserMessageHookFunc) { h.Hook("quit", &UserMessageHook{callback: cb}) }

// HookKICK hooks a method onto a KICK command
func (h *Hooks) HookKICK(cb KickFunc) { h.Hook("kick", &KickHook{callback: cb}) }

// HookNICK hooks a method onto a NICK command
func (h *Hooks) HookNICK(cb NickFunc) { h.Hook("nick", &NickHook{callback: cb}) }

// HookTOPIC hooks a method onto a TOPIC command
func (h *Hooks) HookTOPIC(cb UserMessageHookFunc) { h.Hook("nick", &UserMessageHook{callback: cb}) }
