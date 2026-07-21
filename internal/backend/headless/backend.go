package headless

import (
	"sync"

	"github.com/omnitui/omnitui/internal/backend"
)

type Backend struct {
	width, height int
	events        chan backend.Event
	frames        [][]byte
	mu            sync.Mutex
	closed        bool
}

// TestBackend is the descriptive name used by integration tests.
type TestBackend = Backend

func New(width, height int) *Backend {
	return &Backend{width: width, height: height, events: make(chan backend.Event, 64)}
}

func NewTestBackend(width, height int) *Backend { return New(width, height) }
func (b *Backend) Size() (int, int, error)      { return b.width, b.height, nil }
func (b *Backend) Events() <-chan backend.Event { return b.events }
func (b *Backend) Send(value any)               { b.events <- backend.Event{Value: value} }
func (b *Backend) Resize(width, height int) {
	b.Send(backend.ResizeInput{Width: width, Height: height})
}
func (b *Backend) Write(value []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	clone := append([]byte(nil), value...)
	b.frames = append(b.frames, clone)
	return nil
}
func (b *Backend) Frames() [][]byte {
	b.mu.Lock()
	defer b.mu.Unlock()
	result := make([][]byte, len(b.frames))
	for i := range b.frames {
		result[i] = append([]byte(nil), b.frames[i]...)
	}
	return result
}
func (b *Backend) Close() error {
	b.mu.Lock()
	b.closed = true
	b.mu.Unlock()
	close(b.events)
	return nil
}
