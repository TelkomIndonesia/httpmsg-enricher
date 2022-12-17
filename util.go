package main

import "strings"

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

func MapStringsKeyToLower(m map[string][]string) map[string][]string {
	nm := map[string][]string{}
	for k, v := range m {
		nm[strings.ToLower(k)] = v
	}
	return nm
}
