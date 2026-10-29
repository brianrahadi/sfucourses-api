package names

import (
	"reflect"
	"testing"
)

func TestClean(t *testing.T) {
	cases := map[string][]string{
		"Brian Fraser":                    {"Brian Fraser"},
		"  Brian   Fraser ":               {"Brian Fraser"},
		". .":                             nil,
		"Sessional":                       nil,
		"Sessional Instructor":            nil,
		"Sessional (Lab Instructor)":      nil,
		"Instructor TBA":                  nil,
		"TBA .":                           nil,
		". TBA":                           nil,
		"- TBA":                           nil,
		"TBA TBA":                         nil,
		"TBA Sessional":                   nil,
		"Faculty ,":                       nil,
		"":                                nil,
		"Erin Barley (Lab Instructor)":    {"Erin Barley"},
		"Erin Barley (Lecture and Lab)":   {"Erin Barley"},
		"Rolf Mathewes (Lectures)":        {"Rolf Mathewes"},
		"Peter Hollmann (Lab Instructor)": {"Peter Hollmann"},
		"Faraz Hach (UBC)":                {"Faraz Hach"},
		"Megan Barker, Lab Instructor":    {"Megan Barker"},
		"Cary Campbell & Michael Ling":    {"Cary Campbell", "Michael Ling"},
		"Baran (Abu) Fakhri":              {"Baran (Abu) Fakhri"},
		"Baran (Abu) Fakhri Fakhri":       {"Baran (Abu) Fakhri"},
		"Jean-François O'Neil":            {"Jean-François O'Neil"},
		"Tom Staffordson":                 {"Tom Staffordson"},
		"Yan Yan":                         {"Yan Yan"},
		"Man Man Wong":                    {"Man Man Wong"},
		"Travers Travers":                 {"Travers Travers"},
		"TBD .":                           nil,
	}
	for raw, want := range cases {
		if got := Clean(raw); !reflect.DeepEqual(got, want) {
			t.Errorf("Clean(%q) = %q, want %q", raw, got, want)
		}
	}
}

func TestCleanAll(t *testing.T) {
	got := CleanAll([]string{
		"Erin Barley (Lab Instructor)", "Erin Barley (Lecture and Lab)",
		"Sessional", "Cary Campbell & Michael Ling", "Michael Ling",
	})
	want := []string{"Erin Barley", "Cary Campbell", "Michael Ling"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("CleanAll = %q, want %q", got, want)
	}
}
