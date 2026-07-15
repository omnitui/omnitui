package text

func Truncate(value string, width int, ellipsis bool) string {
	if Width(value) <= width {
		return value
	}
	if width <= 0 {
		return ""
	}
	marker := ""
	if ellipsis {
		marker = "…"
	}
	if Width(marker) >= width {
		return PrefixByWidth(marker, width)
	}
	return PrefixByWidth(value, width-Width(marker)) + marker
}
