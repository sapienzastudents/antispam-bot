package antispam

import (
	"strings"
	"unicode"
)

// ChineseChars calculate the percent of the string that is in chinese (Han) chars (unicode).
//
// Time complexity: O(n) where "n" is the number of runes in a string
func ChineseChars(str string) float64 {
	// Base: if the string is empty
	if len(str) == 0 || strings.TrimSpace(str) == "" {
		return 0
	}

	// Count chinese runes in string
	badchars := 0
	totalchars := 0
	// Note that "totalchars" != len(str), so we need to count runes "manually" using totalchars
	// The len() function returns the length in byte, but chars might be multi-byte
	// In fact, Go uses the type "rune" which is more robust (even if there are some "corner cases")
	for _, runeValue := range str {
		if unicode.Is(unicode.Han, runeValue) {
			// Chinese character
			badchars++
		}
		totalchars++
	}

	return float64(badchars) / float64(totalchars)
}
