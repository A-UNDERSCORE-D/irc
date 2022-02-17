package chatcommand //nolint:testpackage // I want to test internals here too

import (
	"reflect"
	"testing"

	"awesome-dragon.science/go/irc/event"
	"awesome-dragon.science/go/irc/user"
)

func TestHandler_AddCommand(t *testing.T) {
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.h.AddCommand(tt.args.name, tt.args.help, tt.args.requiredPermissions, tt.args.requiredArgs, tt.args.callback); (err != nil) != tt.wantErr {
				t.Errorf("Handler.AddCommand() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandler_OnMessage(t *testing.T) {
	type args struct {
		msg *event.Message
	}
	tests := []struct {
		name    string
		h       *Handler
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.h.OnMessage(tt.args.msg); (err != nil) != tt.wantErr {
				t.Errorf("Handler.OnMessage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandler_getCommand(t *testing.T) {
	type args struct {
		splitMsg    []string
		currentNick string
	}
	tests := []struct {
		name     string
		h        *Handler
		args     args
		wantCmd  *command
		wantArgs []string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotCmd, gotArgs := tt.h.getCommand(tt.args.splitMsg, tt.args.currentNick)
			if !reflect.DeepEqual(gotCmd, tt.wantCmd) {
				t.Errorf("Handler.getCommand() gotCmd = %v, want %v", gotCmd, tt.wantCmd)
			}
			if !reflect.DeepEqual(gotArgs, tt.wantArgs) {
				t.Errorf("Handler.getCommand() gotArgs = %v, want %v", gotArgs, tt.wantArgs)
			}
		})
	}
}

func TestHandler_helpCommandCallback(t *testing.T) {
	type args struct {
		args *Argument
	}
	tests := []struct {
		name    string
		h       *Handler
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.h.helpCommandCallback(tt.args.args); (err != nil) != tt.wantErr {
				t.Errorf("Handler.helpCommandCallback() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandler_DoHelp(t *testing.T) {
	type args struct {
		commandName string
	}
	tests := []struct {
		name string
		h    *Handler
		args args
		want string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := tt.h.DoHelp(tt.args.commandName); got != tt.want {
				t.Errorf("Handler.DoHelp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_reply(t *testing.T) {
	type args struct {
		target  string
		message string
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
			tt.h.reply(tt.args.target, tt.args.message)
		})
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

func TestHandler_executeCommandIfExists(t *testing.T) {
	type args struct {
		message     string
		target      string
		replyTarget string
		sourceUser  *user.EphemeralUser
		currentNick string
		ev          *event.Message
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
			tt.h.executeCommandIfExists(tt.args.message, tt.args.target, tt.args.replyTarget, tt.args.sourceUser, tt.args.currentNick, tt.args.ev)
		})
	}
}
