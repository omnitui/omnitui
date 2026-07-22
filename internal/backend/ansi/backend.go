package ansi

import (
	"io"
	"os"
	"sync"

	"github.com/omnitui/omnitui/v2/internal/backend"
	"golang.org/x/term"
)

type Backend struct {
	input          io.Reader
	output         io.Writer
	width, height  int
	events         chan backend.Event
	done           chan struct{}
	closeOnce      sync.Once
	eventsMu       sync.Mutex
	eventsClosed   bool
	sizeMu         sync.RWMutex
	terminal       *os.File
	sizeTerminal   *os.File
	state          *term.State
	outputTerminal *os.File
	outputState    *outputState
}

func New(input io.Reader, output io.Writer) (*Backend, error) {
	if input == nil {
		input = os.Stdin
	}
	if output == nil {
		output = os.Stdout
	}
	b := &Backend{input: input, output: output, width: 80, height: 24, events: make(chan backend.Event, 32), done: make(chan struct{})}
	if file, ok := input.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		b.terminal = file
		b.sizeTerminal = file
		if width, height, err := term.GetSize(int(file.Fd())); err == nil && width > 0 && height > 0 {
			b.width, b.height = width, height
		}
		if state, err := term.MakeRaw(int(file.Fd())); err == nil {
			b.state = state
		}
	}
	if file, ok := output.(*os.File); ok && term.IsTerminal(int(file.Fd())) {
		b.outputTerminal = file
		b.sizeTerminal = file
		if width, height, err := term.GetSize(int(file.Fd())); err == nil && width > 0 && height > 0 {
			b.width, b.height = width, height
		}
		state, err := enableVirtualTerminalOutput(file)
		if err != nil {
			if b.terminal != nil && b.state != nil {
				_ = term.Restore(int(b.terminal.Fd()), b.state)
			}
			return nil, err
		}
		b.outputState = state
	}
	_, _ = io.WriteString(output, "\x1b[?1049h\x1b[?25l\x1b[?1000h\x1b[?1002h\x1b[?1006h")
	go b.readLoop()
	if b.sizeTerminal != nil {
		go b.resizeLoop()
	}
	return b, nil
}

func (b *Backend) Size() (int, int, error) {
	b.sizeMu.RLock()
	defer b.sizeMu.RUnlock()
	return b.width, b.height, nil
}
func (b *Backend) Events() <-chan backend.Event { return b.events }
func (b *Backend) Write(value []byte) error     { _, err := b.output.Write(value); return err }

func (b *Backend) readLoop() {
	defer b.closeEvents()
	parser := &Parser{}
	buffer := make([]byte, 4096)
	for {
		select {
		case <-b.done:
			return
		default:
		}
		n, err := b.input.Read(buffer)
		if n > 0 {
			for _, event := range parser.Feed(buffer[:n]) {
				if !b.emit(event) {
					return
				}
			}
		}
		if err != nil {
			for _, event := range parser.Flush() {
				if !b.emit(event) {
					return
				}
			}
			return
		}
	}
}

func (b *Backend) Close() error {
	var err error
	b.closeOnce.Do(func() {
		close(b.done)
		b.closeEvents()
		if b.terminal != nil && b.state != nil {
			err = term.Restore(int(b.terminal.Fd()), b.state)
		}
		_, writeErr := io.WriteString(b.output, "\x1b[?1006l\x1b[?1002l\x1b[?1000l\x1b[?25h\x1b[?1049l\x1b[0m")
		if err == nil {
			err = writeErr
		}
		if restoreErr := restoreVirtualTerminalOutput(b.outputTerminal, b.outputState); err == nil {
			err = restoreErr
		}
	})
	return err
}

func (b *Backend) emit(event backend.Event) bool {
	b.eventsMu.Lock()
	defer b.eventsMu.Unlock()
	if b.eventsClosed {
		return false
	}
	select {
	case b.events <- event:
		return true
	case <-b.done:
		return false
	}
}

func (b *Backend) closeEvents() {
	b.eventsMu.Lock()
	if !b.eventsClosed {
		b.eventsClosed = true
		close(b.events)
	}
	b.eventsMu.Unlock()
}

func (b *Backend) emitResize() bool {
	width, height, err := term.GetSize(int(b.sizeTerminal.Fd()))
	if err != nil || width <= 0 || height <= 0 {
		return true
	}
	b.sizeMu.Lock()
	changed := width != b.width || height != b.height
	b.width, b.height = width, height
	b.sizeMu.Unlock()
	if !changed {
		return true
	}
	return b.emit(backend.Event{Value: backend.ResizeInput{Width: width, Height: height}})
}
