package text

import "github.com/rivo/uniseg"

func Graphemes(value string) []string {
	if value == "" {
		return nil
	}
	result := make([]string, 0, len(value))
	iterator := uniseg.NewGraphemes(value)
	for iterator.Next() {
		result = append(result, iterator.Str())
	}
	return result
}

func Width(value string) int { return uniseg.StringWidth(value) }
