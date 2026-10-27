// Package prereqsync refreshes prerequisite/corequisite text in outlines.json
// from the SFU Course Outlines API.
//
// fetch-outlines keeps the first outline it sees for each course while walking
// terms oldest-first, so catalog wording can lag years behind. This package
// walks the newest terms first, takes each course's current requirement text,
// and merges only the requirement fields (plus courses not yet in the file),
// leaving titles, descriptions and offerings to the existing scripts.
//
// It uses only the standard library so it can run anywhere `go` does.
package prereqsync

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/brianrahadi/sfucourses-api/internal/model"
)

// DefaultBaseURL is SFU's public course outline API.
const DefaultBaseURL = "https://www.sfu.ca/bin/wcm/course-outlines"

// Term is a (year, season) pair such as {"2026", "fall"}.
type Term struct{ Year, Season string }

func (t Term) String() string { return t.Year + "-" + t.Season }

// TermsAround returns the next, current and previous terms for date, newest
// first. Next term's outlines are published ahead of time with the latest
// catalog wording, so it is checked first.
func TermsAround(date time.Time) []Term {
	seasons := []string{"spring", "summer", "fall"}
	idx := (int(date.Month()) - 1) / 4 // Jan-Apr spring, May-Aug summer, Sep-Dec fall
	year := date.Year()
	at := func(offset int) Term {
		i := idx + offset
		y := year
		for i < 0 {
			i += 3
			y--
		}
		for i > 2 {
			i -= 3
			y++
		}
		return Term{fmt.Sprint(y), seasons[i]}
	}
	return []Term{at(1), at(0), at(-1)}
}

// Client fetches outlines from the SFU API.
type Client struct {
	BaseURL string
	HTTP    *http.Client
	Workers int           // concurrent course fetches (default 8)
	Retries int           // per request (default 3)
	Backoff time.Duration // initial retry delay (default 2s)
	Log     func(format string, args ...any)
}

// FetchStats summarises a fetch.
type FetchStats struct {
	Terms    []string `json:"terms"`
	Courses  int      `json:"courses"`  // distinct courses with an outline
	Requests int      `json:"requests"` // HTTP requests made
	Failures int      `json:"failures"` // requests that failed after retries
}

type listItem struct {
	Value     string `json:"value"`
	ClassType string `json:"classType"`
}

type outlineRes struct {
	Info model.CourseOutline `json:"info"`
}

func (c *Client) defaults() {
	if c.BaseURL == "" {
		c.BaseURL = DefaultBaseURL
	}
	if c.HTTP == nil {
		c.HTTP = &http.Client{Timeout: 30 * time.Second}
	}
	if c.Workers <= 0 {
		c.Workers = 8
	}
	if c.Retries <= 0 {
		c.Retries = 3
	}
	if c.Backoff <= 0 {
		c.Backoff = 2 * time.Second
	}
	if c.Log == nil {
		c.Log = func(string, ...any) {}
	}
}

// Fetch returns the newest outline for every course offered in terms, which
// must be ordered newest first. Keys are "DEPT NUMBER" (upper case).
func (c *Client) Fetch(ctx context.Context, terms []Term) (map[string]model.CourseOutline, FetchStats, error) {
	c.defaults()
	var (
		mu    sync.Mutex
		out   = map[string]model.CourseOutline{}
		stats FetchStats
	)
	count := func(err error) {
		mu.Lock()
		stats.Requests++
		if err != nil {
			stats.Failures++
		}
		mu.Unlock()
	}

	for _, term := range terms {
		stats.Terms = append(stats.Terms, term.String())
		var depts []listItem
		err := c.get(ctx, &depts, term.Year, term.Season)
		count(err)
		if err != nil {
			c.Log("term %s: %v (skipping)", term, err)
			continue
		}

		type job struct{ dept, number, key string }
		var jobs []job
		for _, d := range depts {
			var courses []listItem
			err := c.get(ctx, &courses, term.Year, term.Season, d.Value)
			count(err)
			if err != nil {
				c.Log("%s %s: %v", term, d.Value, err)
				continue
			}
			for _, cr := range courses {
				key := strings.ToUpper(d.Value) + " " + strings.ToUpper(cr.Value)
				mu.Lock()
				_, seen := out[key]
				mu.Unlock()
				if !seen { // a newer term already supplied this course
					jobs = append(jobs, job{d.Value, cr.Value, key})
				}
			}
		}
		c.Log("term %s: %d departments, %d courses to fetch", term, len(depts), len(jobs))

		ch := make(chan job)
		var wg sync.WaitGroup
		for w := 0; w < c.Workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := range ch {
					var sections []listItem
					err := c.get(ctx, &sections, term.Year, term.Season, j.dept, j.number)
					count(err)
					if err != nil || len(sections) == 0 {
						continue
					}
					var res outlineRes
					err = c.get(ctx, &res, term.Year, term.Season, j.dept, j.number, pickSection(sections))
					count(err)
					if err != nil {
						continue
					}
					info := res.Info
					key := j.key
					if info.Dept != "" && info.Number != "" {
						info.Dept = strings.ToUpper(info.Dept)
						info.Number = strings.ToUpper(info.Number)
						key = info.Dept + " " + info.Number
					}
					mu.Lock()
					if _, seen := out[key]; !seen {
						out[key] = info
					}
					mu.Unlock()
				}
			}()
		}
	feed:
		for _, j := range jobs {
			select {
			case ch <- j:
			case <-ctx.Done():
				break feed
			}
		}
		close(ch)
		wg.Wait()
		if err := ctx.Err(); err != nil {
			return nil, stats, err
		}
	}
	stats.Courses = len(out)
	return out, stats, nil
}

// pickSection prefers an enrollment section (lectures carry the course-level
// requirement text) over labs/tutorials.
func pickSection(sections []listItem) string {
	for _, s := range sections {
		if strings.EqualFold(s.ClassType, "e") {
			return s.Value
		}
	}
	return sections[0].Value
}

func (c *Client) get(ctx context.Context, target any, path ...string) error {
	url := c.BaseURL + "?" + strings.Join(path, "/")
	var lastErr error
	delay := c.Backoff
	for attempt := 0; attempt < c.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
			delay *= 2
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "sfucourses-api prereq-sync (+https://github.com/brianrahadi/sfucourses-api)")
		resp, err := c.HTTP.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}
		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("GET %s: 404", url) // not retryable
		}
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("GET %s: status %d", url, resp.StatusCode)
			continue
		}
		if err := json.Unmarshal(body, target); err != nil {
			return fmt.Errorf("GET %s: decode: %w", url, err)
		}
		return nil
	}
	return lastErr
}

// Change records one updated requirement field.
type Change struct {
	Code  string `json:"code"`
	Field string `json:"field"` // "prerequisites" | "corequisites"
	Old   string `json:"old"`
	New   string `json:"new"`
}

// MergeResult describes what Merge changed.
type MergeResult struct {
	Outlines []model.CourseOutline
	Changes  []Change
	Added    []string
	// KeptOnEmpty lists fields where the newest outline had no text but the
	// file did; the old text is kept unless MergeOptions.AllowClear is set.
	KeptOnEmpty []Change
}

// MergeOptions controls Merge.
type MergeOptions struct {
	AddNew     bool // append courses missing from the file
	AllowClear bool // let an empty upstream field erase existing text
}

// Merge applies fetched requirement text to existing outlines. Only the
// prerequisites/corequisites fields of existing courses are touched. Courses
// that aren't in existing are appended when addNew is set. The result is
// sorted the same way fetch-outlines writes the file.
func Merge(existing []model.CourseOutline, fetched map[string]model.CourseOutline, opt MergeOptions) MergeResult {
	res := MergeResult{Outlines: make([]model.CourseOutline, 0, len(existing))}
	seen := make(map[string]bool, len(existing))
	for _, o := range existing {
		key := strings.ToUpper(o.Dept) + " " + strings.ToUpper(o.Number)
		seen[key] = true
		if f, ok := fetched[key]; ok {
			update := func(field string, cur *string, upstream string) {
				n := clean(upstream)
				if n == clean(*cur) {
					return
				}
				ch := Change{key, field, *cur, n}
				if n == "" && !opt.AllowClear {
					res.KeptOnEmpty = append(res.KeptOnEmpty, ch)
					return
				}
				res.Changes = append(res.Changes, ch)
				*cur = n
			}
			update("prerequisites", &o.Prerequisites, f.Prerequisites)
			update("corequisites", &o.Corequisites, f.Corequisites)
		}
		res.Outlines = append(res.Outlines, o)
	}
	if opt.AddNew {
		for key, f := range fetched {
			if seen[key] || f.Dept == "" || f.Number == "" {
				continue
			}
			f.Prerequisites = clean(f.Prerequisites)
			f.Corequisites = clean(f.Corequisites)
			if f.Offerings == nil {
				f.Offerings = []model.CourseOffering{}
			}
			res.Outlines = append(res.Outlines, f)
			res.Added = append(res.Added, key)
		}
	}
	sort.SliceStable(res.Outlines, func(i, j int) bool {
		a, b := res.Outlines[i], res.Outlines[j]
		if a.Dept != b.Dept {
			return a.Dept < b.Dept
		}
		return a.Number < b.Number
	})
	sort.Slice(res.Changes, func(i, j int) bool { return res.Changes[i].Code < res.Changes[j].Code })
	sort.Strings(res.Added)
	return res
}

// clean trims whitespace; the API sometimes returns trailing spaces/newlines
// that would otherwise show up as weekly no-op diffs.
func clean(s string) string { return strings.TrimSpace(s) }
