package forms

import "strings"

// Completer is the function type for the autocompletion callback of each pair.
type Completer func(text string) [][2]string

// CompleterContains creates a Completer from a list of possible completion
// strings. `caseSens' changes if the matching should be case sensitive.
func CompleterContains(possible []string, caseSens bool) Completer {
	var contains func(s, substr string) bool
	if caseSens {
		contains = strings.Contains
	} else {
		contains = func(s, substr string) bool {
			return strings.Contains(
				strings.ToLower(s),
				strings.ToLower(substr),
			)
		}
	}

	var matches = make([][2]string, 0, len(possible))

	return func(text string) [][2]string {
		matches = matches[:0]

		for _, p := range possible {
			if contains(text, p) {
				matches = append(matches, [2]string{p, ""})
			}
		}

		return matches
	}
}
