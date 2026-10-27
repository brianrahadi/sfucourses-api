package prereq

import (
	"regexp"
	"strings"

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

// Groups returns a prerequisite tree as "one of" groups that must all be met
// (conjunctive normal form), so it is exact for any and/or shape:
//
//	CMPT 225 and (CMPT 295 or ENSC 254)  ->  [[CMPT 225] [CMPT 295 ENSC 254]]
//	(A and B) or C                       ->  [[A C] [B C]]
//
// Unit and permission leaves are omitted ("CMPT 225 or permission" gives
// [[CMPT 225]]). Groups that contain another group are redundant and dropped.
// If a tree would expand past maxGroups, each top-level part is flattened
// into a single group instead.
func Groups(n *model.PrereqNode) [][]string {
	if n == nil {
		return [][]string{}
	}
	if g, ok := cnf(n); ok {
		return absorb(g)
	}
	// Fallback for very large trees: one group per top-level AND child.
	out := [][]string{}
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

const maxGroups = 16

// cnf converts a tree to clauses. A special leaf contributes an empty clause,
// which acts as "false" inside an OR and is removed at the end.
func cnf(n *model.PrereqNode) ([][]string, bool) {
	switch n.Type {
	case "course":
		if strings.Contains(n.ID, ":") {
			return [][]string{{}}, true
		}
		return [][]string{{n.ID}}, true
	case "and":
		var out [][]string
		for i := range n.Children {
			c, ok := cnf(&n.Children[i])
			if !ok {
				return nil, false
			}
			out = append(out, c...)
			if len(out) > maxGroups {
				return nil, false
			}
		}
		return out, true
	case "or":
		out := [][]string{{}}
		for i := range n.Children {
			c, ok := cnf(&n.Children[i])
			if !ok {
				return nil, false
			}
			var next [][]string
			for _, a := range out {
				for _, b := range c {
					next = append(next, unionOrdered(a, b))
				}
			}
			if len(next) > maxGroups {
				return nil, false
			}
			out = next
		}
		return out, true
	}
	return [][]string{}, true
}

func unionOrdered(a, b []string) []string {
	out := append([]string{}, a...)
	for _, x := range b {
		if !contains(out, x) {
			out = append(out, x)
		}
	}
	return out
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

// absorb drops empty groups, duplicates, and any group that is a superset of
// another (if one of {A} is required, one of {A, B} adds nothing).
func absorb(groups [][]string) [][]string {
	out := [][]string{}
	for i, g := range groups {
		if len(g) == 0 {
			continue
		}
		redundant := false
		for j, h := range groups {
			if i == j || len(h) == 0 || len(h) > len(g) {
				continue
			}
			if subset(h, g) && (len(h) < len(g) || j < i) {
				redundant = true
				break
			}
		}
		if !redundant {
			out = append(out, g)
		}
	}
	return out
}

func subset(small, big []string) bool {
	for _, x := range small {
		if !contains(big, x) {
			return false
		}
	}
	return true
}
