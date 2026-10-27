// Package prereq turns SFU's free-text prerequisite/corequisite strings into
// structured requirement trees and a flat course graph.
//
// The parser is deliberately tolerant: catalog text mixes course codes with
// prose ("Completion of", "with a minimum grade of C-", "or permission of the
// instructor"), so anything that is not a course, a unit count, a permission
// clause or a logical connector is skipped instead of aborting the parse.
//
// Pipeline:
//  1. splitSections: separate "Corequisite:" / "Recommended:" / "Note:" parts.
//  2. sentences: split on sentence boundaries, dropping sentences that state
//     exclusions ("may not take ... for further credit") or recommendations.
//  3. tokenize: emit course / special / connector tokens, expanding shorthand
//     like "MATH 150, 151", "CA (or FPA) 160" and "EDUC 401/402".
//  4. parse: build an and/or tree using SFU's list conventions
//     ("A, B or C" is one OR list; "A, B, and C" one AND list; "one of ..."
//     makes a list OR) with precedence  or  >  and  >  ", or"  >  ", and"  >  ";".
package prereq

import (
	"regexp"
	"strings"

	"github.com/brianrahadi/sfucourses-api/internal/model"
)

// ---------------------------------------------------------------------------
// Tokens

type tokKind int

const (
	tCourse tokKind = iota
	tSpecial
	tLParen
	tRParen
	tAnd
	tOr
	tComma
	tSemi
	tSlash
	tOneOf
)

type token struct {
	kind tokKind
	val  string
}

// ---------------------------------------------------------------------------
// Parser

// Parser parses requirement text. Any upper-case 2-5 letter code followed by a
// course number is accepted (SFU text references retired subjects like FPA,
// FNST or HS that are no longer in the catalog). Department codes passed to New
// additionally match when mis-cased in the source ("Math 232", "Bus 792").
type Parser struct {
	known map[string]bool
}

// New returns a Parser that also recognises mis-cased spellings of depts.
func New(depts []string) *Parser {
	p := &Parser{known: make(map[string]bool, len(depts))}
	for _, d := range depts {
		p.known[strings.ToUpper(strings.TrimSpace(d))] = true
	}
	return p
}

var defaultParser = &Parser{}

// Requirements is the structured result for one course.
type Requirements struct {
	Prereq *model.PrereqNode // courses/conditions needed beforehand
	Coreq  *model.PrereqNode // courses that may (or must) be taken concurrently
}

// Parse converts a raw prerequisite string into a PrereqNode tree.
// Corequisites embedded in the prerequisite text ("..., Corequisite: STAT 285")
// are AND-ed in as "COREQUISITE:<code>" leaves, matching the API's
// historical response shape. Returns nil when nothing structured is found.
func Parse(raw string) *model.PrereqNode { return defaultParser.Parse(raw) }

// Parse is the method form of the package-level Parse.
func (p *Parser) Parse(raw string) *model.PrereqNode {
	req := p.ParseRequirements(raw, "")
	return combineWithCoreqs(req.Prereq, req.Coreq)
}

// ParseRequirements parses a course's prerequisite and corequisite fields,
// routing any corequisite clause found inside the prerequisite text into Coreq.
func (p *Parser) ParseRequirements(prereqText, coreqText string) Requirements {
	main, embeddedCoreq := splitSections(prereqText)
	coreqMain, coreqExtra := splitSections(coreqText)

	pre, conc := p.parseSectionSplit(main)
	return Requirements{
		Prereq: pre,
		Coreq:  joinAnd(conc, p.parseSection(embeddedCoreq), p.parseSection(coreqMain), p.parseSection(coreqExtra)),
	}
}

func combineWithCoreqs(pre, co *model.PrereqNode) *model.PrereqNode {
	if co == nil {
		return pre
	}
	co = prefixLeaves(co, "COREQUISITE:")
	return joinAnd(pre, co)
}

// parseSection parses one section (sentences are AND-ed together).
func (p *Parser) parseSection(text string) *model.PrereqNode {
	pre, _ := p.parseSectionSplit(text)
	return pre
}

// parseSectionSplit also returns courses from "taken concurrently with" sentences.
func (p *Parser) parseSectionSplit(text string) (pre, concurrent *model.PrereqNode) {
	text = normalize(text)
	if text == "" {
		return nil, nil
	}
	var parts, conc []*model.PrereqNode
	type substitution struct {
		alt     *model.PrereqNode
		targets []string
	}
	var subs []substitution
	lastDept := ""
	for _, s := range sentences(text) {
		s = dropClauses(s)
		if s == "" {
			continue
		}
		if loc := substitutedRe.FindStringIndex(s); loc != nil {
			alt := finalize(parseTokens(p.tokenize(s[:loc[0]], &lastDept)))
			targets := Leaves(finalize(parseTokens(p.tokenize(s[loc[1]:], &lastDept))))
			if alt != nil && len(targets) > 0 {
				subs = append(subs, substitution{alt, targets})
			}
			continue
		}
		toks := p.tokenize(s, &lastDept)
		n := parseTokens(toks)
		if n == nil {
			continue
		}
		if concurrentRe.MatchString(s) {
			conc = append(conc, n)
		} else {
			parts = append(parts, n)
		}
	}
	pre = finalize(joinAnd(parts...))
	for _, sub := range subs {
		pre = finalize(substitute(pre, sub.targets, sub.alt))
	}
	return pre, finalize(joinAnd(conc...))
}

// substitute makes alt an alternative to the smallest subtree whose leaves are
// exactly targets ("... may be substituted for MATH 150 or MATH 151").
// If no such subtree exists the tree is returned unchanged.
func substitute(n *model.PrereqNode, targets []string, alt *model.PrereqNode) *model.PrereqNode {
	if n == nil {
		return nil
	}
	if sameSet(Leaves(n), targets) {
		return orNode(n, alt)
	}
	if n.Type == "course" {
		return n
	}
	c := *n
	c.Children = make([]model.PrereqNode, len(n.Children))
	for i := range n.Children {
		c.Children[i] = *substitute(&n.Children[i], targets, alt)
	}
	return &c
}

func sameSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	m := make(map[string]bool, len(a))
	for _, x := range a {
		m[x] = true
	}
	for _, x := range b {
		if !m[x] {
			return false
		}
	}
	return true
}

// dropClauses removes ';'-separated clauses that are recommendations or
// exclusions ("MBB 308 or 309W; Students may not take MBB 498 concurrently ...").
func dropClauses(s string) string {
	if !exclusionRe.MatchString(s) && !recommendRe.MatchString(s) {
		return s
	}
	var kept []string
	for _, c := range strings.Split(s, ";") {
		if exclusionRe.MatchString(c) || recommendRe.MatchString(c) {
			continue
		}
		kept = append(kept, c)
	}
	return strings.TrimSpace(strings.Join(kept, ";"))
}

// ---------------------------------------------------------------------------
// Pre-processing

var (
	// Section markers. Text after a marker belongs to that section until the next marker.
	// "Prerequisite/Corequisite:" and "Prerequisite or corequisite:" count as
	// prerequisites (they may be completed beforehand); "Recommended
	// Prerequisite:" is a recommendation.
	sectionRe = regexp.MustCompile(`(?i)\b((?:strongly\s+|highly\s+)?recommended(?:\s+(?:pre|co)-?req\w*)?|pre-?req\w*(?:\s*(?:/|or)\s*co-?req\w*)?|co-?req\w*|notes?)\s*:`)
	// Inline corequisite phrase without a colon ("or co-requisite EDUC 403").
	inlineCoreqRe = regexp.MustCompile(`(?i)\bco-?requisites?\b`)

	exclusionRe = regexp.MustCompile(`(?i)\b(?:may\s+not|must\s+not|not\s+be\s+taken|cannot|can\s+not|will\s+not|not\s+be\s+eligible|not\s+receive|for\s+further\s+credit|who\s+have\s+(?:taken|credit|received|completed)|before\s+enrolling|prior\s+to\s+enrolling|immediately\s+following|not\s+open\s+to)\b`)
	recommendRe = regexp.MustCompile(`(?i)\brecommend`)
	// "MATH 154 or 157 may be substituted for MATH 150 or 151."
	substitutedRe = regexp.MustCompile(`(?i)\b(?:may|can)\s+be\s+(?:substituted|used|taken)\s+(?:for|in\s+place\s+of|instead\s+of)\b`)
	// "Normally taken concurrently with ARCH 433" describes corequisites.
	concurrentRe = regexp.MustCompile(`(?i)\bconcurrently\s+with\b`)

	gradeRe     = regexp.MustCompile(`,?\s*(?i:(?:all|both|each)\s+)?(?i:(?:with|having|earning|obtaining)\s+)?(?i:an?\s+)?(?i:(?:minimum|min\.?)\s+)?(?i:letter\s+)?(?i:grades?\s+)(?i:of\s+)?(?i:at\s+least\s+)?(?i:an?\s+)?[A-D](?:[+-]|\b)(?i:\s+or\s+(?:better|higher|above|greater))?`)
	gradeTailRe = regexp.MustCompile(`\b[A-D][+-]?\s+(?i:or\s+(?:better|higher|above))`)
	andOrRe     = regexp.MustCompile(`(?i)\band\s*/\s*or\b`)
	// "(STAT 240 is a suitable substitute for STAT 341)" names alternatives.
	substituteRe = regexp.MustCompile(`\(\s*([^()]*\b(?i:substitute|in\s+place\s+of|instead\s+of|accepted\s+in\s+lieu)\b[^()]*)\)`)
	spaceRe      = regexp.MustCompile(`\s+`)
)

func normalize(s string) string {
	r := strings.NewReplacer(
		"\u00a0", " ", "\u2013", "-", "\u2014", "-", "\u2019", "'",
		"\u201c", "\"", "\u201d", "\"", "\r", " ", "\n", " ", "\t", " ",
	)
	s = r.Replace(s)
	s = andOrRe.ReplaceAllString(s, "or")
	s = substituteRe.ReplaceAllString(s, "(or $1)")
	s = gradeRe.ReplaceAllString(s, "")
	s = gradeTailRe.ReplaceAllString(s, "")
	s = spaceRe.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// splitSections returns (prerequisite part, corequisite part). Recommended and
// note sections are dropped.
func splitSections(s string) (main, coreq string) {
	if strings.TrimSpace(s) == "" {
		return "", ""
	}
	locs := sectionRe.FindAllStringSubmatchIndex(s, -1)
	if len(locs) == 0 {
		return splitInlineCoreq(s)
	}
	var mainB, coB strings.Builder
	mainB.WriteString(s[:locs[0][0]])
	for i, loc := range locs {
		end := len(s)
		if i+1 < len(locs) {
			end = locs[i+1][0]
		}
		body := s[loc[1]:end]
		label := strings.ToLower(s[loc[2]:loc[3]])
		if strings.HasPrefix(label, "pre") {
			mainB.WriteString(" ")
			mainB.WriteString(body)
			continue
		}
		// Corequisite / recommended / note sections run to the end of their
		// sentence; anything after goes back to the prerequisite text.
		first, rest := body, ""
		if ss := sentences(strings.TrimSpace(body)); len(ss) > 1 {
			first = ss[0]
			rest = strings.Join(ss[1:], ". ")
		}
		if strings.HasPrefix(label, "co") {
			coB.WriteString(" ")
			coB.WriteString(first)
			coB.WriteString(".")
		}
		if rest != "" {
			mainB.WriteString(". ")
			mainB.WriteString(rest)
		}
	}
	m, c := splitInlineCoreq(mainB.String())
	return m, strings.TrimSpace(c + " " + coB.String())
}

// splitInlineCoreq handles "...; or co-requisite EDUC 403." style phrases.
func splitInlineCoreq(s string) (string, string) {
	loc := inlineCoreqRe.FindStringIndex(s)
	if loc == nil {
		return s, ""
	}
	return s[:loc[0]], s[loc[1]:]
}

// sentences splits on ". " followed by an upper-case letter or "(".
func sentences(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] != '.' {
			continue
		}
		j := i + 1
		for j < len(s) && s[j] == ' ' {
			j++
		}
		if j == len(s) || (j > i+1 && (isUpper(s[j]) || s[j] == '(')) {
			out = append(out, strings.TrimSpace(s[start:i]))
			start = j
			i = j - 1
		}
	}
	if start < len(s) {
		out = append(out, strings.TrimSpace(s[start:]))
	}
	return out
}

func isUpper(b byte) bool { return b >= 'A' && b <= 'Z' }

// ---------------------------------------------------------------------------
// Tokenizer

var (
	deptCode   = `[A-Z]{2,5}`
	courseNum  = `(?:\d{3}|X\d{2})[A-Z]?` // "225", "201W", "X99"
	courseRe   = regexp.MustCompile(`^(` + deptCode + `)\s?(` + courseNum + `)\b`)
	courseCIRe = regexp.MustCompile(`^([A-Za-z]{2,5})\s?(` + courseNum + `)\b`) // "Math 232"
	aliasRe    = regexp.MustCompile(`^(` + deptCode + `)\s*\(\s*(?i:or|formerly|and)\s+(` + deptCode + `)\s*\)\s*(` + courseNum + `)\b`)
	slashDept  = regexp.MustCompile(`^(` + deptCode + `)\s*/\s*(` + deptCode + `)\s?(` + courseNum + `)\b`) // "MBB/HSCI 426"
	deptLike   = regexp.MustCompile(`^[A-Za-z]{2,5}\s?\d`)
	numberRe   = regexp.MustCompile(`^(\d{3}[A-Z]?)\b`)
	levelRe    = regexp.MustCompile(`^\s*(?:-|\s)?\s*(?i:level|division)\b`)
	numLevelRe = regexp.MustCompile(`^\d{3}\s*(?i:or|and|/|-|to)\s*\d{3}\s*-?\s*(?i:level|division)\b`)
	// "CMPT 300 or 400 division course": the course match is followed by a range tail
	levelTailRe = regexp.MustCompile(`^\s*(?i:or|and|/|-|to)\s*\d{3}\s*-?\s*(?i:level|division)\b`)
	unitsRe     = regexp.MustCompile(`^(?i)(\d{1,3})\s+(?:(?:upper|lower)[- ]division\s+|undergraduate\s+|graduate\s+|credit\s+)?(?:units?|credits?|credit\s+hours?)\b`)
	oneOfRe     = regexp.MustCompile(`^(?i)(?:(?:either|any)\s+)?(?:one|1)\s+(?:(?:course|courses)\s+)?(?:of|from)\b(?:\s+the\s+following(?:\s+courses)?)?\s*:?|^(?i)(?:either|any\s+of)\b(?:\s+the\s+following(?:\s+courses)?)?\s*:?`)
	permRe      = regexp.MustCompile(`^(?i)(?:(?:the\s+)?(?:written\s+)?(?:permission|approval|consent)\s+(?:of|from)\s+(?:the\s+)?(?:course\s+)?(?:instructors?|department|program|school|faculty|chair|advisor|director|supervisor|graduate\s+program\s+chair|undergraduate\s+(?:advisor|chair|curriculum\s+chair))|(?:instructor|departmental|department)\s+(?:permission|approval|consent))`)
	wordRe      = regexp.MustCompile(`^[A-Za-z][A-Za-z'-]*`)

	looseCourseRe = regexp.MustCompile(`\b[A-Z]{2,5}\s?\d{3}[A-Z]?\b`)
)

// Words that look like department codes but never are.
var notDept = map[string]bool{
	"OR": true, "AND": true, "ONE": true, "ALL": true, "BC": true, "IB": true,
	"AP": true, "GPA": true, "CGPA": true, "UBC": true, "SFU": true, "TOEFL": true,
	"IELTS": true,
}

func (p *Parser) isDept(d string) bool { return !notDept[d] }

// isKnownMiscased accepts "Math"/"Bus" style spellings of catalog departments.
func (p *Parser) isKnownMiscased(d string) bool {
	if !isUpper(d[0]) { // "Math 232" yes, "is 300" no
		return false
	}
	u := strings.ToUpper(d)
	return p.known[u] && !notDept[u] && !connectorWord[u]
}

// Title-case words that start clauses, never subjects.
var connectorWord = map[string]bool{"AND": true, "OR": true, "PLUS": true, "EITHER": true, "ANY": true, "BOTH": true, "ALL": true}

// Words that may sit between a course and a shorthand number ("PHIL 100, both
// of 210"); any other word means a following bare number isn't a course.
var listWord = map[string]bool{"formerly": true, "previously": true, "and": true, "or": true, "plus": true, "both": true, "all": true, "either": true, "of": true, "the": true, "following": true}

func (p *Parser) tokenize(s string, lastDept *string) []token {
	var toks []token
	prose := false // prose seen since the last course
	i := 0
	for i < len(s) {
		ch := s[i]
		switch ch {
		case ' ', ':', '"', '.':
			i++
			continue
		case '(', '[':
			toks = append(toks, token{kind: tLParen})
			i++
			continue
		case ')', ']':
			toks = append(toks, token{kind: tRParen})
			prose = false
			i++
			continue
		case ',':
			toks = append(toks, token{kind: tComma})
			prose = false
			i++
			continue
		case ';':
			toks = append(toks, token{kind: tSemi})
			i++
			continue
		case '/':
			toks = append(toks, token{kind: tSlash})
			i++
			continue
		case '&', '+':
			toks = append(toks, token{kind: tAnd})
			i++
			continue
		}
		rest := s[i:]

		// "CA (or FPA) 160" -> (CA 160 or FPA 160)
		if m := aliasRe.FindStringSubmatch(rest); m != nil && p.isDept(m[1]) {
			toks = append(toks,
				token{kind: tLParen},
				token{kind: tCourse, val: m[1] + " " + m[3]},
				token{kind: tOr},
				token{kind: tCourse, val: m[2] + " " + m[3]},
				token{kind: tRParen})
			*lastDept = m[1]
			prose = false
			i += len(m[0])
			continue
		}

		// "MBB/HSCI 426" -> (MBB 426 or HSCI 426)
		if m := slashDept.FindStringSubmatch(rest); m != nil && p.isDept(m[1]) && p.isDept(m[2]) {
			toks = append(toks,
				token{kind: tLParen},
				token{kind: tCourse, val: m[1] + " " + m[3]},
				token{kind: tOr},
				token{kind: tCourse, val: m[2] + " " + m[3]},
				token{kind: tRParen})
			*lastDept = m[2]
			prose = false
			i += len(m[0])
			continue
		}

		m := courseRe.FindStringSubmatch(rest)
		if m == nil || !p.isDept(m[1]) {
			if ci := courseCIRe.FindStringSubmatch(rest); ci != nil && p.isKnownMiscased(ci[1]) {
				m = []string{ci[0], strings.ToUpper(ci[1]), ci[2]}
			} else {
				m = nil
			}
		}
		if m != nil {
			*lastDept = m[1]
			i += len(m[0])
			if levelRe.MatchString(s[i:]) { // "CMPT 300-level course"
				continue
			}
			if lt := levelTailRe.FindString(s[i:]); lt != "" {
				i += len(lt)
				continue
			}
			toks = append(toks, token{kind: tCourse, val: m[1] + " " + strings.ToUpper(m[2])})
			prose = false
			continue
		}
		// An unrecognised subject followed by a number ("BC 12", "Math 11"):
		// later bare numbers must not inherit the previous department.
		if isUpper(rest[0]) && deptLike.MatchString(rest) {
			if w := wordRe.FindString(rest); !connectorWord[strings.ToUpper(w)] {
				*lastDept = ""
			}
		}

		if m := unitsRe.FindStringSubmatch(rest); m != nil {
			toks = append(toks, token{kind: tSpecial, val: "UNITS:" + m[1]})
			i += len(m[0])
			continue
		}

		if m := numberRe.FindStringSubmatch(rest); m != nil {
			i += len(m[0])
			if *lastDept == "" || prose || levelRe.MatchString(s[i:]) || numLevelRe.MatchString(rest) {
				// skip "300-level", "300 or 400 division", and the 2nd number of such ranges
				if nm := numLevelRe.FindString(rest); nm != "" {
					i = i - len(m[0]) + len(nm)
				}
				continue
			}
			toks = append(toks, token{kind: tCourse, val: *lastDept + " " + m[1]})
			prose = false
			continue
		}

		if m := oneOfRe.FindString(rest); m != "" {
			toks = append(toks, token{kind: tOneOf})
			i += len(m)
			continue
		}

		if m := permRe.FindString(rest); m != "" {
			toks = append(toks, token{kind: tSpecial, val: "PERMISSION:" + strings.ToLower(spaceRe.ReplaceAllString(m, " "))})
			i += len(m)
			continue
		}

		if w := wordRe.FindString(rest); w != "" {
			switch strings.ToLower(w) {
			case "and", "plus":
				toks = append(toks, token{kind: tAnd})
			case "or", "formerly", "previously":
				toks = append(toks, token{kind: tOr})
			}
			if !listWord[strings.ToLower(w)] {
				prose = true
			}
			i += len(w)
			continue
		}

		// digits that aren't course numbers (e.g. "2.0", "12"), other symbols
		i++
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
	}
	return toks
}

// ---------------------------------------------------------------------------
// Grammar

type sep int

const (
	sepComma    sep = iota // ,
	sepAnd                 // and
	sepOr                  // or
	sepCommaAnd            // , and
	sepCommaOr             // , or
)

// parseTokens parses one sentence: segments separated by ';' are OR-ed
// unless the segment starts with "and".
func parseTokens(toks []token) *model.PrereqNode {
	var result *model.PrereqNode
	depth := 0
	start := 0
	flush := func(seg []token) {
		op := "or"
		// leading connector after ';' decides how the segment joins
		for len(seg) > 0 && (seg[0].kind == tAnd || seg[0].kind == tOr || seg[0].kind == tComma) {
			if seg[0].kind == tAnd {
				op = "and"
			} else if seg[0].kind == tOr {
				op = "or"
			}
			seg = seg[1:]
		}
		n := parseList(seg, false)
		if n == nil {
			return
		}
		if result == nil {
			result = n
			return
		}
		result = &model.PrereqNode{Type: op, Children: []model.PrereqNode{*result, *n}}
	}
	for i, t := range toks {
		switch t.kind {
		case tLParen:
			depth++
		case tRParen:
			if depth > 0 {
				depth--
			}
		case tSemi:
			if depth == 0 {
				flush(toks[start:i])
				start = i + 1
			}
		}
	}
	flush(toks[start:])
	return result
}

// parseList parses a run of operands and connectors. In oneOf mode (after
// "one of"/"either"), commas mean OR.
func parseList(toks []token, oneOf bool) *model.PrereqNode {
	var ops []*model.PrereqNode
	var seps []sep
	var pending []tokKind // connectors seen since the last operand
	slash := false

	push := func(n *model.PrereqNode) {
		if n == nil {
			return
		}
		if len(ops) == 0 {
			ops = append(ops, n)
			pending = pending[:0]
			slash = false
			return
		}
		if slash { // "401/402": bind tightly to the previous operand
			ops[len(ops)-1] = orNode(ops[len(ops)-1], n)
			pending = pending[:0]
			slash = false
			return
		}
		seps = append(seps, mergeSeps(pending))
		ops = append(ops, n)
		pending = pending[:0]
	}

	for i := 0; i < len(toks); i++ {
		t := toks[i]
		switch t.kind {
		case tCourse, tSpecial:
			push(&model.PrereqNode{Type: "course", ID: t.val})
		case tAnd, tOr, tComma, tSemi:
			if t.kind == tSemi {
				t.kind = tOr
			}
			pending = append(pending, t.kind)
		case tSlash:
			if len(ops) > 0 && len(pending) == 0 {
				slash = true
			}
		case tRParen:
			// stray close paren: ignore
		case tLParen:
			end := matchParen(toks, i)
			inner := toks[i+1 : end]
			i = end
			// "(or ENSC 283)" / "(or 210)": alternative for the previous operand
			lead, body := leadingConnector(inner)
			if lead != -1 && len(ops) > 0 && len(pending) == 0 {
				if n := parseTokens(body); n != nil {
					if lead == tOr {
						ops[len(ops)-1] = orNode(ops[len(ops)-1], n)
					} else {
						ops[len(ops)-1] = andNode(ops[len(ops)-1], n)
					}
				}
				continue
			}
			push(parseTokens(body))
		case tOneOf:
			push(parseList(toks[i+1:], true))
			i = len(toks)
		}
	}
	if len(ops) == 0 {
		return nil
	}
	return combine(ops, seps, oneOf)
}

func matchParen(toks []token, open int) int {
	depth := 0
	for j := open; j < len(toks); j++ {
		switch toks[j].kind {
		case tLParen:
			depth++
		case tRParen:
			depth--
			if depth == 0 {
				return j
			}
		}
	}
	return len(toks) // unbalanced: treat the rest as the group
}

func leadingConnector(toks []token) (tokKind, []token) {
	lead := tokKind(-1)
	i := 0
	for i < len(toks) && (toks[i].kind == tOr || toks[i].kind == tAnd || toks[i].kind == tComma) {
		if toks[i].kind != tComma {
			lead = toks[i].kind
		}
		i++
	}
	return lead, toks[i:]
}

func mergeSeps(p []tokKind) sep {
	comma, op := false, tokKind(-1)
	for _, k := range p {
		switch k {
		case tComma:
			comma = true
		case tAnd, tOr:
			op = k
		}
	}
	switch {
	case comma && op == tAnd:
		return sepCommaAnd
	case comma && op == tOr:
		return sepCommaOr
	case op == tOr:
		return sepOr
	case op == tAnd:
		return sepAnd
	case comma:
		return sepComma
	}
	return sepAnd // adjacent operands with only prose between them
}

// combine applies SFU list conventions then operator precedence.
func combine(ops []*model.PrereqNode, seps []sep, oneOf bool) *model.PrereqNode {
	if oneOf {
		for i, s := range seps {
			if s == sepComma || s == sepCommaOr {
				seps[i] = sepOr
			}
		}
	} else {
		ops, seps = collapseCommaLists(ops, seps)
	}
	return splitBy(ops, seps, []sep{sepCommaAnd, sepCommaOr, sepAnd, sepOr})
}

// collapseCommaLists turns "A, B, C or D" into OR(A,B,C,D) and
// "A, B, and C" (or "A, B, C") into AND(A,B,C).
func collapseCommaLists(ops []*model.PrereqNode, seps []sep) ([]*model.PrereqNode, []sep) {
	var nOps []*model.PrereqNode
	var nSeps []sep
	i := 0 // index into seps; operand i precedes seps[i]
	nOps = append(nOps, ops[0])
	for i < len(seps) {
		if seps[i] != sepComma {
			nSeps = append(nSeps, seps[i])
			nOps = append(nOps, ops[i+1])
			i++
			continue
		}
		// comma run: operands [start .. j]
		start := len(nOps) - 1
		members := []*model.PrereqNode{nOps[start]}
		j := i
		for j < len(seps) && seps[j] == sepComma {
			members = append(members, ops[j+1])
			j++
		}
		typ := "and"
		if j < len(seps) {
			switch seps[j] {
			case sepOr, sepCommaOr:
				typ = "or"
				members = append(members, ops[j+1])
				j++
				for j < len(seps) && seps[j] == sepOr {
					members = append(members, ops[j+1])
					j++
				}
			case sepAnd, sepCommaAnd:
				last := ops[j+1]
				j++
				for j < len(seps) && seps[j] == sepOr { // "A, B and C or D" -> A, B, (C or D)
					last = orNode(last, ops[j+1])
					j++
				}
				members = append(members, last)
			}
		}
		nOps[start] = &model.PrereqNode{Type: typ, Children: deref(members)}
		i = j
	}
	return nOps, nSeps
}

func splitBy(ops []*model.PrereqNode, seps []sep, order []sep) *model.PrereqNode {
	if len(ops) == 1 {
		return ops[0]
	}
	if len(order) == 0 { // shouldn't happen; default AND
		return &model.PrereqNode{Type: "and", Children: deref(ops)}
	}
	s := order[0]
	typ := "and"
	if s == sepCommaOr || s == sepOr {
		typ = "or"
	}
	var groups []*model.PrereqNode
	gStart := 0
	for i, x := range seps {
		if x == s {
			groups = append(groups, splitBy(ops[gStart:i+1], seps[gStart:i], order[1:]))
			gStart = i + 1
		}
	}
	if len(groups) == 0 {
		return splitBy(ops, seps, order[1:])
	}
	groups = append(groups, splitBy(ops[gStart:], seps[gStart:], order[1:]))
	return &model.PrereqNode{Type: typ, Children: deref(groups)}
}

// ---------------------------------------------------------------------------
// Tree helpers

var validIDPattern = regexp.MustCompile(`^[A-Z]{2,5} (?:\d{3}|X\d{2})[A-Z]?$|^(UNITS|PERMISSION|COREQUISITE):`)

func deref(ns []*model.PrereqNode) []model.PrereqNode {
	out := make([]model.PrereqNode, 0, len(ns))
	for _, n := range ns {
		if n != nil {
			out = append(out, *n)
		}
	}
	return out
}

func orNode(a, b *model.PrereqNode) *model.PrereqNode {
	return &model.PrereqNode{Type: "or", Children: deref([]*model.PrereqNode{a, b})}
}

func andNode(a, b *model.PrereqNode) *model.PrereqNode {
	return &model.PrereqNode{Type: "and", Children: deref([]*model.PrereqNode{a, b})}
}

func joinAnd(ns ...*model.PrereqNode) *model.PrereqNode {
	var kept []*model.PrereqNode
	for _, n := range ns {
		if n != nil {
			kept = append(kept, n)
		}
	}
	switch len(kept) {
	case 0:
		return nil
	case 1:
		return kept[0]
	}
	return finalize(&model.PrereqNode{Type: "and", Children: deref(kept)})
}

func prefixLeaves(n *model.PrereqNode, prefix string) *model.PrereqNode {
	if n == nil {
		return nil
	}
	c := *n
	if c.Type == "course" {
		if !strings.Contains(c.ID, ":") {
			c.ID = prefix + c.ID
		}
		return &c
	}
	c.Children = make([]model.PrereqNode, len(n.Children))
	for i := range n.Children {
		c.Children[i] = *prefixLeaves(&n.Children[i], prefix)
	}
	return &c
}

func finalize(n *model.PrereqNode) *model.PrereqNode {
	return dedupe(prune(flatten(n)))
}

// prune removes invalid leaves and collapses single-child operators.
func prune(n *model.PrereqNode) *model.PrereqNode {
	if n == nil {
		return nil
	}
	if n.Type == "course" {
		if validIDPattern.MatchString(n.ID) {
			return n
		}
		return nil
	}
	var kept []model.PrereqNode
	for i := range n.Children {
		if pr := prune(&n.Children[i]); pr != nil {
			kept = append(kept, *pr)
		}
	}
	switch len(kept) {
	case 0:
		return nil
	case 1:
		return &kept[0]
	}
	n.Children = kept
	return n
}

// flatten collapses nested same-type nodes: AND(AND(A,B),C) -> AND(A,B,C).
func flatten(n *model.PrereqNode) *model.PrereqNode {
	if n == nil || n.Type == "course" {
		return n
	}
	var flat []model.PrereqNode
	for i := range n.Children {
		c := flatten(&n.Children[i])
		if c == nil {
			continue
		}
		if c.Type == n.Type {
			flat = append(flat, c.Children...)
		} else {
			flat = append(flat, *c)
		}
	}
	n.Children = flat
	return n
}

// dedupe removes repeated identical leaves under the same operator.
func dedupe(n *model.PrereqNode) *model.PrereqNode {
	if n == nil || n.Type == "course" {
		return n
	}
	seen := map[string]bool{}
	var out []model.PrereqNode
	for i := range n.Children {
		c := dedupe(&n.Children[i])
		if c.Type == "course" {
			if seen[c.ID] {
				continue
			}
			seen[c.ID] = true
		}
		out = append(out, *c)
	}
	if len(out) == 1 {
		return &out[0]
	}
	n.Children = out
	return n
}

// Leaves returns the leaf IDs of a tree in order of appearance (deduplicated).
func Leaves(n *model.PrereqNode) []string {
	var out []string
	seen := map[string]bool{}
	var walk func(*model.PrereqNode)
	walk = func(n *model.PrereqNode) {
		if n == nil {
			return
		}
		if n.Type == "course" {
			if !seen[n.ID] {
				seen[n.ID] = true
				out = append(out, n.ID)
			}
			return
		}
		for i := range n.Children {
			walk(&n.Children[i])
		}
	}
	walk(n)
	return out
}
