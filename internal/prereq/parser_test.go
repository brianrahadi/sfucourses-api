package prereq

import (
	"testing"

	"github.com/brianrahadi/sfucourses-api/internal/model"
)

func node(typ string, children ...model.PrereqNode) model.PrereqNode {
	return model.PrereqNode{Type: typ, Children: children}
}

func course(id string) model.PrereqNode {
	return model.PrereqNode{Type: "course", ID: id}
}

func TestParseEmpty(t *testing.T) {
	if got := Parse(""); got != nil {
		t.Errorf("Parse(\"\") = %v, want nil", got)
	}
	if got := Parse("  "); got != nil {
		t.Errorf("Parse(\"  \") = %v, want nil", got)
	}
}

func TestParseSingleCourse(t *testing.T) {
	got := Parse("CMPT 225")
	want := course("CMPT 225")
	if got.ID != want.ID || got.Type != want.Type {
		t.Errorf("Parse(\"CMPT 225\") = %v, want %v", got, want)
	}
}

func TestParseSimpleAnd(t *testing.T) {
	got := Parse("CMPT 225 and MACM 101")
	if got.Type != "and" {
		t.Fatalf("expected type 'and', got %q", got.Type)
	}
	if len(got.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(got.Children))
	}
	if got.Children[0].ID != "CMPT 225" {
		t.Errorf("child[0].ID = %q, want \"CMPT 225\"", got.Children[0].ID)
	}
	if got.Children[1].ID != "MACM 101" {
		t.Errorf("child[1].ID = %q, want \"MACM 101\"", got.Children[1].ID)
	}
}

func TestParseSimpleOr(t *testing.T) {
	got := Parse("CMPT 125 or CMPT 135")
	if got.Type != "or" {
		t.Fatalf("expected type 'or', got %q", got.Type)
	}
	if len(got.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(got.Children))
	}
	if got.Children[0].ID != "CMPT 125" || got.Children[1].ID != "CMPT 135" {
		t.Errorf("unexpected children: %v", got.Children)
	}
}

func TestParseCMPT300(t *testing.T) {
	got := Parse("CMPT 225 and (CMPT 295 or ENSC 254), all with a minimum grade of C-.")
	if got.Type != "and" {
		t.Fatalf("expected root type 'and', got %q", got.Type)
	}
	if len(got.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(got.Children))
	}
	if got.Children[0].ID != "CMPT 225" {
		t.Errorf("child[0] = %v, want course CMPT 225", got.Children[0])
	}
	orNode := got.Children[1]
	if orNode.Type != "or" || len(orNode.Children) != 2 {
		t.Errorf("child[1] = %v, want or(CMPT 295, ENSC 254)", orNode)
	}
}

func TestParseCMPT307(t *testing.T) {
	raw := "CMPT 225, (MACM 201 or CMPT 210), (MATH 150 or MATH 151), and (MATH 232 or MATH 240), all with a minimum grade of C-."
	got := Parse(raw)
	if got.Type != "and" {
		t.Fatalf("expected root type 'and', got %q", got.Type)
	}
	if len(got.Children) != 4 {
		t.Fatalf("expected 4 children, got %d: %v", len(got.Children), got)
	}
}

func TestParseCMPT383(t *testing.T) {
	raw := "CMPT 225 and (MACM 101 or (ENSC 251 and ENSC 252)), all with a minimum grade of C-."
	got := Parse(raw)
	if got.Type != "and" {
		t.Fatalf("expected root type 'and', got %q", got.Type)
	}
	if len(got.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(got.Children))
	}
	orChild := got.Children[1]
	if orChild.Type != "or" {
		t.Fatalf("expected child[1] type 'or', got %q", orChild.Type)
	}
	inner := orChild.Children[1]
	if inner.Type != "and" {
		t.Fatalf("expected nested 'and', got %q", inner.Type)
	}
}

func TestParseMATH232(t *testing.T) {
	raw := "MATH 150 or 151 or MACM 101, with a minimum grade of C-; or MATH 154 or 157, both with a grade of at least B."
	got := Parse(raw)
	if got.Type != "or" {
		t.Fatalf("expected root type 'or', got %q", got.Type)
	}
}

func TestParseShorthandDept(t *testing.T) {
	got := Parse("BISC 101, BISC 102, both with a minimum grade of C-.")
	if got.Type != "and" {
		t.Fatalf("expected type 'and', got %q", got.Type)
	}
	if len(got.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(got.Children))
	}
	if got.Children[0].ID != "BISC 101" || got.Children[1].ID != "BISC 102" {
		t.Errorf("unexpected children: %v", got.Children)
	}
}

func TestParseUnits(t *testing.T) {
	got := Parse("45 units")
	if got == nil {
		t.Fatal("expected non-nil node")
	}
	if got.Type != "course" || got.ID != "UNITS:45" {
		t.Errorf("got %v, want course(UNITS:45)", got)
	}
}

func TestParsePermission(t *testing.T) {
	got := Parse("Permission of the instructor")
	if got == nil {
		t.Fatal("expected non-nil node")
	}
	if got.Type != "course" || got.ID != "PERMISSION:permission of the instructor" {
		t.Errorf("got %v", got)
	}
}

func TestParseCorequisite(t *testing.T) {
	got := Parse("ACMA 201, Corequisite: STAT 285")
	if got == nil {
		t.Fatal("expected non-nil node")
	}
	if got.Type != "and" {
		t.Fatalf("expected type 'and', got %q", got.Type)
	}
	if len(got.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(got.Children))
	}
	if got.Children[1].ID != "COREQUISITE:STAT 285" {
		t.Errorf("child[1] = %v, want COREQUISITE:STAT 285", got.Children[1])
	}
}

func TestParseComplexOrOfAnds(t *testing.T) {
	raw := "(CMPT 125 and MACM 101) or (CMPT 225 and CMPT 295)"
	got := Parse(raw)
	if got.Type != "or" {
		t.Fatalf("expected root type 'or', got %q: %v", got.Type, got)
	}
	if len(got.Children) != 2 {
		t.Fatalf("expected 2 children, got %d: %v", len(got.Children), got)
	}
	for i, child := range got.Children {
		if child.Type != "and" {
			t.Errorf("child[%d].Type = %q, want 'and'", i, child.Type)
		}
	}
}

func TestParsePrunesInvalidCourseCodes(t *testing.T) {
	got := Parse("BC Math 12 or equivalent is recommended")
	if got != nil {
		t.Errorf("expected nil for unparseable prereq, got %v", got)
	}
}

func TestParsePrunesProse(t *testing.T) {
	got := Parse("Students must apply and receive permission from the co-op coordinator")
	if got != nil {
		t.Errorf("expected nil for prose-only prereq, got %v", got)
	}
}

func TestParseUnbalancedParens(t *testing.T) {
	cases := []struct {
		input     string
		expectNil bool
	}{
		{"(CMPT 225", false},
		{"CMPT 225)", false},
		{"((CMPT 225)", false},
		{"(", true},
		{")", true},
	}
	for _, c := range cases {
		got := Parse(c.input)
		if c.expectNil && got != nil {
			t.Errorf("Parse(%q) = %v, want nil", c.input, got)
		}
		if !c.expectNil && got == nil {
			t.Errorf("Parse(%q) returned nil, expected a node", c.input)
		}
	}
}

func TestParseAll(t *testing.T) {
	outlines := []model.CourseOutline{
		{Dept: "CMPT", Number: "300", Prerequisites: "CMPT 225"},
		{Dept: "CMPT", Number: "100", Prerequisites: ""},
	}
	m := ParseAll(outlines)
	if _, ok := m["CMPT 300"]; !ok {
		t.Error("expected CMPT 300 in map")
	}
	if _, ok := m["CMPT 100"]; ok {
		t.Error("did not expect CMPT 100 in map (empty prereqs)")
	}
}

func TestParseCMPT225(t *testing.T) {
	raw := "(MACM 101 and (CMPT 125, CMPT 129 or CMPT 135)) or (ENSC 251 and ENSC 252), all with a minimum grade of C-."
	got := Parse(raw)
	if got.Type != "or" {
		t.Fatalf("expected root type 'or', got %q: %v", got.Type, got)
	}
	if len(got.Children) != 2 {
		t.Fatalf("expected 2 children, got %d: %v", len(got.Children), got)
	}
	left := got.Children[0]
	if left.Type != "and" {
		t.Errorf("left child type = %q, want 'and': %v", left.Type, left)
	}
	right := got.Children[1]
	if right.Type != "and" {
		t.Errorf("right child type = %q, want 'and': %v", right.Type, right)
	}
}

func TestParseSimpleCourseWithGrade(t *testing.T) {
	got := Parse("CMPT 135 with a minimum grade of C-.")
	if got.Type != "course" || got.ID != "CMPT 135" {
		t.Errorf("got %v, want course(CMPT 135)", got)
	}
}

func TestParseOrWithSemicolon(t *testing.T) {
	raw := "MATH 152 with a minimum grade of C; or MATH 155 or MATH 158, with a grade of at least B."
	got := Parse(raw)
	if got.Type != "or" {
		t.Fatalf("expected root type 'or', got %q: %v", got.Type, got)
	}
}

func TestParseOrOfCourses(t *testing.T) {
	got := Parse("ARCH 101 or ARCH 201")
	if got.Type != "or" {
		t.Fatalf("expected type 'or', got %q", got.Type)
	}
	if len(got.Children) != 2 {
		t.Fatalf("expected 2 children, got %d", len(got.Children))
	}
}

func TestParseUnitsWithCourses(t *testing.T) {
	got := Parse("60 units, including either HSCI 130 or BPK 140, with a minimum grade of C-")
	if got == nil {
		t.Fatal("expected non-nil node")
	}
}

func TestParseBareUnits(t *testing.T) {
	got := Parse("30 units")
	if got == nil {
		t.Fatal("expected non-nil node")
	}
	if got.Type != "course" || got.ID != "UNITS:30" {
		t.Errorf("got %v, want course(UNITS:30)", got)
	}
}

// --- Catalog-derived cases (real SFU text) --------------------------------

func leaves(n *model.PrereqNode) []string { return Leaves(n) }

func eqSet(t *testing.T, raw string, got []string, want ...string) {
	t.Helper()
	g := map[string]bool{}
	for _, x := range got {
		g[x] = true
	}
	w := map[string]bool{}
	for _, x := range want {
		w[x] = true
	}
	if len(g) != len(w) {
		t.Errorf("%q\n got  %v\n want %v", raw, got, want)
		return
	}
	for k := range w {
		if !g[k] {
			t.Errorf("%q\n got  %v\n want %v", raw, got, want)
			return
		}
	}
}

func TestShorthandNumbersFormOrList(t *testing.T) {
	raw := "MATH 150, 151, 154 or 157."
	got := Parse(raw)
	if got.Type != "or" || len(got.Children) != 4 {
		t.Fatalf("got %+v, want or of 4", got)
	}
	eqSet(t, raw, leaves(got), "MATH 150", "MATH 151", "MATH 154", "MATH 157")
}

func TestOxfordAndList(t *testing.T) {
	raw := "BISC 101, 312, CHEM 102, and 103."
	got := Parse(raw)
	if got.Type != "and" || len(got.Children) != 4 {
		t.Fatalf("got %+v, want and of 4", got)
	}
	eqSet(t, raw, leaves(got), "BISC 101", "BISC 312", "CHEM 102", "CHEM 103")
}

func TestProseDoesNotAbortParse(t *testing.T) {
	cases := map[string][]string{
		"Completion of 21 units and IAT 102 with a minimum grade of C-.":                                 {"UNITS:21", "IAT 102"},
		"Either MATH 232 or MATH 240, with a minimum grade of C-.":                                       {"MATH 232", "MATH 240"},
		"Completion of HSCI 351 Co-op Practicum II.":                                                     {"HSCI 351"},
		"One of the following courses: ARCH 100, ARCH 201, HUM 110, or by permission of the instructor.": {"ARCH 100", "ARCH 201", "HUM 110", "PERMISSION:permission of the instructor"},
	}
	for raw, want := range cases {
		eqSet(t, raw, leaves(Parse(raw)), want...)
	}
}

func TestDeptAliasAndNumberAlias(t *testing.T) {
	raw := "CA (or FPA) 160. A course materials fee is required."
	eqSet(t, raw, leaves(Parse(raw)), "CA 160", "FPA 160")

	raw = "ACMA 201 (or 210), with a minimum grade of C."
	got := Parse(raw)
	if got.Type != "or" {
		t.Fatalf("got %+v, want or", got)
	}
	eqSet(t, raw, leaves(got), "ACMA 201", "ACMA 210")

	raw = "MBB/HSCI 426 or permission of the instructor."
	eqSet(t, raw, leaves(Parse(raw)), "MBB 426", "HSCI 426", "PERMISSION:permission of the instructor")

	raw = "Completion of EDUC 401/402."
	eqSet(t, raw, leaves(Parse(raw)), "EDUC 401", "EDUC 402")
}

func TestOneOfWithParenGroups(t *testing.T) {
	raw := "MATH 152, and one of: (PHYS 102 and PHYS 133), (PHYS 121 and PHYS 133), (PHYS 126 and PHYS 133) or (PHYS 141). All with a grade of C- or better."
	got := Parse(raw)
	if got.Type != "and" || len(got.Children) != 2 {
		t.Fatalf("got %+v", got)
	}
	if got.Children[1].Type != "or" || len(got.Children[1].Children) != 4 {
		t.Fatalf("one-of group = %+v, want or of 4", got.Children[1])
	}
}

func TestCommaAndSeparatesOrGroups(t *testing.T) {
	raw := "GEOG 111 or EASC 101; BISC 204 or GEOG 215; and a minimum of 60 units."
	got := Parse(raw)
	if got.Type != "and" {
		t.Fatalf("got %+v, want and at root", got)
	}
	raw = "MATH 151 or 154, and MATH 152 or 155"
	got = Parse(raw)
	if got.Type != "and" || len(got.Children) != 2 || got.Children[0].Type != "or" {
		t.Fatalf("got %+v, want and(or, or)", got)
	}
}

func TestCorequisiteSectionSeparated(t *testing.T) {
	req := New(nil).ParseRequirements("ACMA 201 (or 210), with a minimum grade of C. Corequisite: STAT 285.", "")
	eqSet(t, "prereq", Leaves(req.Prereq), "ACMA 201", "ACMA 210")
	eqSet(t, "coreq", Leaves(req.Coreq), "STAT 285")

	req = New(nil).ParseRequirements("EASC 101. /Corequsite: EASC 309 with a C- or better.", "")
	eqSet(t, "typo coreq", Leaves(req.Coreq), "EASC 309")

	req = New(nil).ParseRequirements("ARCH 282 or 372. Normally taken concurrently with ARCH 433 and 435.", "")
	eqSet(t, "concurrent prereq", Leaves(req.Prereq), "ARCH 282", "ARCH 372")
	eqSet(t, "concurrent coreq", Leaves(req.Coreq), "ARCH 433", "ARCH 435")
}

func TestRecommendedAndExclusionsDropped(t *testing.T) {
	cases := map[string][]string{
		"BUS 360W with a minimum grade of C-; 60 units. Recommended: BUS 346.":                   {"BUS 360W", "UNITS:60"},
		"(MACM 201 or CMPT 210) with a minimum grade of C-. CMPT 307 and 308 are recommended.":   {"MACM 201", "CMPT 210"},
		"60 units, including GEOG 213; GEOG 313 and EASC 201 recommended":                        {"UNITS:60", "GEOG 213"},
		"MBB 308 or 309W; Students may not take MBB 498 concurrently with MBB 481 or BISC 490.":  {"MBB 308", "MBB 309W"},
		"MSE 410. Must not be taken concurrently with MSE 493 or MSE 494.":                       {"MSE 410"},
		"BPK 305. Students with credit for BPK 306 may not take this course for further credit.": {"BPK 305"},
		"CRIM 103 is recommended.": {},
	}
	for raw, want := range cases {
		eqSet(t, raw, leaves(Parse(raw)), want...)
	}
}

func TestUnknownSubjectDoesNotLeakDepartment(t *testing.T) {
	raw := "IAT 374 (or WKTM 400)."
	eqSet(t, raw, leaves(Parse(raw)), "IAT 374", "WKTM 400")

	raw = "BC Math 12 or MATH 100, 151"
	eqSet(t, raw, leaves(Parse(raw)), "MATH 100", "MATH 151")

	raw = "CMPT 120; Math 11 or 12."
	eqSet(t, raw, leaves(Parse(raw)), "CMPT 120")
}

func TestLevelPhrasesAreNotCourses(t *testing.T) {
	raw := "MATH 152 and any 300-level CMPT course or CMPT 300 or 400 division course."
	eqSet(t, raw, leaves(Parse(raw)), "MATH 152")
	raw = "12 units or one 100-division English course."
	eqSet(t, raw, leaves(Parse(raw)), "UNITS:12")
}

func TestMiscasedKnownDepartments(t *testing.T) {
	p := New([]string{"MATH", "BUS", "IS"})
	eqSet(t, "Math", Leaves(p.Parse("Math 232, MATH 251 and a minimum of 80 credits.")), "MATH 232", "MATH 251", "UNITS:80")
	eqSet(t, "Bus", Leaves(p.Parse("Bus 792 Financial Literacy for Entrepreneurs.")), "BUS 792")
	eqSet(t, "is", Leaves(p.Parse("CMPT 120; this is 300 hours of work")), "CMPT 120")
}

func TestBuildGraphAndOverrides(t *testing.T) {
	outlines := []model.CourseOutline{
		{Dept: "CMPT", Number: "225", Prerequisites: "CMPT 125 or CMPT 135. Corequisite: MACM 101."},
		{Dept: "CMPT", Number: "125"},
		{Dept: "CMPT", Number: "135"},
		{Dept: "CMPT", Number: "300", Prerequisites: "CMPT 300 or CMPT 225"}, // self-reference dropped
		{Dept: "CA", Number: "160", Prerequisites: "CA (or FPA) 120."},
		{Dept: "CA", Number: "120"},
		{Dept: "CRIM", Number: "480", Prerequisites: "CRlM 380."},
	}
	fixed := "CRIM 380"
	r := Build(outlines, Overrides{"CRIM 480": {Prerequisites: &fixed}})
	byCode := map[string]model.PrereqGraphEntry{}
	for _, g := range r.Graph {
		byCode[g.Dept+" "+g.Number] = g
	}
	eqSet(t, "CMPT 225", byCode["CMPT 225"].Prerequisites, "CMPT 125", "CMPT 135")
	eqSet(t, "CMPT 225 coreq", byCode["CMPT 225"].Corequisites, "MACM 101")
	eqSet(t, "CMPT 300", byCode["CMPT 300"].Prerequisites, "CMPT 225")
	// FPA 120 isn't in the catalog and CA 120 is: the defunct alias is dropped from the graph...
	eqSet(t, "CA 160", byCode["CA 160"].Prerequisites, "CA 120")
	// ...but kept in the full tree.
	tree := r.Trees["CA 160"]
	eqSet(t, "CA 160 tree", Leaves(&tree), "CA 120", "FPA 120")
	eqSet(t, "override", byCode["CRIM 480"].Prerequisites, "CRIM 380")
	if r.Stats.WithParsedPrereq != 4 {
		t.Errorf("stats = %+v", r.Stats)
	}
}

func TestFoundationsAndSectionLabels(t *testing.T) {
	raw := "Pre-Calculus 11 (or equivalent) with a grade of at least B, or SFU FAN X99 course with a grade of at least C. This course may not be counted toward the Mathematics minor."
	eqSet(t, raw, leaves(Parse(raw)), "FAN X99")

	req := New(nil).ParseRequirements("Prerequisite/Corequisite: EASC 201, 204, 205, and 206. All with a grade of C- or better.", "")
	eqSet(t, "pre/co", Leaves(req.Prereq), "EASC 201", "EASC 204", "EASC 205", "EASC 206")

	req = New(nil).ParseRequirements("Permission of instructor. Recommended Prerequisite: ENSC 251, 316.", "")
	eqSet(t, "recommended prereq", Leaves(req.Prereq), "PERMISSION:permission of instructor")

	req = New(nil).ParseRequirements("BPK 207. Corequisite: BPK 305. Majors from outside BPK require BPK 205 (or BISC 305).", "")
	eqSet(t, "coreq sentence", Leaves(req.Coreq), "BPK 305")
	eqSet(t, "after coreq", Leaves(req.Prereq), "BPK 207", "BPK 205", "BISC 305")

	raw = "90 units including STAT 350 and one of STAT 341, STAT 260, or CMPT 225 (STAT 240 is a suitable substitute for STAT 341)."
	got := Parse(raw)
	if got.Type != "and" || len(got.Children) != 3 || got.Children[2].Type != "or" {
		t.Fatalf("got %+v", got)
	}

	raw = "GEOG 251 or one of STAT 201, 203 (formerly 103), 205, or 270."
	eqSet(t, raw, leaves(Parse(raw)), "GEOG 251", "STAT 201", "STAT 203", "STAT 103", "STAT 205", "STAT 270")
}
