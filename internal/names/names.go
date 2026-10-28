// Package names cleans up instructor names as they appear in SFU's section
// data, which sometimes carries placeholders ("TBA", "Sessional", ". ."),
// roles ("Erin Barley (Lab Instructor)", "Megan Barker, Lab Instructor"),
// several people in one entry ("Cary Campbell & Michael Ling"), or a
// repeated surname ("Baran (Abu) Fakhri Fakhri").
package names

import (
	"regexp"
	"strings"
)

// Words that, on their own, mean "no instructor assigned yet".
var placeholderWords = map[string]bool{
	"tba": true, "tbd": true, "sessional": true, "faculty": true,
	"staff": true, "instructor": true, "instructors": true,
}

// Words that describe a role rather than a person.
var roleWords = regexp.MustCompile(`(?i)\b(lab|labs|lecture|lectures|lecturer|instructor|tutorial|tutorials|seminar|ta|coordinator)\b`)

// An all-caps parenthetical like "(UBC)" is an affiliation, not part of a name.
var affiliation = regexp.MustCompile(`^[A-Z]{2,6}$`)

var parenthetical = regexp.MustCompile(`\s*\(([^)]*)\)`)
var whitespace = regexp.MustCompile(`\s+`)
var letters = regexp.MustCompile(`\p{L}{2,}`)
var separators = regexp.MustCompile(`\s*[&;/]\s*`)

// IsPlaceholder reports whether name stands for "not assigned yet" rather
// than a person.
func IsPlaceholder(name string) bool {
	words := strings.FieldsFunc(strings.ToLower(name), func(r rune) bool {
		return r == ' ' || r == '.' || r == ',' || r == '-' || r == '(' || r == ')'
	})
	if len(words) == 0 {
		return true
	}
	for _, w := range words {
		if !placeholderWords[w] {
			return !letters.MatchString(name)
		}
	}
	return true
}

// cleanOne removes roles and affiliations from a single person's name.
func cleanOne(name string) string {
	name = parenthetical.ReplaceAllStringFunc(name, func(m string) string {
		inner := strings.TrimSpace(parenthetical.FindStringSubmatch(m)[1])
		if inner == "" || roleWords.MatchString(inner) || affiliation.MatchString(inner) {
			return ""
		}
		return m // e.g. a nickname: "Baran (Abu) Fakhri"
	})

	// "Megan Barker, Lab Instructor" → "Megan Barker"; "Faculty ," → "Faculty"
	if i := strings.Index(name, ","); i >= 0 {
		rest := strings.TrimSpace(name[i+1:])
		if rest == "" || roleWords.MatchString(rest) {
			name = name[:i]
		}
	}

	words := strings.Fields(whitespace.ReplaceAllString(name, " "))
	// "Baran (Abu) Fakhri Fakhri" → "Baran (Abu) Fakhri". Only a repeated
	// trailing surname on a longer name: "Yan Yan" and "Man Man Wong" are
	// real names.
	nameWords := 0
	for _, w := range words {
		if !strings.HasPrefix(w, "(") {
			nameWords++
		}
	}
	if n := len(words); nameWords >= 3 && strings.EqualFold(words[n-1], words[n-2]) {
		words = words[:n-1]
	}
	return strings.Join(words, " ")
}

// Clean turns one raw instructor entry into zero or more person names:
// placeholders are dropped and entries naming several people are split.
func Clean(raw string) []string {
	var out []string
	for _, part := range separators.Split(raw, -1) {
		name := cleanOne(strings.TrimSpace(part))
		if name == "" || IsPlaceholder(name) {
			continue
		}
		out = append(out, name)
	}
	return out
}

// CleanAll cleans a list of raw names, keeping order and removing duplicates.
func CleanAll(raw []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, r := range raw {
		for _, name := range Clean(r) {
			if !seen[name] {
				seen[name] = true
				out = append(out, name)
			}
		}
	}
	return out
}
