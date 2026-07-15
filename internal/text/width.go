package text

func GraphemeWidths(value string) []int {
	graphemes := Graphemes(value)
	widths := make([]int, len(graphemes))
	for index, grapheme := range graphemes {
		widths[index] = Width(grapheme)
	}
	return widths
}

func PrefixByWidth(value string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	used := 0
	result := ""
	for _, grapheme := range Graphemes(value) {
		width := Width(grapheme)
		if width > 0 && used+width > maxWidth {
			break
		}
		result += grapheme
		used += width
	}
	return result
}
