package main

func Truncate(text string, width int) string {
	if width < 0 {
		return ""
	}

	r := []rune(text)
	if width > len(r) {
		return text
	}
	trunc := r[:width]
	return string(trunc) + "..."
}
