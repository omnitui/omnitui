package ansi

import (
	"bytes"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/viniciusfonseca/omnitui/internal/backend"
)

type Parser struct {
	pending   []byte
	paste     bool
	pasteData []byte
	buttons   uint8
}

// Flush resolves an isolated escape byte at end of input.
func (p *Parser) Flush() []backend.Event {
	if len(p.pending) == 1 && p.pending[0] == 0x1b {
		p.pending = nil
		return []backend.Event{{Value: backend.KeyInput{Key: 2}}}
	}
	return nil
}

func (p *Parser) Feed(data []byte) []backend.Event {
	p.pending = append(p.pending, data...)
	var events []backend.Event
	for len(p.pending) > 0 {
		if p.paste {
			end := bytes.Index(p.pending, []byte("\x1b[201~"))
			if end < 0 {
				p.pasteData = append(p.pasteData, p.pending...)
				p.pending = nil
				break
			}
			p.pasteData = append(p.pasteData, p.pending[:end]...)
			events = append(events, backend.Event{Value: backend.PasteInput{Text: string(p.pasteData)}})
			p.pasteData = nil
			p.paste = false
			p.pending = p.pending[end+len("\x1b[201~"):]
			continue
		}
		if p.pending[0] == 0x1b {
			if len(p.pending) == 1 {
				break
			}
			if p.pending[1] != '[' {
				if p.pending[1] == ']' {
					break
				}
				if p.pending[1] >= 0x20 && p.pending[1] != 0x7f {
					r, size := utf8.DecodeRune(p.pending[1:])
					if r != utf8.RuneError || size > 1 {
						events = append(events, key(r, 2))
						p.pending = p.pending[1+size:]
						continue
					}
				}
				events = append(events, backend.Event{Value: backend.KeyInput{Key: 2}})
				p.pending = p.pending[1:]
				continue
			}
			if len(p.pending) >= 6 && string(p.pending[:6]) == "\x1b[200~" {
				p.paste = true
				p.pending = p.pending[6:]
				continue
			}
			end := sequenceEnd(p.pending[2:])
			if end < 0 {
				break
			}
			sequence := string(p.pending[2 : 2+end+1])
			p.pending = p.pending[2+end+1:]
			parsed := parseCSI(sequence)
			for index := range parsed {
				if mouse, ok := parsed[index].Value.(backend.MouseInput); ok {
					if mouse.Button != 0 {
						bit := uint8(1 << (mouse.Button - 1))
						if mouse.Action == 1 {
							p.buttons |= bit
						}
						if mouse.Action == 2 {
							p.buttons &^= bit
						}
					}
					mouse.Buttons = p.buttons
					parsed[index].Value = mouse
				}
			}
			events = append(events, parsed...)
			continue
		}
		if p.pending[0] < 0x20 || p.pending[0] == 0x7f {
			switch p.pending[0] {
			case 0x03:
				events = append(events, backend.Event{Value: backend.KeyInput{Key: 0, Modifiers: 1}})
			case '\r', '\n':
				events = append(events, backend.Event{Value: backend.KeyInput{Key: 1}})
			case '\t':
				events = append(events, backend.Event{Value: backend.KeyInput{Key: 3}})
			case 0x7f:
				events = append(events, backend.Event{Value: backend.KeyInput{Key: 5}})
			}
			p.pending = p.pending[1:]
			continue
		}
		r, size := utf8.DecodeRune(p.pending)
		if r == utf8.RuneError && size == 1 {
			break
		}
		events = append(events, key(r, 0))
		p.pending = p.pending[size:]
	}
	return events
}

func sequenceEnd(value []byte) int {
	for index, current := range value {
		if current >= 0x40 && current <= 0x7e {
			return index
		}
	}
	return -1
}

func parseCSI(sequence string) []backend.Event {
	if strings.HasPrefix(sequence, "<") {
		if event, ok := parseMouse(sequence[1:]); ok {
			return []backend.Event{{Value: event}}
		}
		return nil
	}
	final := sequence[len(sequence)-1]
	body := sequence[:len(sequence)-1]
	params := strings.Split(body, ";")
	first := 0
	modifier := uint8(0)
	if len(params) > 0 && params[0] != "" {
		first, _ = strconv.Atoi(params[0])
	}
	if len(params) > 1 {
		if value, err := strconv.Atoi(params[1]); err == nil {
			modifier = csiModifiers(value)
		}
	}
	var keyValue uint16
	switch final {
	case 'A':
		keyValue = 8
	case 'B':
		keyValue = 9
	case 'C':
		keyValue = 11
	case 'D':
		keyValue = 10
	case 'H':
		keyValue = 12
	case 'F':
		keyValue = 13
	case 'Z':
		keyValue = 4
	case '~':
		switch first {
		case 1, 7:
			keyValue = 12
		case 2:
			keyValue = 7
		case 3:
			keyValue = 6
		case 4, 8:
			keyValue = 13
		case 5:
			keyValue = 14
		case 6:
			keyValue = 15
		case 15:
			keyValue = 16
		case 17:
			keyValue = 17
		case 18:
			keyValue = 18
		case 19:
			keyValue = 19
		case 20:
			keyValue = 20
		case 21:
			keyValue = 21
		case 23:
			keyValue = 22
		case 24:
			keyValue = 23
		}
	default:
		return nil
	}
	if keyValue == 0 {
		return nil
	}
	return []backend.Event{{Value: backend.KeyInput{Key: keyValue, Modifiers: modifier}}}
}

func csiModifiers(value int) uint8 {
	switch value {
	case 2:
		return 4
	case 3, 4:
		return 2
	case 5, 6:
		return 6
	case 7, 8:
		return 3
	default:
		return 0
	}
}
func key(r rune, modifiers uint8) backend.Event {
	return backend.Event{Value: backend.KeyInput{Key: 0, Rune: r, Modifiers: modifiers}}
}
