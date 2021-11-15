package ircmsg

import (
	"awesome-dragon.science/go/irc/user"
	"github.com/ergochat/irc-go/ircmsg"
)

//nolint:revive // It'd be literally the same for all of them
// Hook functions
type (
	UserHookFunc        = func(user user.EphemeralUser)
	UserMessageHookFunc = func(user user.EphemeralUser, channel, message string)
	KickFunc            = func(kicker user.EphemeralUser, channel, kickee, message string)
	NickFunc            = func(user user.EphemeralUser, newNick string)
)

//nolint:revive // It'd be literally the same for all of them
// Hook types
type (
	UserHook        struct{ callback UserHookFunc }
	UserMessageHook struct{ callback UserMessageHookFunc }
	KickHook        struct{ callback KickFunc }
	NickHook        struct{ callback NickFunc }
)

func (g *UserHook) Fire(raw *ircmsg.Message) { //nolint:revive // It'd be literally the same for all of them
	g.callback(user.FromMessage(raw))
}

func (g *UserMessageHook) Fire(raw *ircmsg.Message) { //nolint:revive // It'd be literally the same for all of them
	g.callback(user.FromMessage(raw), raw.Params[0], raw.Params[len(raw.Params)-1])
}

func (k *KickHook) Fire(raw *ircmsg.Message) { //nolint:revive // It'd be literally the same for all of them
	k.callback(user.FromMessage(raw), raw.Params[0], raw.Params[1], raw.Params[2])
}

func (n *NickHook) Fire(raw *ircmsg.Message) { //nolint:revive // It'd be literally the same for all of them
	n.callback(user.FromMessage(raw), raw.Params[0])
}
