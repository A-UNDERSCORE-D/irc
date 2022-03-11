package chatcommand

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"awesome-dragon.science/go/irc/event"
	"awesome-dragon.science/go/irc/permissions"
	"awesome-dragon.science/go/irc/permissions/function"
	"awesome-dragon.science/go/irc/user"
	"github.com/ergochat/irc-go/ircmsg"
)

func makeMessage(line string) *ircmsg.Message {
	msg, err := ircmsg.ParseLine(line)
	if err != nil {
		panic(err)
	}

	return &msg
}

func TestHandler_AddCommand(t *testing.T) {
	t.Parallel()

	type args struct {
		name                string
		help                string
		requiredPermissions []string
		requiredArgs        int
		callback            Callback
	}

	tests := []struct {
		name    string
		h       *Handler
		args    args
		wantErr bool
	}{
		{
			name: "normal",
			h:    &Handler{},
			args: args{
				name: "test",
				help: "x",
			},
			wantErr: false,
		},
		{
			name: "bad name",
			h:    &Handler{},
			args: args{
				name: "test is bad",
				help: "",
			},
			wantErr: true,
		},
		{
			name: "already exists",
			h: func() *Handler {
				h := &Handler{}
				_ = h.AddCommand("test", "", nil, -1, nil)

				return h
			}(),

			args:    args{name: "test"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.h.AddCommand(
				tt.args.name, tt.args.help, tt.args.requiredPermissions, tt.args.requiredArgs, tt.args.callback,
			); (err != nil) != tt.wantErr {
				t.Errorf("Handler.AddCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandler_OnMessage(t *testing.T) { //nolint:funlen // TDT and all that
	t.Parallel()

	tests := []struct {
		name       string
		msg        *event.Message
		wantErr    bool
		wantCalled bool
	}{
		{
			name: "good",
			msg: &event.Message{
				Raw:         makeMessage(":test!a@b PRIVMSG bot :bot: test"),
				CurrentNick: "bot",
			},
			wantErr:    false,
			wantCalled: true,
		},
		{
			name: "bad",
			msg: &event.Message{
				Raw:         makeMessage(":test!a@b PRIVMSG bot :bot: test error"),
				CurrentNick: "bot",
			},
			wantErr:    true,
			wantCalled: true,
		},
		{
			name: "panic",
			msg: &event.Message{
				Raw:         makeMessage(":test!a@b PRIVMSG bot :bot: test panic"),
				CurrentNick: "bot",
			},
			wantErr:    true,
			wantCalled: true,
		},
		{
			name: "nocall",
			msg: &event.Message{
				Raw:         makeMessage(":test!a@b PRIVMSG bot :bot: asdf"),
				CurrentNick: "bot",
			},
			wantErr:    false,
			wantCalled: false,
		},
		{
			name:       "not privmsg",
			msg:        &event.Message{Raw: makeMessage(":test!a@b WALLOPS :Sandcats need hugs")},
			wantErr:    false,
			wantCalled: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tt.msg.SourceUser = user.FromMessage(tt.msg.Raw, nil)

			called := false
			h := &Handler{}
			err := h.AddCommand("test", "x", nil, -1, func(a *Argument) error {
				fmt.Println(tt.name, a.Arguments, a.Event.Raw)
				called = true
				fmt.Println(called)

				switch a.ArgString() {
				case "error":
					return errors.New("error") //nolint:goerr113 // Intentional.
				case "panic":
					panic("panic")

				default:
					return nil
				}
			})
			if err != nil {
				t.Errorf("Handler.OnMessage() could not create handler: %s", err)
			}

			if err := h.OnMessage(tt.msg); (err != nil) != tt.wantErr {
				t.Errorf("Handler.OnMessage() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantCalled != called {
				t.Errorf("Handler.OnMessage() called = %t, wantCalled %t", called, tt.wantCalled)
			}
		})
	}
}

func TestHandler_getCommand(t *testing.T) { //nolint:funlen // TDT and all that
	t.Parallel()

	testCmd := &command{
		name:                "TEST",
		help:                "x",
		requiredArgs:        -1,
		requiredPermissions: nil,
		callback:            nil,
	}

	type args struct {
		splitMsg    []string
		currentNick string
	}

	tests := []struct {
		name     string
		args     args
		wantCmd  *command
		wantArgs []string
	}{
		{
			name: "exists",
			args: args{
				splitMsg:    strings.Split("~test test args are testy", " "),
				currentNick: "bot",
			},
			wantCmd:  testCmd,
			wantArgs: []string{"test", "args", "are", "testy"},
		},
		{
			name: "exists with ping and arg",
			args: args{
				splitMsg:    strings.Split("bot: test ping message", " "),
				currentNick: "bot",
			},
			wantCmd:  testCmd,
			wantArgs: []string{"ping", "message"},
		},
		{
			name: "exists with ping and no arg",
			args: args{
				splitMsg:    strings.Split("bot: test", " "),
				currentNick: "bot",
			},
			wantCmd:  testCmd,
			wantArgs: []string{},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &Handler{}
			h.Prefix = "~"
			if err := h.AddCommand("test", "x", nil, -1, nil); err != nil {
				t.Errorf("could not create test command: %s", err)
			}

			gotCmd, gotArgs := h.getCommand(tt.args.splitMsg, tt.args.currentNick)
			if !reflect.DeepEqual(gotCmd, tt.wantCmd) {
				t.Errorf("Handler.getCommand() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("Handler.getCommand() gotArgs = %#v, want %#v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestHandler_DoHelp(t *testing.T) {
	t.Parallel()

	h := &Handler{}
	if err := h.AddCommand("test", "this is a test", nil, -1, nil); err != nil {
		t.Errorf("could not add command: %s", err)
	}

	tests := []struct {
		name string
		args string
		want string
	}{
		{
			name: "exists",
			args: "test",
			want: "Help for command \x02TEST\x02: this is a test",
		},
		{
			name: "empty",
			args: "",
			want: "Available commands: \x02TEST\x02",
		},
		{
			name: "noexist",
			args: "awoo", // No awoo. $500 fine.
			want: "\x02AWOO\x02 does not exist, try help",
		},
	}

	for _, tt := range tests {
		tt := tt

		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if res := h.DoHelp(tt.args); res != tt.want {
				t.Errorf("Handler.DoHelp() = %q, want %q", res, tt.want)
			}
		})
	}
}

func TestHandler_reply(t *testing.T) {
	t.Parallel()

	called := false

	h := &Handler{}

	h.MessageFunc = func(string, string) error {
		called = true

		return nil
	}
	if err := h.AddCommand("test", "test", nil, -1, func(a *Argument) error {
		a.Reply("")

		return nil
	}); err != nil {
		t.Errorf("could not add command: %s", err)
	}

	m := makeMessage(":x!x@x PRIVMSG bot :bot: test")

	if err := h.OnMessage(&event.Message{
		Raw:         m,
		SourceUser:  user.FromMessage(m, nil),
		CurrentNick: "bot",
	}); err != nil {
		t.Errorf("could not call OnMessage: %s", err)
	}

	if !called {
		t.Error("reply did not call MessageFunc")
	}
}

func TestHandler_replyf(t *testing.T) {
	type args struct {
		target string
		format string
		args   []interface{}
	}
	tests := []struct {
		name string
		h    *Handler
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tt.h.replyf(tt.args.target, tt.args.format, tt.args.args...)
		})
	}
}

func TestHandler_callOK(t *testing.T) {
	t.Parallel()

	type args struct {
		cmd        *command
		sourceUser *user.EphemeralUser
		args       []string
	}
	tests := []struct {
		name              string
		permissionHandler permissions.Handler
		args              args
		want              bool
		wantReply         string
	}{
		{
			name:              "simple",
			permissionHandler: function.Handler(func(*user.EphemeralUser, []string) (bool, error) { return true, nil }),
			args: args{
				cmd: &command{
					name: "test",
					help: "test",
				},
			},
			want:      true,
			wantReply: "",
		},
		{
			name: "error",
			permissionHandler: function.Handler(func(*user.EphemeralUser, []string) (bool, error) {
				return false, errors.New("x") //nolint:goerr113 // testing
			}),
			args:      args{cmd: &command{}},
			want:      false,
			wantReply: "Access denied.",
		},
		{
			name:              "too few args",
			permissionHandler: function.Handler(func(*user.EphemeralUser, []string) (bool, error) { return true, nil }),
			args:              args{cmd: &command{requiredArgs: 3}, args: []string{"too few args"}},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			h := &Handler{}
			reply := ""
			h.MessageFunc = func(_, s2 string) error { reply = s2; return nil } //nolint:nlreturn // oneliner is fine

			h.PermissionHandler = tt.permissionHandler
			if got := h.callOK(tt.args.cmd, "", tt.args.sourceUser, tt.args.args); got != tt.want {
				t.Errorf("Handler.callOK() = %v, want %v", got, tt.want)
			}

			if tt.wantReply != "" && tt.wantReply != reply {
				t.Errorf("Handler.callOK() reply = %q, want %q", reply, tt.wantReply)
			}
		})
	}
}

func TestHandler_RemoveCommand(t *testing.T) {
	t.Parallel()
	h := &Handler{
		Prefix:            "~",
		MessageFunc:       func(string, string) error { panic("not implemented") },
		PermissionHandler: nil,
	}

	callCount := 0

	ev := &event.Message{
		SourceUser: &user.EphemeralUser{
			User: user.User{NUH: ircmsg.NUH{Name: "a", User: "b", Host: "c"}},
		},
		CurrentNick:   "test",
		AvailableCaps: nil,
	}

	h.AddCommand("test", "test", nil, -1, func(a *Argument) error { callCount++; return nil })

	if err := h.executeCommandIfExists("~test", "a", "b", ev); err != nil {
		t.Errorf("Handler.RemoveCommand() returned error from callback: %s", err)
	}

	if callCount != 1 {
		t.Error("Handler.RemoveCommand() callback did not correctly fire")
	}

	if err := h.RemoveCommand("test"); err != nil {
		t.Errorf("Handler.RemoveCommand() failed to remove command: %s", err)
	}

	if err := h.executeCommandIfExists("~test", "a", "b", ev); err != nil {
		t.Errorf("Handler.RemoveCommand() running removed command errored: %s", err)
	}

	if callCount != 1 {
		t.Errorf("Handler.RemoveCommand() running removed command executed callback (%d)", callCount)
	}
}
