package prereq

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	"github.com/brianrahadi/sfucourses-api/internal/model"
)

// Override replaces the catalog text for one course before parsing. Use it for
// the handful of entries SFU writes in a way the parser can't follow; the
// parser still produces the tree, so overrides stay small and reviewable.
// A field left empty keeps the catalog text.
type Override struct {
	Prerequisites *string `json:"prerequisites,omitempty"`
	Corequisites  *string `json:"corequisites,omitempty"`
	Reason        string  `json:"reason,omitempty"`
}

// Overrides maps "DEPT NUMBER" to an Override.
type Overrides map[string]Override

// LoadOverrides reads an overrides file. A missing file is not an error.
func LoadOverrides(path string) (Overrides, error) {
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return Overrides{}, nil
	}
	if err != nil {
		return nil, err
	}
	var o Overrides
	if err := json.Unmarshal(b, &o); err != nil {
		return nil, err
	}
	return o, nil
}

// Result holds everything derived from the catalog in one pass.
type Result struct {
	Trees model.PrereqMap          // "DEPT NUM" -> prerequisite tree (coreqs as COREQUISITE: leaves)
	Graph []model.PrereqGraphEntry // flat edges for the course graph, sorted by dept/number
	Stats Stats
}

// Stats summarises parse coverage; the weekly job prints it.
type Stats struct {
	Courses          int      `json:"courses"`
	WithPrereqText   int      `json:"withPrereqText"`
	WithParsedPrereq int      `json:"withParsedPrereq"`
	PrereqEdges      int      `json:"prereqEdges"`
	CoreqEdges       int      `json:"coreqEdges"`
	Unparsed         []string `json:"unparsed"` // text mentions a course code but nothing was extracted
}

// ParseAll parses prerequisites for all outlines and returns a PrereqMap.
func ParseAll(outlines []model.CourseOutline) model.PrereqMap {
	return Build(outlines, nil).Trees
}

// Build parses every outline using the catalog's own department list.
func Build(outlines []model.CourseOutline, overrides Overrides) Result {
	deptSet := map[string]bool{}
	for _, o := range outlines {
		deptSet[strings.ToUpper(o.Dept)] = true
	}
	depts := make([]string, 0, len(deptSet))
	for d := range deptSet {
		depts = append(depts, d)
	}
	p := New(depts)
	// Single-course callers (tests) pass outlines with few departments; fall back
	// to permissive matching so cross-department prereqs still parse.
	if len(depts) < 20 {
		p = defaultParser
	}

	res := Result{Trees: make(model.PrereqMap, len(outlines))}
	for _, o := range outlines {
		code := strings.ToUpper(o.Dept) + " " + strings.ToUpper(o.Number)
		preText, coText := o.Prerequisites, o.Corequisites
		if ov, ok := overrides[code]; ok {
			if ov.Prerequisites != nil {
				preText = *ov.Prerequisites
			}
			if ov.Corequisites != nil {
				coText = *ov.Corequisites
			}
		}

		req := p.ParseRequirements(preText, coText)
		req.Prereq = removeSelf(req.Prereq, code)
		req.Coreq = removeSelf(req.Coreq, code)

		res.Stats.Courses++
		if strings.TrimSpace(preText) != "" {
			res.Stats.WithPrereqText++
		}
		if tree := combineWithCoreqs(req.Prereq, req.Coreq); tree != nil {
			res.Trees[code] = *tree
		}

		entry := model.PrereqGraphEntry{
			Dept:          strings.ToUpper(o.Dept),
			Number:        strings.ToUpper(o.Number),
			Prerequisites: courseLeaves(dropRetiredAlternatives(req.Prereq, deptSet)),
			Corequisites:  courseLeaves(dropRetiredAlternatives(req.Coreq, deptSet)),
		}
		if len(entry.Prerequisites) > 0 {
			res.Stats.WithParsedPrereq++
		} else if mentionsCourse(preText) {
			res.Stats.Unparsed = append(res.Stats.Unparsed, code)
		}
		res.Stats.PrereqEdges += len(entry.Prerequisites)
		res.Stats.CoreqEdges += len(entry.Corequisites)
		res.Graph = append(res.Graph, entry)
	}
	sort.Slice(res.Graph, func(i, j int) bool {
		if res.Graph[i].Dept != res.Graph[j].Dept {
			return res.Graph[i].Dept < res.Graph[j].Dept
		}
		return res.Graph[i].Number < res.Graph[j].Number
	})
	sort.Strings(res.Stats.Unparsed)
	return res
}

// courseLeaves returns only real course codes (no UNITS:/PERMISSION: specials).
func courseLeaves(n *model.PrereqNode) []string {
	out := []string{}
	for _, id := range Leaves(n) {
		if !strings.Contains(id, ":") {
			out = append(out, id)
		}
	}
	return out
}

func removeSelf(n *model.PrereqNode, code string) *model.PrereqNode {
	if n == nil {
		return nil
	}
	if n.Type == "course" {
		if n.ID == code {
			return nil
		}
		return n
	}
	var kept []model.PrereqNode
	for i := range n.Children {
		if c := removeSelf(&n.Children[i], code); c != nil {
			kept = append(kept, *c)
		}
	}
	switch len(kept) {
	case 0:
		return nil
	case 1:
		return &kept[0]
	}
	c := *n
	c.Children = kept
	return &c
}

func mentionsCourse(s string) bool {
	return looseCourseRe.MatchString(s)
}

// dropRetiredAlternatives removes OR-alternatives whose subject code no longer
// exists ("CA (or FPA) 160" keeps CA 160) when a sibling uses a live subject.
// Course-level absence is not used: the catalog only holds recently offered
// courses, so a missing ARCH 201 is not retired. The full API tree keeps them.
func dropRetiredAlternatives(n *model.PrereqNode, liveDepts map[string]bool) *model.PrereqNode {
	if n == nil || n.Type == "course" {
		return n
	}
	c := *n
	c.Children = make([]model.PrereqNode, 0, len(n.Children))
	for i := range n.Children {
		c.Children = append(c.Children, *dropRetiredAlternatives(&n.Children[i], liveDepts))
	}
	if c.Type != "or" {
		return &c
	}
	isRetired := func(ch model.PrereqNode) bool {
		if ch.Type != "course" || strings.Contains(ch.ID, ":") {
			return false
		}
		return !liveDepts[strings.SplitN(ch.ID, " ", 2)[0]]
	}
	anyLive := false
	for _, ch := range c.Children {
		if ch.Type == "course" && !strings.Contains(ch.ID, ":") && !isRetired(ch) {
			anyLive = true
			break
		}
	}
	if !anyLive {
		return &c
	}
	kept := c.Children[:0]
	for _, ch := range c.Children {
		if !isRetired(ch) {
			kept = append(kept, ch)
		}
	}
	c.Children = kept
	return &c
}
