package ansi

import "testing"

func FuzzParserNeverPanics(f *testing.F) {
	f.Add([]byte("\x1b[<0;1;1M"))
	f.Add([]byte("\x1b[200~paste\x1b[201~"))
	f.Fuzz(func(t *testing.T, input []byte) { parser := &Parser{}; parser.Feed(input); parser.Flush() })
}
