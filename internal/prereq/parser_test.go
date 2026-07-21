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
