package multi

import (
	"sync"

	"awesome-dragon.science/go/irc/event"
)

// Handler is a MessageHandler implementation that allows multiple implementations to accept the same messages
type Handler struct {
	handlers []event.MessageHandler
	mu       sync.Mutex
}

// OnMessage implements MessageHandler
func (m *Handler) OnMessage(msg *event.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	outErr := &event.MultiError{}

	for _, h := range m.handlers {
		if err := h.OnMessage(msg); err != nil {
			outErr.Errors = append(outErr.Errors, err)
		}
	}

	if len(outErr.Errors) == 0 {
		return nil
	}

	return outErr
}

// AddHandlers adds a handler to the MultiHandler instance
func (m *Handler) AddHandlers(h ...event.MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.handlers = append(m.handlers, h...)
}

// RemoveHandler removes a handler from the MultiHandler instance
func (m *Handler) RemoveHandler(h event.MessageHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	targetIdx := -1

	for i, v := range m.handlers {
		if v == h {
			targetIdx = i

			break
		}
	}

	if targetIdx == -1 {
		return
	}

	// this is a "normal" delete where a pointer value is involved.
	// see the slicetricks wiki entry on the go github repo for more.
	copy(m.handlers[targetIdx:], m.handlers[targetIdx+1:])
	m.handlers[len(m.handlers)-1] = nil
	m.handlers = m.handlers[:len(m.handlers)-1]
}
