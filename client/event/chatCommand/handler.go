package chatcommand

import (
	"strings"

	"awesome-dragon.science/go/irc/client/event"
	"awesome-dragon.science/go/irc/permissions"
	"awesome-dragon.science/go/irc/user"
)

type Command struct {
	name                string
	help                string
	requiredArgs        int
	requiredPermissions int
	callback            func(Argument) error
}

// Handler implements a chat message command system
type Handler struct {
	prefix            string
	callbacks         map[string]*Command
	NoticeFunc        func(string, string) error
	PrivmsgFunc       func(string, string) error
	PermissionHandler permissions.Handler
	PreferNotices     bool
}

// OnMessage implements event.MessageHandler
func (h *Handler) OnMessage(msg *event.Message) error {
	if msg.Raw.Command != "PRIVMSG" {
		return nil
	}

	message := msg.Raw.Params[len(msg.Raw.Params)-1]
	targetChan := msg.Raw.Params[0]
	sourceUser := msg.SourceUser
	if sourceUser == nil {
		u := user.FromMessage(msg.Raw, msg.AvailableCaps)
		sourceUser = &u
	}

	h.checkCommand(message, targetChan, sourceUser, msg.CurrentNick)

	return nil
}

func (h *Handler) extractPossibleCommand(message, currentNick string) (string, string) {
	message = strings.TrimSpace(message)
	if len(message) == 0 {
		return "", ""
	}

	split := strings.Split(message, " ")

	if currentNick != "" && strings.HasPrefix(split[0], currentNick) {
		split = split[1:]
		if len(split) == 0 {
			return "", ""
		}
	}

	split[0] = strings.TrimPrefix(split[0], h.prefix)

	return split[0], strings.Join(split[1:], " ")
}

func (h *Handler) getCommand2()

func (h *Handler) getCommand(splitMsg []string, currentNick string) (*Command, []string) {
	if len(splitMsg) == 0 {
		return nil, nil
	}

	cmd := splitMsg[0]
	args := splitMsg[1:]

	if currentNick != "" && strings.HasPrefix(cmd, currentNick) {
		if len(splitMsg) < 2 {
			return nil, nil // Cant extract a command here
		}
		cmd = splitMsg[1]
		args = splitMsg[2:]
	} else if strings.HasPrefix(cmd, h.prefix) {
		cmd = strings.TrimPrefix(cmd, h.prefix)
	} else {
		return nil, nil
	}

	res, exists := h.callbacks[strings.ToLower(cmd)]
	if !exists {
		return nil, nil
	}

	return res, args
}

func (h *Handler) checkCommand(message, targetchan string, sourceUser *user.EphemeralUser, currentNick string) {
	splitMsg := strings.SplitN(message, " ", 2)

	cmd, args := h.getCommand(splitMsg, currentNick)

	if cmd == nil {
		return
	}
	_ = args
	_ = cmd
}
