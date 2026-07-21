//go:build !windows

package ansi

import "os"

type outputState struct{}

func enableVirtualTerminalOutput(*os.File) (*outputState, error) { return nil, nil }

func restoreVirtualTerminalOutput(*os.File, *outputState) error { return nil }
