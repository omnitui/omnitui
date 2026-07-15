package text

// Wrap splits a string into terminal-width lines. Newlines are always hard
// breaks. Word wrapping backs up to the last whitespace when possible.
func Wrap(value string, width int, words bool) []string {
	if width <= 0 {
		return []string{""}
	}
	var lines []string
	for _, paragraph := range splitLines(value) {
		if paragraph == "" {
			lines = append(lines, "")
			continue
		}
		remaining := paragraph
		for Width(remaining) > width {
			cut := PrefixByWidth(remaining, width)
			if cut == "" {
				cut = firstGrapheme(remaining)
			}
			if words {
				if index := lastWhitespace(cut); index > 0 {
					cut = cut[:index]
				}
			}
			if cut == "" {
				cut = firstGrapheme(remaining)
			}
			lines = append(lines, cut)
			remaining = trimLeadingSpace(remaining[len(cut):])
		}
		lines = append(lines, remaining)
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func splitLines(value string) []string {
	result := []string{""}
	for _, runeValue := range value {
		if runeValue == '\n' {
			result = append(result, "")
		} else {
			result[len(result)-1] += string(runeValue)
		}
	}
	return result
}

func firstGrapheme(value string) string {
	graphemes := Graphemes(value)
	if len(graphemes) == 0 {
		return ""
	}
	return graphemes[0]
}

func lastWhitespace(value string) int {
	last := -1
	for index, runeValue := range value {
		if runeValue == ' ' || runeValue == '\t' {
			last = index
		}
	}
	return last
}

func trimLeadingSpace(value string) string {
	for len(value) > 0 && (value[0] == ' ' || value[0] == '\t') {
		value = value[1:]
	}
	return value
}
