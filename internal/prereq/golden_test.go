package prereq

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"strings"
	"testing"

	"github.com/brianrahadi/sfucourses-api/internal/model"
)

// The golden test runs the parser over a frozen copy of the catalog text
// (testdata/catalog.json) and compares the graph with testdata/graph.golden.json.
// Any parser change shows up as a line-level diff of the affected courses in
// review. After an intended change, regenerate with:
//
//	go test ./internal/prereq -run Golden -update
//
// The fixture is frozen on purpose: weekly data refreshes don't touch it, so
// this test only fails when the parser's behaviour changes.
var update = flag.Bool("update", false, "rewrite testdata/graph.golden.json")

const (
	catalogFixture = "testdata/catalog.json"
	graphGolden    = "testdata/graph.golden.json"
)

func TestGolden(t *testing.T) {
	raw, err := os.ReadFile(catalogFixture)
	if err != nil {
		t.Fatal(err)
	}
	var outlines []model.CourseOutline
	if err := json.Unmarshal(raw, &outlines); err != nil {
		t.Fatal(err)
	}
	overrides, err := LoadOverrides("../store/json/prereq_overrides.json")
	if err != nil {
		t.Fatal(err)
	}
	got := renderGraph(Build(outlines, overrides).Graph)

	if *update {
		if err := os.WriteFile(graphGolden, got, 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(graphGolden)
	if err != nil {
		t.Fatalf("%v (run with -update to create it)", err)
	}
	if bytes.Equal(got, want) {
		return
	}
	gotLines := strings.Split(string(got), "\n")
	wantLines := strings.Split(string(want), "\n")
	shown := 0
	for i := 0; i < len(gotLines) || i < len(wantLines); i++ {
		var g, w string
		if i < len(gotLines) {
			g = gotLines[i]
		}
		if i < len(wantLines) {
			w = wantLines[i]
		}
		if g != w {
			t.Errorf("line %d\n  want %s\n  got  %s", i+1, w, g)
			if shown++; shown == 20 {
				t.Log("... more differences; run with -update after checking the change is intended")
				break
			}
		}
	}
	if shown == 0 {
		t.Error("golden output differs in length")
	}
}

// renderGraph writes one course per line so diffs are reviewable.
func renderGraph(g []model.PrereqGraphEntry) []byte {
	var b bytes.Buffer
	b.WriteString("[\n")
	for i, e := range g {
		line, _ := json.Marshal(e)
		b.Write(line)
		if i < len(g)-1 {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	b.WriteString("]\n")
	return b.Bytes()
}
