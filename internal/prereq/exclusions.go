package prereq

import (
	"regexp"

	"github.com/brianrahadi/sfucourses-api/internal/model"
)

var (
	// A sentence restricts taking two courses together when it says a student
	// may not/cannot take, receive credit, or enrol ...
	restrictRe = regexp.MustCompile(`(?i)\b(?:may\s+not|must\s+not|cannot|can\s+not|not\s+receive|not\s+be\s+taken)\b`)
	// ... about credit, taking, or concurrency.
	restrictObjRe = regexp.MustCompile(`(?i)\b(?:credit|concurrent(?:ly)?|take|taken|enrol\w*|register\w*)\b`)
)

// Exclusions returns courses that SFU won't give credit for alongside this
// one, or that may not be taken in the same term, as stated in any of texts
// (typically the notes, description and prerequisite fields):
//
//	"Students with credit for BPK 326 may not take this course for further credit."
//	"CRIM 370 and 470 may not be taken concurrently."
//
// Sequencing notes ("before enrolling in ENSC 440") are not exclusions.
func (p *Parser) Exclusions(self string, texts ...string) []string {
	var out []string
	seen := map[string]bool{self: true}
	for _, text := range texts {
		lastDept := ""
		for _, s := range sentences(normalize(text)) {
			if !restrictRe.MatchString(s) || !restrictObjRe.MatchString(s) {
				lastDept = "" // don't carry "CRIM" from an unrelated sentence
				continue
			}
			for _, t := range p.tokenize(s, &lastDept) {
				if t.kind != tCourse || seen[t.val] || !validIDPattern.MatchString(t.val) {
					continue
				}
				seen[t.val] = true
				out = append(out, t.val)
			}
		}
	}
	if out == nil {
		out = []string{}
	}
	return out
}

// Exclusions is the package-level form using permissive department matching.
func Exclusions(self string, texts ...string) []string {
	return defaultParser.Exclusions(self, texts...)
}

// Groups summarises a prerequisite tree as a list of "one of" groups that must
// all be satisfied: AND(CMPT 225, OR(CMPT 295, ENSC 254)) becomes
// [[CMPT 225], [CMPT 295, ENSC 254]]. This is exact for the common shapes
// (a course, an OR list, an AND of courses/OR lists); nested combinations are
// flattened into the nearest group. Unit and permission conditions are omitted.
func Groups(n *model.PrereqNode) [][]string {
	out := [][]string{}
	if n == nil {
		return out
	}
	add := func(m *model.PrereqNode) {
		if g := courseLeaves(m); len(g) > 0 {
			out = append(out, g)
		}
	}
	if n.Type == "and" {
		for i := range n.Children {
			add(&n.Children[i])
		}
		return out
	}
	add(n)
	return out
}
