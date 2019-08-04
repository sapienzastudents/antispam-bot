package main

import (
	"strings"
	"unicode"
)

func arabicChars(str string) float64 {
	if len(str) == 0 || strings.TrimSpace(str) == "" {
		return 0
	}

	badchars := 0
	totalchars := 0
	// Note that "totalchars" != len(str)
	// The len() function returns the length in byte, but chars might be multi-byte
	// In fact, GoLang uses the type "rune" which is more robust (even if there are some "corner cases")
	for _, runeValue := range str {
		if unicode.Is(unicode.Arabic, runeValue) {
			// Chinese character
			badchars += 1
		}
		totalchars += 1
	}

	return float64(badchars) / float64(totalchars)
}
