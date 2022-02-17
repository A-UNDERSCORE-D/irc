package chatcommand

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"

	"awesome-dragon.science/go/irc/event"
	"awesome-dragon.science/go/irc/permissions"
	"awesome-dragon.science/go/irc/user"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("irc") //nolint:gochecknoglobals // logger is fine :P

// Handler implements a chat message command system
type Handler struct {
	Prefix    string
	mu        sync.Mutex // Protects callbacks
	callbacks map[string]*command
	// Function used to send messages. It is expected that this can handle overlong lines cleanly.
	MessageFunc func(string, string) error
	// A permission system to use. If nil, no permission checks take place
	PermissionHandler permissions.Handler
}

// AddCommand errors
var (
	ErrCmdExists      = errors.New("command exists")
	ErrInvalidCmdName = errors.New("invalid command name")
	ErrNoHelp         = errors.New("cannot add a command with no help message")
)

// AddCommand adds a new command to the Handler
func (h *Handler) AddCommand(
	name, help string, requiredPermissions []string, requiredArgs int, callback Callback,
) error {
	name = strings.ToUpper(strings.TrimSpace(name))
	if name == "" || strings.ContainsAny(name, " ") {
		return fmt.Errorf("%w: invalid command name: %q", ErrInvalidCmdName, name)
	}

	upperName := strings.ToUpper(name)
	c := &command{
		name:                upperName,
		help:                help,
		requiredArgs:        requiredArgs,
		requiredPermissions: requiredPermissions,
		callback:            callback,
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if h.callbacks == nil {
		h.callbacks = make(map[string]*command)
	}

	_, exists := h.callbacks[upperName]
	if exists {
		return fmt.Errorf("%w: %s", ErrCmdExists, upperName)
	}

	h.callbacks[upperName] = c

	return nil
}

// OnMessage implements event.MessageHandler
func (h *Handler) OnMessage(msg *event.Message) error {
	if msg.Raw.Command != "PRIVMSG" {
		return nil
	}

	message := msg.Raw.Params[len(msg.Raw.Params)-1]
	messageTarget := msg.Raw.Params[0]

	sourceUser := msg.SourceUser
	if sourceUser == nil {
		u := user.FromMessage(msg.Raw, msg.AvailableCaps)
		sourceUser = &u
	}

	replyTarget := messageTarget
	if replyTarget[0] == '#' {
		replyTarget = sourceUser.Nick
	}

	h.executeCommandIfExists(message, messageTarget, replyTarget, sourceUser, msg.CurrentNick, msg)

	return nil
}

func (h *Handler) getCommand(splitMsg []string, currentNick string) (cmd *command, args []string) {
	if len(splitMsg) == 0 {
		return nil, nil
	}

	cmdName := splitMsg[0]
	args = splitMsg[1:]

	switch {
	case currentNick != "" && strings.HasPrefix(cmdName, currentNick):
		if len(splitMsg) < 2 {
			return nil, nil // Cant extract a command here
		}

		cmdName = splitMsg[1]
		args = splitMsg[2:]

	case strings.HasPrefix(cmdName, h.Prefix):
		cmdName = strings.TrimPrefix(cmdName, h.Prefix)

	default:
		return nil, nil
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	res, exists := h.callbacks[strings.ToUpper(cmdName)]
	if !exists {
		if strings.EqualFold(cmdName, "help") {
			return &command{
				name:                "help",
				help:                "Helps with commands",
				requiredArgs:        0,
				requiredPermissions: nil,
				callback:            h.helpCommandCallback,
			}, nil
		}

		return nil, nil
	}

	return res, args
}

func (h *Handler) helpCommandCallback(args *Argument) error {
	name := ""
	if len(args.Arguments) > 0 {
		name = args.Arguments[0]
	}

	args.Reply(h.DoHelp(name))

	return nil
}

// DoHelp generates the help message for a given command, or a general help message if the command is ""
func (h *Handler) DoHelp(commandName string) string {
	h.mu.Lock()
	defer h.mu.Unlock()

	if commandName == "" {
		keys := make([]string, 0, len(h.callbacks))
		for k := range h.callbacks {
			keys = append(keys, "\x02"+k+"\x02")
		}

		sort.Strings(keys)
		// all commands

		return "Available commands: " + strings.Join(keys, ",")
	}

	commandName = strings.ToUpper(commandName)
	// particular command
	cmd, exists := h.callbacks[commandName]

	if !exists {
		return fmt.Sprintf("\x02%s\x02 does not exist, try %shelp", commandName, h.Prefix)
	}

	return fmt.Sprintf("Help for command \x02%s\x02: %s", commandName, cmd.help)
}

func (h *Handler) reply(target, message string) {
	if err := h.MessageFunc(target, message); err != nil {
		log.Errorf("Failed to send message %q to %q: %s", target, message, err)
	}
}

func (h *Handler) replyf(target, format string, args ...interface{}) {
	h.reply(target, fmt.Sprintf(format, args...))
}

func (h *Handler) executeCommandIfExists(
	message, target, replyTarget string, sourceUser *user.EphemeralUser, currentNick string, ev *event.Message,
) {
	splitMsg := strings.SplitN(message, " ", 2)

	cmd, args := h.getCommand(splitMsg, currentNick)

	if cmd == nil {
		return
	}

	if h.PermissionHandler != nil {
		// Next up, permissions
		allowed, err := h.PermissionHandler.IsAuthorised(sourceUser, cmd.requiredPermissions)
		if err != nil {
			log.Infof("Permission check for %s on command %q errored: %s", sourceUser.UserHost, cmd.name, err)
		}

		if !allowed {
			h.reply(replyTarget, "Access denied.")

			return
		}
	} else {
		log.Debug("Permissions handler is nil. Skipping all permissions checks.")
	}

	if cmd.requiredArgs != -1 && len(args) < cmd.requiredArgs {
		h.replyf(replyTarget, "\x02%s\x02 Requires at least \x02%d\x02 arguments.", cmd.name, cmd.requiredArgs)

		return
	}

	argsToSend := &Argument{
		CommandName: cmd.name,
		Arguments:   args,
		Event:       ev,
		SourceUser:  sourceUser,
		CurrentNick: currentNick,
		Target:      target,
		Reply:       func(msg string) { h.reply(replyTarget, msg) },
	}

	defer func() {
		if res := recover(); res != nil {
			log.Criticalf("Caught panic while running command! %#v", res)
		}
	}()

	if err := cmd.callback(argsToSend); err != nil {
		log.Errorf("Error while running command %q's callback: %s", cmd.name, err)
	}
}
