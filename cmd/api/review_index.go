package main

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// Review files only change when the review workflow commits new data and the
// API redeploys, so listing thousands of files on every request is wasted
// work. The listings are indexed once and refreshed hourly.
const reviewIndexTTL = time.Hour

type indexedFile struct {
	path string
	// pos is the file's position in directory scan order, so lookups return
	// files in the same order a full scan would.
	pos int
}

type reviewFileIndex struct {
	byName map[string][]indexedFile
}

// lookup returns the files whose name (without .json) matches any of the
// variations case-insensitively, in scan order and without duplicates.
func (idx *reviewFileIndex) lookup(variations []string) []string {
	seen := make(map[string]bool)
	var matches []indexedFile
	for _, v := range variations {
		for _, f := range idx.byName[strings.ToLower(v)] {
			if !seen[f.path] {
				seen[f.path] = true
				matches = append(matches, f)
			}
		}
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].pos < matches[j].pos })

	paths := make([]string, len(matches))
	for i, m := range matches {
		paths[i] = m.path
	}
	return paths
}

type cachedReviewIndex struct {
	mu      sync.Mutex
	index   *reviewFileIndex
	builtAt time.Time
	build   func() (*reviewFileIndex, error)
}

func (c *cachedReviewIndex) get() (*reviewFileIndex, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.index != nil && time.Since(c.builtAt) < reviewIndexTTL {
		return c.index, nil
	}
	index, err := c.build()
	if err != nil {
		if c.index != nil {
			// Keep serving the last good index rather than failing requests.
			return c.index, nil
		}
		return nil, err
	}
	c.index = index
	c.builtAt = time.Now()
	return index, nil
}

var instructorReviewFiles = &cachedReviewIndex{
	build: func() (*reviewFileIndex, error) { return buildInstructorReviewIndex(instructorReviewDataDir) },
}

var courseReviewFiles = &cachedReviewIndex{
	build: func() (*reviewFileIndex, error) { return buildCourseReviewIndex(courseReviewDataDir) },
}

func addJSONFile(idx *reviewFileIndex, dir, name string, pos *int) {
	if !strings.HasSuffix(name, ".json") {
		return
	}
	key := strings.ToLower(strings.TrimSuffix(name, ".json"))
	idx.byName[key] = append(idx.byName[key], indexedFile{path: filepath.Join(dir, name), pos: *pos})
	*pos++
}

// buildInstructorReviewIndex indexes <root>/<department>/<instructor>.json.
func buildInstructorReviewIndex(root string) (*reviewFileIndex, error) {
	departments, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	idx := &reviewFileIndex{byName: make(map[string][]indexedFile)}
	pos := 0
	for _, dept := range departments {
		if !dept.IsDir() {
			continue
		}
		deptPath := filepath.Join(root, dept.Name())
		files, err := os.ReadDir(deptPath)
		if err != nil {
			continue
		}
		for _, file := range files {
			if !file.IsDir() {
				addJSONFile(idx, deptPath, file.Name(), &pos)
			}
		}
	}
	return idx, nil
}

// buildCourseReviewIndex indexes <root>/<COURSE>.json.
func buildCourseReviewIndex(root string) (*reviewFileIndex, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}
	idx := &reviewFileIndex{byName: make(map[string][]indexedFile)}
	pos := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			addJSONFile(idx, root, entry.Name(), &pos)
		}
	}
	return idx, nil
}
