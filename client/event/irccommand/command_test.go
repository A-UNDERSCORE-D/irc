package irccommand

import (
	"fmt"
	"reflect"
	"testing"

	"awesome-dragon.science/go/irc/client/event"
	"github.com/ergochat/irc-go/ircmsg"
)

func Test_keys(t *testing.T) {
	t.Parallel()

	m := map[int]event.CallbackFunc{1: nil, 3: nil, 2: nil, 8: nil, 1337: nil}

	want := []int{1, 2, 3, 8, 1337}

	if res := keys(m); !reflect.DeepEqual(res, want) {
		t.Errorf("keys() = %v, want %v", res, want)
	}
}

func TestHandler_collectCallbacks(t *testing.T) {
	t.Parallel()

	h := &Handler{
		hooks: map[string]map[int]event.CallbackFunc{
			"TEST": {
				0: func(*event.Message) error { panic("not implemented") },
				2: func(*event.Message) error { panic("not implemented") },
				4: func(*event.Message) error { panic("not implemented") },
				1: func(*event.Message) error { panic("not implemented") },
				5: func(*event.Message) error { panic("not implemented") },
				6: func(*event.Message) error { panic("not implemented") },
				3: func(*event.Message) error { panic("not implemented") },
			},
		},
	}

	want := []event.CallbackFunc{}
	for i := 0; i < 7; i++ {
		want = append(want, h.hooks["TEST"][i])
	}

	res := h.collectCallbacks("TEST", false)
	for i := range want {
		if fmt.Sprint(want[i]) != fmt.Sprint(res[i]) {
			t.Errorf("Handler.collectCallbacks() = %v, want %v", res, want)
		}
	}
}

func mustParseLine(line string) *ircmsg.Message {
	res, err := ircmsg.ParseLine(line)
	if err != nil {
		panic(err)
	}

	return &res
}

func TestHandler_OnMessage(t *testing.T) {
	t.Parallel()

	var called bool

	h := &Handler{}

	h.AddCallback("test", func(*event.Message) error {
		called = true

		return nil
	})

	if err := h.OnMessage(&event.Message{Raw: mustParseLine(":a!b@c TEST stuff")}); err != nil {
		t.Errorf("Error returned from OnMessage: %s", err)
	}

	if !called {
		t.Error("OnMessage did not call callback")
	}

	called = false

	if err := h.OnMessage(&event.Message{Raw: mustParseLine(":a!b@c PRIVMSG #libera :beep")}); err != nil {
		t.Errorf("Error returned from OnMessage: %s", err)
	}

	if called {
		t.Error("OnMessage called when it should not have")
	}
}

func TestHandler_AddCallback(t *testing.T) {
	h := &Handler{}

	h.AddCallback("test", func(m *event.Message) error { panic("not implemented") })

	if len(h.hooks) != 1 {
		t.Error("did not add callback entry to map")
	}

	if len(h.hooks["TEST"]) != 1 {
		t.Error("did not add callback function to map")
	}
}

func TestHandler_RemoveCallback(t *testing.T) {
	t.Parallel()
	t.Run("single", func(t *testing.T) {
		t.Parallel()
		h := &Handler{}
		id := h.AddCallback("test", func(m *event.Message) error { panic("Not implemented") })

		h.RemoveCallback(id)

		if len(h.hooks) != 0 {
			t.Error("Remove did not correctly reset hooks")
		}
	})

	t.Run("multi", func(t *testing.T) {
		t.Parallel()
		h := &Handler{}
		id := h.AddCallback("test", func(m *event.Message) error { panic("Not implemented") })
		_ = h.AddCallback("test", func(m *event.Message) error { panic("Not implemented") })
		_ = h.AddCallback("otherTest", func(m *event.Message) error { panic("Not implemented") })

		h.RemoveCallback(id)

		if len(h.hooks) != 2 {
			t.Error("Remove did not correctly reset hooks")
		}

		if len(h.hooks["TEST"]) != 1 {
			t.Log(h.hooks)
			t.Error("Removed too many hooks")
		}

		if len(h.hooks["OTHERTEST"]) != 1 {
			t.Error("Removed unexpected hook")
		}
	})
}

func TestHandler_WaitFor(t *testing.T) {
	t.Parallel()
	h := &Handler{}

	c := h.WaitFor("TEST")

	h.OnMessage(&event.Message{Raw: mustParseLine(":a!b@c TEST")})

	select {
	case <-c:
	default:
		t.Error("Channel did not have line passed to it")
	}
}
