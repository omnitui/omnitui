//go:build windows

package ansi

import (
	"os"

	"golang.org/x/sys/windows"
)

type outputState struct{ mode uint32 }

func enableVirtualTerminalOutput(file *os.File) (*outputState, error) {
	handle := windows.Handle(file.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(handle, &mode); err != nil {
		return nil, err
	}
	updated := mode | windows.ENABLE_PROCESSED_OUTPUT | windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING
	if err := windows.SetConsoleMode(handle, updated); err != nil {
		return nil, err
	}
	return &outputState{mode: mode}, nil
}

func restoreVirtualTerminalOutput(file *os.File, state *outputState) error {
	if file == nil || state == nil {
		return nil
	}
	return windows.SetConsoleMode(windows.Handle(file.Fd()), state.mode)
}
