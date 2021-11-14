// Package event defines a simple event system for use with irc
package event_test

import (
	"reflect"
	"runtime"
	"sync"
	"testing"
	"time"

	"awesome-dragon.science/go/irc/event"
	"github.com/ergochat/irc-go/ircmsg"
)

// spell-checker: words ircmsg
func TestManager_Fire(t *testing.T) {
	t.Parallel()

	t.Run("multiple", func(t *testing.T) {
		t.Parallel()
		manager := event.NewManager()
		count := 0
		total := 10
		for i := 0; i < total; i++ {
			manager.AddCallback("test", func(*ircmsg.Message) { count++ }, false)
		}

		manager.Fire("test", &ircmsg.Message{})

		if count != total {
			t.Fatalf("Expected %d callbacks fired, %d fired instead", total, count)
		}
	})

	t.Run("multiple goroutine", func(t *testing.T) {
		t.Parallel()

		m := event.NewManager()
		wg := sync.WaitGroup{}
		total := 10
		donechan := make(chan int)
		for i := 0; i < total; i++ {
			wg.Add(1)
			m.AddCallback("test", func(*ircmsg.Message) { donechan <- 1; wg.Done() }, true)
		}
		valueChan := make(chan int)
		go func() {
			count := 0
			for i := range donechan {
				count += i
			}
			valueChan <- count
		}()

		m.Fire("test", &ircmsg.Message{})
		wg.Wait()
		close(donechan)
		runtime.Gosched()
		if res := <-valueChan; res != total {
			t.Fatalf("Expected %d callbacks, got %d", total, res)
		}
	})
}

func TestManager_RemoveCallback(t *testing.T) {
	t.Parallel()

	t.Run("exists", func(t *testing.T) {
		t.Parallel()

		manager := event.NewManager()

		res := manager.AddCallback("test", func(*ircmsg.Message) {}, false)
		if manager.GetEvent(res) == nil {
			t.Fatalf("Unexpected nil callback")
		}

		manager.RemoveCallback(res)

		if manager.GetEvent(res) != nil {
			t.Fatalf("Failed to remove callback")
		}
	})

	t.Run("not exist", func(t *testing.T) {
		t.Parallel()

		m := event.NewManager()
		m.RemoveCallback(1337) // Doesn't exist, should not error
	})
}

func TestManager_GetEvent_AddCallback(t *testing.T) {
	t.Parallel()
	t.Run("Exists", func(t *testing.T) {
		t.Parallel()
		m := event.NewManager()
		cb := func(msg *ircmsg.Message) {}
		res := m.AddCallback("test", cb, false)

		if ev := (m.GetEvent(res)); ev == nil || reflect.DeepEqual(cb, ev.Func()) || ev.ID() != res {
			t.Fatalf("GetEvent returned unexpected result %#v", ev)
		}
	})

	t.Run("No Exist", func(t *testing.T) {
		t.Parallel()
		m := event.NewManager()
		if ev := m.GetEvent(1337); ev != nil {
			t.Fatalf("GetEvent expected nil")
		}
	})
}

func TestManager_AddOneShotCallback(t *testing.T) {
	t.Parallel()

	m := event.NewManager()
	called := 0

	m.AddOneShotCallback("test", func(*ircmsg.Message) { called++ }, false)

	for i := 0; i < 10; i++ {
		m.Fire("test", &ircmsg.Message{})
	}

	if called != 1 {
		t.Fatalf("Expected exactly 1 call of a OneShot callback. Got %d", called)
	}
}

func TestManager_WaitFor(t *testing.T) {
	t.Parallel()

	m := event.NewManager()
	c := m.WaitFor("test")
	done := make(chan bool)

	go func() {
		select {
		case <-c:
			done <- true
		case <-time.After(time.Second):
			done <- false
		}
	}()

	m.Fire("test", &ircmsg.Message{})

	ok := <-done
	if !ok {
		t.Fatal("Timeout Exceeded")
	}
}
