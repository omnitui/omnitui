package headless

import "testing"

func TestBackendRecordsFramesAndEvents(t *testing.T) {
	backend := New(10, 3)
	backend.Send("input")
	if got := (<-backend.Events()).Value; got != "input" {
		t.Fatalf("event = %#v", got)
	}
	if err := backend.Write([]byte("frame")); err != nil {
		t.Fatal(err)
	}
	frames := backend.Frames()
	if len(frames) != 1 || string(frames[0]) != "frame" {
		t.Fatalf("frames = %#v", frames)
	}
	if err := backend.Close(); err != nil {
		t.Fatal(err)
	}
}
