package raw

import (
	"sync"

	"awesome-dragon.science/go/irc/event"
)

// Handler is a simple MessageHandler implementation that simply passes the message along
type Handler struct {
	listeners map[int]func(msg *event.Message) error
	lastIdx   int
	mu        sync.Mutex
}

// OnMessage is the main interface to this Handler. It implements client.MessageHandler
func (r *Handler) OnMessage(msg *event.Message) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	outErr := &event.MultiError{}

	for _, l := range r.listeners {
		if err := l(msg); err != nil {
			outErr.Errors = append(outErr.Errors, err)
		}
	}

	if outErr.Errors == nil {
		return nil
	}

	return outErr
}

// AddCallback adds a callback to RawHandler's callback list
func (r *Handler) AddCallback(f func(msg *event.Message) error) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lastIdx++
	r.listeners[r.lastIdx] = f

	return r.lastIdx
}

// RemoveCallback removes the callback specified by id, if it exists.
func (r *Handler) RemoveCallback(id int) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.listeners, id)
}
