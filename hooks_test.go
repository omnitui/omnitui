package omnitui

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"testing"

	"github.com/omnitui/omnitui/internal/backend"
	"github.com/omnitui/omnitui/internal/backend/headless"
	"github.com/omnitui/omnitui/internal/core"
)

type effectProps struct {
	Dependency int
	Enabled    bool
}

type duplicateEffectProbe struct{}

func (duplicateEffectProbe) InitialState(struct{}) struct{} { return struct{}{} }
func (duplicateEffectProbe) Render(ctx Context, _ struct{}, _ struct{}, _ Children) Element {
	UseEffect(ctx, "same", 1, func(context.Context) Cleanup { return nil })
	UseEffect(ctx, "same", 1, func(context.Context) Cleanup { return nil })
	return None()
}

func TestDuplicateHookKeyPanics(t *testing.T) {
	typeValue := Define[struct{}, struct{}]("DuplicateEffectProbe", duplicateEffectProbe{})
	app := New(Create(typeValue, struct{}{}), Options{})
	app.width, app.height = 1, 1
	defer func() {
		if recover() == nil {
			t.Fatal("expected duplicate hook key panic")
		}
	}()
	_ = app.render()
}

type effectProbe struct{ log *[]string }

func (effectProbe) InitialState(effectProps) struct{} { return struct{}{} }
func (probe effectProbe) Render(ctx Context, props effectProps, _ struct{}, _ Children) Element {
	*probe.log = append(*probe.log, fmt.Sprintf("render:%d", props.Dependency))
	if props.Enabled {
		UseEffect(ctx, "probe", props.Dependency, func(effectContext context.Context) Cleanup {
			*probe.log = append(*probe.log, fmt.Sprintf("setup:%d", props.Dependency))
			return func() {
				canceled := false
				select {
				case <-effectContext.Done():
					canceled = true
				default:
				}
				*probe.log = append(*probe.log, fmt.Sprintf("cleanup:%d:%t", props.Dependency, canceled))
			}
		})
	}
	return None()
}

func TestUseEffectLifecycle(t *testing.T) {
	log := []string{}
	typeValue := Define[effectProps, struct{}]("EffectProbe", effectProbe{log: &log})
	app := New(Create(typeValue, effectProps{Dependency: 1, Enabled: true}), Options{})
	app.width, app.height = 10, 2
	render := func(props effectProps) {
		t.Helper()
		app.root = Create(typeValue, props)
		app.invalidated = true
		if err := app.render(); err != nil {
			t.Fatal(err)
		}
	}
	render(effectProps{Dependency: 1, Enabled: true})
	render(effectProps{Dependency: 1, Enabled: true})
	render(effectProps{Dependency: 2, Enabled: true})
	render(effectProps{Dependency: 2})
	render(effectProps{Dependency: 3, Enabled: true})
	app.root = None()
	app.invalidated = true
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"render:1", "setup:1",
		"render:1",
		"render:2", "cleanup:1:true", "setup:2",
		"render:2", "cleanup:2:true",
		"render:3", "setup:3", "cleanup:3:true",
	}
	if !reflect.DeepEqual(log, want) {
		t.Fatalf("effect log = %#v, want %#v", log, want)
	}
}

func TestEffectsCleanUpWhenRunReturns(t *testing.T) {
	log := []string{}
	typeValue := Define[effectProps, struct{}]("EffectProbe", effectProbe{log: &log})
	headlessBackend := headless.New(8, 2)
	headlessBackend.Send(backend.KeyInput{Modifiers: 1})
	app := newWithBackend(Create(typeValue, effectProps{Dependency: 1, Enabled: true}), Options{}, func() (backend.Backend, error) {
		return headlessBackend, nil
	})
	if err := app.Run(context.Background()); err != ErrInterrupted {
		t.Fatalf("Run() error = %v, want ErrInterrupted", err)
	}
	want := []string{"render:1", "setup:1", "cleanup:1:true"}
	if !reflect.DeepEqual(log, want) {
		t.Fatalf("effect log = %#v, want %#v", log, want)
	}
}

type effectStateProbe struct{ seen *[]int }

func (effectStateProbe) InitialState(struct{}) int { return 0 }
func (probe effectStateProbe) Render(ctx Context, _ struct{}, state int, _ Children) Element {
	*probe.seen = append(*probe.seen, state)
	UseEffect(ctx, "initialize", struct{}{}, func(context.Context) Cleanup {
		SetState(ctx, 1)
		return nil
	})
	return None()
}

func TestEffectStateUpdateSchedulesAnotherRender(t *testing.T) {
	seen := []int{}
	typeValue := Define[struct{}, int]("EffectStateProbe", effectStateProbe{seen: &seen})
	app := New(Create(typeValue, struct{}{}), Options{})
	app.width, app.height = 1, 1
	app.invalidated = true
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(seen, []int{0, 1}) {
		t.Fatalf("rendered states = %v, want [0 1]", seen)
	}
}

type concurrentEffectProbe struct {
	entered chan<- struct{}
	release <-chan struct{}
	trigger <-chan struct{}
	result  chan<- any
}

func (concurrentEffectProbe) InitialState(bool) int { return 0 }
func (probe concurrentEffectProbe) Render(ctx Context, block bool, _ int, _ Children) Element {
	UseEffect(ctx, "worker", struct{}{}, func(context.Context) Cleanup {
		go func() {
			<-probe.trigger
			var recovered any
			func() {
				defer func() { recovered = recover() }()
				SetState(ctx, 1)
			}()
			probe.result <- recovered
		}()
		return nil
	})
	if block {
		probe.entered <- struct{}{}
		<-probe.release
	}
	return None()
}

func TestEffectCanUpdateStateDuringLaterRender(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	trigger := make(chan struct{})
	result := make(chan any)
	typeValue := Define[bool, int]("ConcurrentEffectProbe", concurrentEffectProbe{
		entered: entered,
		release: release,
		trigger: trigger,
		result:  result,
	})
	app := New(Create(typeValue, false), Options{})
	app.width, app.height = 1, 1
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	app.root = Create(typeValue, true)
	app.invalidated = true
	rendered := make(chan error)
	go func() { rendered <- app.render() }()
	<-entered
	close(trigger)
	if recovered := <-result; recovered != nil {
		t.Fatalf("asynchronous state update panicked during a later render: %v", recovered)
	}
	close(release)
	if err := <-rendered; err != nil {
		t.Fatal(err)
	}
	app.root = None()
	app.invalidated = true
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
}

type refProbe struct{ seen *[]*Ref[int] }

func (refProbe) InitialState(int) struct{} { return struct{}{} }
func (probe refProbe) Render(ctx Context, initial int, _ struct{}, _ Children) Element {
	*probe.seen = append(*probe.seen, UseRef(ctx, "counter", initial))
	return None()
}

func TestUseRefPreservesValueWithoutRendering(t *testing.T) {
	seen := []*Ref[int]{}
	typeValue := Define[int, struct{}]("RefProbe", refProbe{seen: &seen})
	app := New(Create(typeValue, 1), Options{})
	app.width, app.height = 1, 1
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	seen[0].Set(7)
	app.root = Create(typeValue, 99)
	app.invalidated = true
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	if seen[0] != seen[1] || seen[1].Get() != 7 {
		t.Fatalf("ref was not preserved: first=%p second=%p value=%d", seen[0], seen[1], seen[1].Get())
	}
}

func TestRefIsSafeForConcurrentUpdates(t *testing.T) {
	ref := &Ref[int]{}
	var wait sync.WaitGroup
	for range 20 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 50 {
				ref.Update(func(value int) int { return value + 1 })
			}
		}()
	}
	wait.Wait()
	if got := ref.Get(); got != 1000 {
		t.Fatalf("ref value = %d, want 1000", got)
	}
}

func TestRefUpdatePanicDoesNotPoisonRef(t *testing.T) {
	ref := &Ref[int]{value: 1}
	func() {
		defer func() { _ = recover() }()
		ref.Update(func(int) int { panic("update failed") })
	}()
	ref.Set(2)
	if got := ref.Get(); got != 2 {
		t.Fatalf("ref value = %d, want 2", got)
	}
}

type viewportProbe struct{ seen *[]Viewport }

func (viewportProbe) InitialState(struct{}) struct{} { return struct{}{} }
func (probe viewportProbe) Render(ctx Context, _ struct{}, _ struct{}, _ Children) Element {
	*probe.seen = append(*probe.seen, UseViewport(ctx))
	return None()
}

func TestUseViewportTracksResize(t *testing.T) {
	seen := []Viewport{}
	typeValue := Define[struct{}, struct{}]("ViewportProbe", viewportProbe{seen: &seen})
	app := New(Create(typeValue, struct{}{}), Options{})
	app.width, app.height = 80, 24
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	if err := app.handleBackendEvent(ResizeEvent{Width: 40, Height: 12}); err != nil {
		t.Fatal(err)
	}
	if err := app.render(); err != nil {
		t.Fatal(err)
	}
	want := []Viewport{{Width: 80, Height: 24}, {Width: 40, Height: 12}}
	if !reflect.DeepEqual(seen, want) {
		t.Fatalf("viewports = %#v, want %#v", seen, want)
	}
}

type focusProbe struct{ handles *[]FocusHandle }

func (focusProbe) InitialState(struct{}) struct{} { return struct{}{} }
func (probe focusProbe) Render(ctx Context, _ struct{}, _ struct{}, _ Children) Element {
	handle := UseFocus(ctx, "input")
	*probe.handles = append(*probe.handles, handle)
	return core.NewHost(core.HostInput, core.InputData{Focus: handle}, nil)
}

func TestUseFocusRequestsAndReleasesFocus(t *testing.T) {
	handles := []FocusHandle{}
	typeValue := Define[struct{}, struct{}]("FocusProbe", focusProbe{handles: &handles})
	app := New(Create(typeValue, struct{}{}), Options{})
	app.width, app.height = 10, 2
	app.invalidated = true
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	handle := handles[len(handles)-1]
	handle.Request()
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	if !handle.Focused() || app.focused == nil || app.focused.host.Kind != core.HostInput {
		t.Fatal("focus request did not focus the bound input")
	}
	handle.Blur()
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	if handle.Focused() || app.focused != nil {
		t.Fatal("blur did not release focus")
	}
	app.root = None()
	app.invalidated = true
	if err := app.cycle(); err != nil {
		t.Fatal(err)
	}
	handle.Request()
	if app.focused != nil {
		t.Fatal("detached focus handle changed focus")
	}
}
