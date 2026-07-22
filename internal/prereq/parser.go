package prereq

import (
	"regexp"
	"strings"

	"github.com/brianrahadi/sfucourses-api/internal/model"
)

type tokenType int

const (
	tokCourse tokenType = iota
	tokLParen
	tokRParen
	tokAnd
	tokOr
	tokComma
	tokSemicolon
	tokSpecial
	tokProse
)

type token struct {
	typ tokenType
	val string
}

var (
	coursePattern = regexp.MustCompile(`^([A-Z]{2,4})\s+(\d{3}[A-Z]?)`)
	numberPattern = regexp.MustCompile(`^(\d{2,3}[A-Z]?)`)
	unitsPattern  = regexp.MustCompile(`(?i)^(\d+)\s+units?\b`)
	validIDPattern = regexp.MustCompile(`^[A-Z]{2,4}\s+\d{3}[A-Z]?$|^(UNITS|PERMISSION|COREQUISITE):`)
)

func stripGradeText(s string) string {
	s = regexp.MustCompile(`(?i),?\s*all\s+with\s+a?\s*minimum\s+grade\s+of\s+[A-Z][+-]?`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`(?i),?\s*both\s+with\s+a?\s*(?:minimum\s+)?grade\s+(?:of\s+)?(?:at\s+least|of)\s+[A-Z][+-]?`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`(?i),?\s*with\s+a?\s*(?:minimum\s+)?grade\s+(?:of\s+)?(?:at\s+least|of)\s+[A-Z][+-]?`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`(?i),?\s*with\s+grades?\s+(?:of\s+)?(?:at\s+least|of)\s+[A-Z][+-]?`).ReplaceAllString(s, "")
	s = regexp.MustCompile(`(?i)\s*;+\s*or\b`).ReplaceAllString(s, " or ")
	s = regexp.MustCompile(`(?i)\s*;+\s*`).ReplaceAllString(s, " or ")
	s = regexp.MustCompile(`,\s*\.`).ReplaceAllString(s, ".")
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func stripFillerPhrases(s string) string {
_phrases := []string{
	`(?i)\bincluding\s+one\s+of\b`,
	`(?i)\bany\s+one\s+of\s+the\s+following\s*:?\s*`,
	`(?i)\bany\s+of\s+the\s+following\s*:?\s*`,
	`(?i)\bone\s+of\s+the\s+following\s*:?\s*`,
	`(?i)\bboth\s+with\b`,
	`(?i)\ball\s+with\b`,
}
	for _, p := range _phrases {
		s = regexp.MustCompile(p).ReplaceAllString(s, "")
	}
	return s
}

func tokenize(input string) []token {
	input = stripGradeText(input)
	input = stripFillerPhrases(input)

	var tokens []token
	i := 0
	n := len(input)
	lastDept := ""

	for i < n {
		ch := input[i]

		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			i++
			continue
		}

		if ch == '(' {
			tokens = append(tokens, token{tokLParen, "("})
			i++
			continue
		}

		if ch == ')' {
			tokens = append(tokens, token{tokRParen, ")"})
			i++
			continue
		}

		if ch == ',' {
			tokens = append(tokens, token{tokComma, ","})
			i++
			continue
		}

		remaining := input[i:]

		if strings.HasPrefix(strings.ToLower(remaining), "corequisite") {
			idx := strings.Index(strings.ToLower(remaining), ":")
			if idx > 0 {
				rest := strings.TrimSpace(remaining[idx+1:])
				endIdx := strings.IndexAny(rest, ",;(()")
				if endIdx < 0 {
					endIdx = len(rest)
				}
				val := strings.TrimSpace(rest[:endIdx])
				tokens = append(tokens, token{tokSpecial, "COREQUISITE:" + strings.ToUpper(val)})
				i += idx + 1 + endIdx
				continue
			}
		}

		if strings.HasPrefix(strings.ToLower(remaining), "permission") {
			endIdx := strings.IndexAny(remaining, ",;()")
			if endIdx < 0 {
				endIdx = len(remaining)
			}
			val := strings.TrimSpace(remaining[:endIdx])
			tokens = append(tokens, token{tokSpecial, "PERMISSION:" + strings.ToLower(val)})
			i += endIdx
			continue
		}

		if m := unitsPattern.FindStringSubmatch(remaining); m != nil {
			tokens = append(tokens, token{tokSpecial, "UNITS:" + m[1]})
			i += len(m[0])
			continue
		}

		if len(remaining) >= 3 && strings.EqualFold(remaining[:3], "and") {
			if len(remaining) == 3 || remaining[3] == ' ' || remaining[3] == '(' || remaining[3] == ')' {
				tokens = append(tokens, token{tokAnd, "and"})
				i += 3
				continue
			}
		}

		if len(remaining) >= 2 && strings.EqualFold(remaining[:2], "or") {
			if len(remaining) == 2 || remaining[2] == ' ' || remaining[2] == '(' || remaining[2] == ')' {
				tokens = append(tokens, token{tokOr, "or"})
				i += 2
				continue
			}
		}

		if m := coursePattern.FindStringSubmatch(strings.ToUpper(remaining)); m != nil {
			code := m[1] + " " + m[2]
			lastDept = m[1]
			tokens = append(tokens, token{tokCourse, code})
			i += len(m[0])
			continue
		}

		if m := numberPattern.FindStringSubmatch(strings.ToUpper(remaining)); m != nil && lastDept != "" {
			code := lastDept + " " + m[1]
			tokens = append(tokens, token{tokCourse, code})
			i += len(m[0])
			continue
		}

		if ch >= 'A' && ch <= 'Z' || ch >= 'a' && ch <= 'z' {
			endIdx := i + 1
			for endIdx < n && input[endIdx] != ' ' && input[endIdx] != '(' && input[endIdx] != ')' && input[endIdx] != ',' && input[endIdx] != ';' {
				endIdx++
			}
			word := remaining[:endIdx-i]
			lower := strings.ToLower(word)
			if lower == "and" {
				tokens = append(tokens, token{tokAnd, "and"})
			} else if lower == "or" {
				tokens = append(tokens, token{tokOr, "or"})
			} else {
				tokens = append(tokens, token{tokProse, word})
			}
			i = endIdx
			continue
		}

		i++
	}

	return tokens
}

type parser struct {
	tokens   []token
	pos      int
	lastDept string
}

func (p *parser) peek() *token {
	if p.pos >= len(p.tokens) {
		return nil
	}
	return &p.tokens[p.pos]
}

func (p *parser) advance() *token {
	if p.pos >= len(p.tokens) {
		return nil
	}
	t := &p.tokens[p.pos]
	p.pos++
	return t
}

// expression = term (and_op term)*
// and_op = 'and' | ','
// Adjacent and_ops (e.g. ", and") are collapsed.
func (p *parser) expression() *model.PrereqNode {
	left := p.term()
	if left == nil {
		return nil
	}

	for {
		t := p.peek()
		if t == nil {
			break
		}
		if t.typ == tokAnd || t.typ == tokComma {
			p.advance()
			// Skip adjacent AND operators (e.g. ", and" -> just one AND)
			for {
				next := p.peek()
				if next != nil && (next.typ == tokAnd || next.typ == tokComma) {
					p.advance()
					continue
				}
				break
			}
			right := p.term()
			if right == nil {
				break
			}
			left = &model.PrereqNode{
				Type:     "and",
				Children: []model.PrereqNode{*left, *right},
			}
			continue
		}
		break
	}

	return left
}

// term = factor (('or' | ';') factor)*
func (p *parser) term() *model.PrereqNode {
	left := p.factor()
	if left == nil {
		return nil
	}

	for {
		t := p.peek()
		if t == nil {
			break
		}
		if t.typ == tokOr || t.typ == tokSemicolon {
			p.advance()
			right := p.factor()
			if right == nil {
				break
			}
			left = &model.PrereqNode{
				Type:     "or",
				Children: []model.PrereqNode{*left, *right},
			}
			continue
		}
		break
	}

	return left
}

// factor = '(' expression ')' | course | special | prose
func (p *parser) factor() *model.PrereqNode {
	t := p.peek()
	if t == nil {
		return nil
	}

	if t.typ == tokLParen {
		p.advance()
		node := p.expression()
		if rt := p.peek(); rt != nil && rt.typ == tokRParen {
			p.advance()
		}
		return node
	}

	tok := p.advance()
	if tok == nil {
		return nil
	}

	switch tok.typ {
	case tokCourse:
		p.lastDept = tok.val[:strings.Index(tok.val, " ")]
		return &model.PrereqNode{Type: "course", ID: tok.val}
	case tokSpecial:
		return &model.PrereqNode{Type: "course", ID: tok.val}
	default:
		return nil
	}
}

// Parse converts a raw prerequisite string into a PrereqNode AST.
// Returns nil if the input is empty or unparseable.
func Parse(raw string) *model.PrereqNode {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	tokens := tokenize(raw)
	if len(tokens) == 0 {
		return nil
	}

	p := &parser{tokens: tokens}
	node := p.expression()

	if node == nil {
		return nil
	}

	return prune(flatten(node))
}

// prune removes invalid course nodes (IDs that aren't valid course codes or
// special tokens) and collapses meaningless parent nodes.
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
	for _, child := range n.Children {
		if pruned := prune(&child); pruned != nil {
			kept = append(kept, *pruned)
		}
	}

	if len(kept) == 0 {
		return nil
	}
	if len(kept) == 1 {
		return &kept[0]
	}

	n.Children = kept
	return n
}

// flatten collapses nested same-type nodes into a single level.
// e.g. AND(AND(A, B), C) -> AND(A, B, C)
func flatten(n *model.PrereqNode) *model.PrereqNode {
	if n == nil || n.Type == "course" {
		return n
	}
	for i := range n.Children {
		flattened := flatten(&n.Children[i])
		n.Children[i] = *flattened
	}
	var flat []model.PrereqNode
	for _, child := range n.Children {
		if child.Type == n.Type {
			flat = append(flat, child.Children...)
		} else {
			flat = append(flat, child)
		}
	}
	n.Children = flat
	return n
}

// ParseAll parses prerequisites for all outlines and returns a PrereqMap.
func ParseAll(outlines []model.CourseOutline) model.PrereqMap {
	m := make(model.PrereqMap, len(outlines))
	for _, o := range outlines {
		code := o.Dept + " " + o.Number
		node := Parse(o.Prerequisites)
		if node != nil {
			m[code] = *node
		}
	}
	return m
}
