package strings

import "strings"

func SplitWords(s string) []string {
	s = strings.Trim(s, " \t\r\n")

	var n int
	var prev bool
	for _, b := range s {
		switch b {
		case ' ', '\t', '\r', '\n':
			if !prev {
				n += 2
				prev = true
			}
		default:
			prev = false
		}
	}
	prev = false

	words := make([]string, 0, n)
	var start int
	for end, b := range s {
		switch b {
		case ' ', '\t', '\r', '\n':
			if !prev {
				words = append(words, s[start:end])
			}
		default:
			if prev {
				start = end
				prev = false
			}
		}
	}

	return words
}
