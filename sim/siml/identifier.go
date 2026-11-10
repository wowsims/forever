package siml

import "strings"

// The identifier a game name reads as (SPEC.md §7.6): ASCII letters lowercased and digits kept, ' and ’ dropped,
// and every run of other characters one _, with none at either end. "Hunter's Mark" is hunters_mark.
func Identifier(s string) string {
	var b strings.Builder
	separated := false
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
		case r >= 'A' && r <= 'Z':
			r += 'a' - 'A'
		case r == '\'' || r == '’':
			continue
		default:
			separated = true
			continue
		}
		if separated && b.Len() > 0 {
			b.WriteByte('_')
		}
		separated = false
		b.WriteRune(r)
	}
	return b.String()
}

// A CamelCase proto name in snake case (SPEC.md §8.2): _ goes before an upper-case letter that follows a
// lower-case letter or a digit, and before one that follows an upper-case letter and precedes a lower-case one.
// FasterAttackII is faster_attack_ii.
func snake(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isUpper(c) && i > 0 {
			prev := s[i-1]
			if isLower(prev) || isDigit(prev) || isUpper(prev) && i+1 < len(s) && isLower(s[i+1]) {
				b.WriteByte('_')
			}
		}
		if isUpper(c) {
			c += 'a' - 'A'
		}
		b.WriteByte(c)
	}
	return b.String()
}

func isUpper(c byte) bool { return c >= 'A' && c <= 'Z' }
func isLower(c byte) bool { return c >= 'a' && c <= 'z' }
func isDigit(c byte) bool { return c >= '0' && c <= '9' }
