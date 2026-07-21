//go:build !(aix || darwin || dragonfly || freebsd || linux || netbsd || openbsd || solaris || windows)

package ansi

func (b *Backend) resizeLoop() {}
