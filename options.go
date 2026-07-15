package omnitui

import (
	"io"
	"os"
	"strings"
)

type ColorProfile uint8

const (
	ColorProfileAuto ColorProfile = iota
	ColorProfileANSI16
	ColorProfileANSI256
	ColorProfileTrueColor
)

type Options struct {
	Input        io.Reader
	Output       io.Writer
	ColorProfile ColorProfile
}

func normalizeOptions(options Options) Options {
	if options.Input == nil {
		options.Input = os.Stdin
	}
	if options.Output == nil {
		options.Output = os.Stdout
	}
	if options.ColorProfile > ColorProfileTrueColor {
		panic("omnitui: invalid color profile")
	}
	return options
}

func resolveColorProfile(profile ColorProfile) uint8 {
	if profile != ColorProfileAuto {
		return uint8(profile)
	}
	if strings.Contains(strings.ToLower(os.Getenv("COLORTERM")), "truecolor") || strings.Contains(strings.ToLower(os.Getenv("COLORTERM")), "24bit") {
		return uint8(ColorProfileTrueColor)
	}
	if strings.Contains(strings.ToLower(os.Getenv("TERM")), "256color") {
		return uint8(ColorProfileANSI256)
	}
	return uint8(ColorProfileANSI16)
}
