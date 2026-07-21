//go:build aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris

package ansi

import (
	"os"
	"os/signal"
	"syscall"
)

func (b *Backend) resizeLoop() {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGWINCH)
	defer signal.Stop(signals)
	for {
		select {
		case <-b.done:
			return
		case <-signals:
			if !b.emitResize() {
				return
			}
		}
	}
}
