package prereqsync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/brianrahadi/sfucourses-api/internal/model"
)

// fakeSFU mimics the SFU outline API: ?year/term, ?year/term/dept,
// ?year/term/dept/number, ?year/term/dept/number/section.
func fakeSFU(t *testing.T) (*httptest.Server, *int32) {
	t.Helper()
	var flaky int32
	routes := map[string]any{
		"2027/spring":      []map[string]string{{"text": "CMPT", "value": "cmpt"}},
		"2027/spring/cmpt": []map[string]string{{"text": "225", "value": "225"}, {"text": "300", "value": "300"}},
		// lab listed first: the lecture (classType e) should be picked
		"2027/spring/cmpt/225":      []map[string]string{{"value": "d101", "classType": "n"}, {"value": "d100", "classType": "e"}},
		"2027/spring/cmpt/225/d100": map[string]any{"info": map[string]string{"dept": "CMPT", "number": "225", "prerequisites": "  (MACM 101 and CMPT 125) or ENSC 251.  ", "corequisites": ""}},
		"2027/spring/cmpt/225/d101": map[string]any{"info": map[string]string{"dept": "CMPT", "number": "225", "prerequisites": "WRONG: lab outline"}},
		"2027/spring/cmpt/300":      []map[string]string{{"value": "d100", "classType": "e"}},
		"2027/spring/cmpt/300/d100": map[string]any{"info": map[string]string{"dept": "CMPT", "number": "300", "prerequisites": "", "corequisites": ""}},

		"2026/fall":      []map[string]string{{"text": "CMPT", "value": "cmpt"}},
		"2026/fall/cmpt": []map[string]string{{"value": "225"}, {"value": "999"}},
		// older term text for 225 must lose to 2027 spring
		"2026/fall/cmpt/225":      []map[string]string{{"value": "d100", "classType": "e"}},
		"2026/fall/cmpt/225/d100": map[string]any{"info": map[string]string{"dept": "CMPT", "number": "225", "prerequisites": "OLD TEXT"}},
		"2026/fall/cmpt/999":      []map[string]string{{"value": "d100", "classType": "e"}},
		"2026/fall/cmpt/999/d100": map[string]any{"info": map[string]string{"dept": "CMPT", "number": "999", "title": "New Course", "prerequisites": "CMPT 225."}},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.URL.RawQuery
		if key == "2026/fall/cmpt/999" && atomic.AddInt32(&flaky, 1) == 1 {
			http.Error(w, "busy", http.StatusServiceUnavailable) // first try fails, retry succeeds
			return
		}
		v, ok := routes[key]
		if !ok {
			http.NotFound(w, r)
			return
		}
		json.NewEncoder(w).Encode(v)
	}))
	t.Cleanup(srv.Close)
	return srv, &flaky
}

func TestFetchNewestTermWinsAndRetries(t *testing.T) {
	srv, flaky := fakeSFU(t)
	c := &Client{BaseURL: srv.URL, Workers: 3, Backoff: time.Millisecond}
	terms := []Term{{"2027", "spring"}, {"2026", "fall"}, {"2026", "summer"} /* 404: skipped */}
	got, stats, err := c.Fetch(context.Background(), terms)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 || stats.Courses != 3 {
		t.Fatalf("got %d courses (%+v)", len(got), stats)
	}
	if p := got["CMPT 225"].Prerequisites; p != "  (MACM 101 and CMPT 125) or ENSC 251.  " {
		t.Errorf("CMPT 225 prereq = %q, want newest lecture outline", p)
	}
	if got["CMPT 999"].Title != "New Course" {
		t.Errorf("CMPT 999 = %+v", got["CMPT 999"])
	}
	if atomic.LoadInt32(flaky) < 2 {
		t.Errorf("expected a retry on 503")
	}
	if stats.Failures != 1 { // the 2026-summer term 404
		t.Errorf("failures = %d, want 1", stats.Failures)
	}
}

func TestMerge(t *testing.T) {
	existing := []model.CourseOutline{
		{Dept: "CMPT", Number: "300", Title: "OS", Prerequisites: "CMPT 225.", Offerings: []model.CourseOffering{{Term: "Fall 2026"}}},
		{Dept: "CMPT", Number: "225", Title: "Data Structures", Description: "keep me", Prerequisites: "OLD", Offerings: []model.CourseOffering{{Term: "Fall 2026"}}},
		{Dept: "ARCH", Number: "100", Prerequisites: "untouched"},
	}
	fetched := map[string]model.CourseOutline{
		"CMPT 225": {Dept: "CMPT", Number: "225", Title: "IGNORED", Prerequisites: " NEW \n", Corequisites: "MACM 101"},
		"CMPT 300": {Dept: "CMPT", Number: "300", Prerequisites: ""}, // empty upstream
		"CMPT 999": {Dept: "CMPT", Number: "999", Prerequisites: "CMPT 225."},
	}

	r := Merge(existing, fetched, MergeOptions{AddNew: true})
	by := map[string]model.CourseOutline{}
	for _, o := range r.Outlines {
		by[o.Dept+" "+o.Number] = o
	}
	if c := by["CMPT 225"]; c.Prerequisites != "NEW" || c.Corequisites != "MACM 101" || c.Title != "Data Structures" || c.Description != "keep me" || len(c.Offerings) != 1 {
		t.Errorf("CMPT 225 merged wrong: %+v", c)
	}
	if by["CMPT 300"].Prerequisites != "CMPT 225." || len(r.KeptOnEmpty) != 1 {
		t.Errorf("empty upstream should keep old text: %+v kept=%v", by["CMPT 300"], r.KeptOnEmpty)
	}
	if by["ARCH 100"].Prerequisites != "untouched" {
		t.Error("course not fetched was modified")
	}
	if n, ok := by["CMPT 999"]; !ok || n.Offerings == nil {
		t.Errorf("new course not added with empty offerings: %+v", n)
	}
	if len(r.Changes) != 2 || len(r.Added) != 1 {
		t.Errorf("changes=%v added=%v", r.Changes, r.Added)
	}
	// sorted by dept then number, like fetch-outlines writes it
	if r.Outlines[0].Dept != "ARCH" || r.Outlines[1].Number != "225" || r.Outlines[3].Number != "999" {
		t.Errorf("order: %v %v %v %v", r.Outlines[0].Dept, r.Outlines[1].Number, r.Outlines[2].Number, r.Outlines[3].Number)
	}

	r = Merge(existing, fetched, MergeOptions{AllowClear: true})
	for _, o := range r.Outlines {
		if o.Number == "300" && o.Prerequisites != "" {
			t.Error("AllowClear should erase")
		}
		if o.Number == "999" {
			t.Error("AddNew=false should not add")
		}
	}
}

func TestMergeNoChangeKeepsBytes(t *testing.T) {
	existing := []model.CourseOutline{{Dept: "CMPT", Number: "225", Prerequisites: "CMPT 125.\r\n"}}
	r := Merge(existing, map[string]model.CourseOutline{"CMPT 225": {Prerequisites: "CMPT 125."}}, MergeOptions{})
	if len(r.Changes) != 0 || r.Outlines[0].Prerequisites != "CMPT 125.\r\n" {
		t.Errorf("whitespace-only difference should be a no-op: %+v", r)
	}
}

func TestTermsAround(t *testing.T) {
	cases := map[string][]Term{
		"2026-10-27": {{"2027", "spring"}, {"2026", "fall"}, {"2026", "summer"}},
		"2026-01-05": {{"2026", "summer"}, {"2026", "spring"}, {"2025", "fall"}},
		"2026-06-01": {{"2026", "fall"}, {"2026", "summer"}, {"2026", "spring"}},
	}
	for d, want := range cases {
		date, _ := time.Parse("2006-01-02", d)
		got := TermsAround(date)
		for i := range want {
			if got[i] != want[i] {
				t.Errorf("%s: got %v want %v", d, got, want)
				break
			}
		}
	}
}

func TestMergeCatalogFields(t *testing.T) {
	existing := []model.CourseOutline{{Dept: "CMPT", Number: "225", Title: "Old Title", Description: "Old desc", Notes: "", Units: "3", Prerequisites: "CMPT 125."}}
	fetched := map[string]model.CourseOutline{"CMPT 225": {
		Title: "Data Structures and Programming", Description: "New desc ", Notes: "Students with credit for CMPT 226 may not take this course for further credit.",
		Units: "3", Prerequisites: "CMPT 125.", Designation: "",
	}}
	r := Merge(existing, fetched, MergeOptions{Fields: CatalogFields})
	o := r.Outlines[0]
	if o.Title != "Data Structures and Programming" || o.Description != "New desc" || o.Notes == "" {
		t.Errorf("catalog fields not refreshed: %+v", o)
	}
	if len(r.Changes) != 3 {
		t.Errorf("changes = %+v, want title/description/notes", r.Changes)
	}
	if !ValidField("notes") || ValidField("offerings") {
		t.Error("ValidField")
	}
}
