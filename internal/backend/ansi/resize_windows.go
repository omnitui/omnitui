//go:build windows

package ansi

import "time"

// Windows consoles do not deliver SIGWINCH. Polling the console dimensions
// keeps resize behavior consistent with Unix terminals without a native input
// event loop.
func (b *Backend) resizeLoop() {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-b.done:
			return
		case <-ticker.C:
			if !b.emitResize() {
				return
			}
		}
	}
}
