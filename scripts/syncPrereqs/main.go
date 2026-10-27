// sync-prereqs refreshes prerequisite/corequisite text in outlines.json from
// the newest SFU terms, then reports how the prerequisite graph changed.
//
// The API parses requirements at load time (internal/prereq), so updating the
// text is all that's needed for /v1/rest/prerequisites and
// /v1/rest/prerequisites/graph to change after the next deploy.
//
//	go run ./scripts/syncPrereqs            # fetch next/current/previous term
//	go run ./scripts/syncPrereqs -dry-run   # report only
//	go run ./scripts/syncPrereqs -terms 2027-spring,2026-fall
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"strings"
	"time"

	"github.com/brianrahadi/sfucourses-api/internal/model"
	"github.com/brianrahadi/sfucourses-api/internal/prereq"
	"github.com/brianrahadi/sfucourses-api/internal/prereqsync"
)

func main() {
	var (
		outlinesPath  = flag.String("outlines", "./internal/store/json/outlines.json", "outlines file to update")
		overridesPath = flag.String("overrides", "./internal/store/json/prereq_overrides.json", "manual prerequisite overrides")
		metaPath      = flag.String("meta", prereqsync.MetaPath, "metadata file written when data changes")
		termsFlag     = flag.String("terms", "", "comma-separated terms newest first, e.g. 2027-spring,2026-fall (default: next, current, previous)")
		baseURL       = flag.String("base-url", prereqsync.DefaultBaseURL, "SFU course outlines API")
		workers       = flag.Int("workers", 8, "concurrent requests")
		addNew        = flag.Bool("add-new", true, "add courses missing from outlines.json")
		allowClear    = flag.Bool("allow-clear", false, "let an empty upstream field erase existing text")
		minCourses    = flag.Int("min-courses", 1000, "abort without writing if fewer courses are fetched (upstream outage guard)")
		maxFailRate   = flag.Float64("max-fail-rate", 0.2, "abort without writing if this fraction of requests fail")
		dryRun        = flag.Bool("dry-run", false, "report changes without writing files")
		summaryPath   = flag.String("summary", os.Getenv("GITHUB_STEP_SUMMARY"), "append a markdown report here")
	)
	flag.Parse()
	log.SetFlags(0)

	terms, err := parseTerms(*termsFlag)
	if err != nil {
		log.Fatal(err)
	}

	raw, err := os.ReadFile(*outlinesPath)
	if err != nil {
		log.Fatalf("read outlines: %v", err)
	}
	var existing []model.CourseOutline
	if err := json.Unmarshal(raw, &existing); err != nil {
		log.Fatalf("parse outlines: %v", err)
	}
	overrides, err := prereq.LoadOverrides(*overridesPath)
	if err != nil {
		log.Fatalf("read overrides: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	client := &prereqsync.Client{
		BaseURL: *baseURL,
		Workers: *workers,
		Log:     func(f string, a ...any) { log.Printf(f, a...) },
	}
	start := time.Now()
	fetched, stats, err := client.Fetch(ctx, terms)
	if err != nil {
		log.Fatalf("fetch: %v", err)
	}
	log.Printf("fetched %d courses from %v in %s (%d requests, %d failed)",
		stats.Courses, stats.Terms, time.Since(start).Round(time.Second), stats.Requests, stats.Failures)

	if stats.Courses < *minCourses {
		log.Fatalf("only %d courses fetched (< %d): SFU API looks unavailable; not writing", stats.Courses, *minCourses)
	}
	if stats.Requests > 0 && float64(stats.Failures)/float64(stats.Requests) > *maxFailRate {
		log.Fatalf("%d/%d requests failed: too unreliable to trust; not writing", stats.Failures, stats.Requests)
	}

	merged := prereqsync.Merge(existing, fetched, prereqsync.MergeOptions{AddNew: *addNew, AllowClear: *allowClear})
	out, err := json.Marshal(merged.Outlines)
	if err != nil {
		log.Fatalf("marshal: %v", err)
	}

	before := prereq.Build(existing, overrides)
	after := prereq.Build(merged.Outlines, overrides)
	changed := !bytes.Equal(bytes.TrimSpace(raw), out)

	report := buildReport(stats, merged, before, after, changed, *dryRun)
	fmt.Print(report)
	if *summaryPath != "" {
		if f, err := os.OpenFile(*summaryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644); err == nil {
			f.WriteString(report)
			f.Close()
		}
	}

	if !changed || *dryRun {
		return
	}
	if err := os.WriteFile(*outlinesPath, out, 0o644); err != nil {
		log.Fatalf("write outlines: %v", err)
	}
	if err := prereqsync.WriteMeta(*metaPath, prereqsync.NewMeta(time.Now(), stats.Terms, after.Stats)); err != nil {
		log.Fatalf("write meta: %v", err)
	}
}

func parseTerms(s string) ([]prereqsync.Term, error) {
	if s == "" {
		return prereqsync.TermsAround(time.Now()), nil
	}
	var terms []prereqsync.Term
	for _, t := range strings.Split(s, ",") {
		y, season, ok := strings.Cut(strings.TrimSpace(t), "-")
		if !ok || len(y) != 4 || !slices.Contains([]string{"spring", "summer", "fall"}, season) {
			return nil, fmt.Errorf("bad term %q (want e.g. 2027-spring)", t)
		}
		terms = append(terms, prereqsync.Term{Year: y, Season: season})
	}
	return terms, nil
}

// buildReport renders a markdown summary for the job log / GitHub step summary.
func buildReport(stats prereqsync.FetchStats, m prereqsync.MergeResult, before, after prereq.Result, changed, dry bool) string {
	var b strings.Builder
	title := "Prerequisite sync"
	if dry {
		title += " (dry run)"
	}
	fmt.Fprintf(&b, "## %s\n\n", title)
	fmt.Fprintf(&b, "Terms: %s · %d courses fetched · %d requests (%d failed)\n\n",
		strings.Join(stats.Terms, ", "), stats.Courses, stats.Requests, stats.Failures)
	if !changed {
		b.WriteString("No requirement text changed.\n\n")
	}

	b.WriteString("| | Before | After |\n|---|---:|---:|\n")
	fmt.Fprintf(&b, "| Courses | %d | %d |\n", before.Stats.Courses, after.Stats.Courses)
	fmt.Fprintf(&b, "| With prerequisite text | %d | %d |\n", before.Stats.WithPrereqText, after.Stats.WithPrereqText)
	fmt.Fprintf(&b, "| With parsed prerequisites | %d | %d |\n", before.Stats.WithParsedPrereq, after.Stats.WithParsedPrereq)
	fmt.Fprintf(&b, "| Prerequisite edges | %d | %d |\n", before.Stats.PrereqEdges, after.Stats.PrereqEdges)
	fmt.Fprintf(&b, "| Corequisite edges | %d | %d |\n", before.Stats.CoreqEdges, after.Stats.CoreqEdges)
	fmt.Fprintf(&b, "| Mentions a course but nothing parsed | %d | %d |\n\n", len(before.Stats.Unparsed), len(after.Stats.Unparsed))

	edges := func(r prereq.Result) map[string]string {
		out := map[string]string{}
		for _, g := range r.Graph {
			out[g.Dept+" "+g.Number] = strings.Join(g.Prerequisites, ", ") + " | " + strings.Join(g.Corequisites, ", ")
		}
		return out
	}
	be, ae := edges(before), edges(after)
	var graphChanged []string
	for code, a := range ae {
		if be[code] != a {
			graphChanged = append(graphChanged, code)
		}
	}
	slices.Sort(graphChanged)
	fmt.Fprintf(&b, "**%d text changes**, **%d new courses**, **%d courses with changed graph edges**.\n\n",
		len(m.Changes), len(m.Added), len(graphChanged))

	if len(m.Changes) > 0 {
		b.WriteString("<details><summary>Text changes</summary>\n\n| Course | Field | Old | New |\n|---|---|---|---|\n")
		for i, c := range m.Changes {
			if i == 100 {
				fmt.Fprintf(&b, "| … | | %d more | |\n", len(m.Changes)-100)
				break
			}
			fmt.Fprintf(&b, "| %s | %s | %s | %s |\n", c.Code, c.Field, cell(c.Old), cell(c.New))
		}
		b.WriteString("\n</details>\n\n")
	}
	if len(m.Added) > 0 {
		fmt.Fprintf(&b, "New courses: %s\n\n", strings.Join(m.Added, ", "))
	}
	if len(m.KeptOnEmpty) > 0 {
		b.WriteString("Upstream returned empty text for these; existing text was kept (rerun with `-allow-clear` if intentional): ")
		var codes []string
		for _, c := range m.KeptOnEmpty {
			codes = append(codes, c.Code+" ("+c.Field+")")
		}
		b.WriteString(strings.Join(codes, ", ") + "\n\n")
	}

	// Newly unparseable text is where an override may be needed.
	wasUnparsed := map[string]bool{}
	for _, c := range before.Stats.Unparsed {
		wasUnparsed[c] = true
	}
	var newly []string
	for _, c := range after.Stats.Unparsed {
		if !wasUnparsed[c] {
			newly = append(newly, c)
		}
	}
	if len(newly) > 0 {
		fmt.Fprintf(&b, "⚠️ Newly unparsed (consider `internal/store/json/prereq_overrides.json`): %s\n\n", strings.Join(newly, ", "))
	}
	return b.String()
}

func cell(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	s = strings.ReplaceAll(s, "|", "\\|")
	if len(s) > 160 {
		s = s[:157] + "…"
	}
	if s == "" {
		return "_(empty)_"
	}
	return s
}
