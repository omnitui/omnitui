package omnitui

import (
	"context"
	"errors"
	"sync/atomic"

	"github.com/viniciusfonseca/omnitui/internal/backend"
	"github.com/viniciusfonseca/omnitui/internal/backend/ansi"
	"github.com/viniciusfonseca/omnitui/internal/core"
	"github.com/viniciusfonseca/omnitui/internal/screen"
)

var ErrInterrupted = errors.New("omnitui: interrupted")

type dispatcher struct{ enqueue func(func()) }

type App struct {
	root            Element
	options         Options
	work            chan func()
	dispatcher      *dispatcher
	running         atomic.Bool
	invalidated     bool
	rootInstance    *instance
	focused         *instance
	capture         *instance
	pressTarget     *instance
	hoverPath       []*instance
	width, height   int
	front, back     *screen.Buffer
	pendingMessages []any
	backend         backend.Backend
	backendFactory  func() (backend.Backend, error)
	interrupted     bool
	focusLost       bool
	backendQueue    []any
	backendClosed   bool
}

func New(root Element, options Options) *App {
	app := &App{root: root, options: normalizeOptions(options), work: make(chan func(), 1024)}
	app.dispatcher = &dispatcher{enqueue: func(work func()) { app.work <- work }}
	return app
}

func newWithBackend(root Element, options Options, factory func() (backend.Backend, error)) *App {
	app := New(root, options)
	app.backendFactory = factory
	return app
}

func (app *App) Run(ctx context.Context) (err error) {
	if !app.running.CompareAndSwap(false, true) {
		return errors.New("omnitui: app is already running")
	}
	app.backendClosed = false
	app.backendQueue = nil
	app.interrupted = false
	defer func() {
		app.running.Store(false)
		var closeErr error
		if app.backend != nil {
			closeErr = app.backend.Close()
			app.backend = nil
		}
		if recovered := recover(); recovered != nil {
			panic(recovered)
		}
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()
	var terminal backend.Backend
	var createErr error
	if app.backendFactory != nil {
		terminal, createErr = app.backendFactory()
	} else {
		terminal, createErr = ansi.New(app.options.Input, app.options.Output)
	}
	if createErr != nil {
		return createErr
	}
	app.backend = terminal
	app.width, app.height, err = app.backend.Size()
	if err != nil {
		return err
	}
	app.invalidated = true
	if err = app.cycle(); err != nil {
		return err
	}
	for {
		if len(app.backendQueue) > 0 {
			value := app.backendQueue[0]
			app.backendQueue = app.backendQueue[1:]
			if err = app.handleBackendEvent(value); err != nil {
				return err
			}
			if err = app.cycle(); err != nil {
				return err
			}
			continue
		}
		if app.backendClosed {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case work := <-app.work:
			if work != nil {
				work()
			}
			if err = app.cycle(); err != nil {
				return err
			}
		case event, ok := <-app.backend.Events():
			if !ok {
				app.backendClosed = true
				continue
			}
			value := event.Value
			if resize, ok := value.(backend.ResizeInput); ok {
				draining := true
				for draining {
					select {
					case next, open := <-app.backend.Events():
						if !open {
							app.backendClosed = true
							draining = false
							continue
						}
						if latest, isResize := next.Value.(backend.ResizeInput); isResize {
							resize = latest
						} else {
							app.backendQueue = append(app.backendQueue, next.Value)
							draining = false
						}
					default:
						draining = false
					}
				}
				value = resize
			}
			if err = app.handleBackendEvent(value); err != nil {
				return err
			}
			if err = app.cycle(); err != nil {
				return err
			}
		}
	}
}

func (app *App) UpdateRoot(root Element) {
	app.work <- func() { app.root = root; app.invalidated = true }
}

func (app *App) Dispatch(message any) {
	app.work <- func() { app.pendingMessages = append(app.pendingMessages, message) }
}

func (app *App) cycle() error {
	app.drainWork()
	for len(app.pendingMessages) > 0 {
		message := app.pendingMessages[0]
		app.pendingMessages = app.pendingMessages[1:]
		app.dispatchMessage(message)
		app.drainWork()
	}
	if app.invalidated {
		if err := app.render(); err != nil {
			return err
		}
		app.drainWork()
		if app.invalidated {
			if err := app.render(); err != nil {
				return err
			}
		}
	}
	if app.interrupted {
		return ErrInterrupted
	}
	return nil
}

func (app *App) drainWork() {
	for {
		select {
		case work := <-app.work:
			if work != nil {
				work()
			}
		default:
			return
		}
	}
}

func (app *App) dispatchMessage(message any) {
	if target := rootHost(app.rootInstance); target != nil {
		dispatchEvent(target, "message", MessageEvent{Value: message})
	}
}

func rootHost(root *instance) *instance {
	if root == nil {
		return nil
	}
	if root.kind() == core.KindHost {
		return root
	}
	for _, child := range root.children {
		if result := rootHost(child); result != nil {
			return result
		}
	}
	return nil
}
