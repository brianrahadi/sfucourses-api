package main

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func jsonHandler(body string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, body)
	})
}

func serve(h http.Handler, req *http.Request) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	serveCacheable(h, rec, req)
	return rec
}

func TestServeCacheableCompressesWhenAccepted(t *testing.T) {
	body := strings.Repeat("sfucourses ", 500)
	req := httptest.NewRequest(http.MethodGet, "/v1/rest/outlines", nil)
	req.Header.Set("Accept-Encoding", "br, gzip")
	rec := serve(jsonHandler(body), req)

	if got := rec.Header().Get("Content-Encoding"); got != "gzip" {
		t.Fatalf("Content-Encoding = %q, want gzip", got)
	}
	if got := rec.Header().Get("Cache-Control"); got != cacheControlDefault {
		t.Errorf("Cache-Control = %q, want default", got)
	}
	if rec.Header().Get("ETag") == "" {
		t.Error("missing ETag")
	}
	zr, err := gzip.NewReader(rec.Body)
	if err != nil {
		t.Fatal(err)
	}
	plain, _ := io.ReadAll(zr)
	if !strings.Contains(string(plain), "sfucourses sfucourses") {
		t.Errorf("decompressed body doesn't contain the payload")
	}
}

func TestServeCacheablePlainWithoutAcceptEncoding(t *testing.T) {
	body := strings.Repeat("x", 4096)
	req := httptest.NewRequest(http.MethodGet, "/v1/rest/instructors", nil)
	rec := serve(jsonHandler(body), req)

	if got := rec.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("Content-Encoding = %q, want none", got)
	}
	if !strings.Contains(rec.Body.String(), body) {
		t.Error("body was altered")
	}
}

func TestServeCacheableNotModified(t *testing.T) {
	h := jsonHandler("hello")
	first := serve(h, httptest.NewRequest(http.MethodGet, "/health", nil))
	etag := first.Header().Get("ETag")
	if got := first.Header().Get("Cache-Control"); got != cacheControlHealth {
		t.Errorf("health Cache-Control = %q", got)
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	req.Header.Set("If-None-Match", "W/"+etag)
	second := serve(h, req)
	if second.Code != http.StatusNotModified {
		t.Fatalf("status = %d, want 304", second.Code)
	}
	if second.Body.Len() != 0 {
		t.Error("304 should have no body")
	}
}

func TestServeCacheableKeepsHandlerCacheControlAndErrors(t *testing.T) {
	custom := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=3600")
		writeJSON(w, http.StatusOK, "graph")
	})
	rec := serve(custom, httptest.NewRequest(http.MethodGet, "/v1/rest/prerequisites/graph", nil))
	if got := rec.Header().Get("Cache-Control"); got != "public, max-age=3600" {
		t.Errorf("Cache-Control = %q, want handler's value", got)
	}

	notFound := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSONError(w, http.StatusNotFound, "resource not found")
	})
	rec = serve(notFound, httptest.NewRequest(http.MethodGet, "/v1/rest/outlines?dept=zzz", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
	if rec.Header().Get("ETag") != "" || rec.Header().Get("Cache-Control") != "" {
		t.Error("errors must not be marked cacheable")
	}
}

func TestAcceptsGzip(t *testing.T) {
	cases := map[string]bool{
		"":                     false,
		"gzip":                 true,
		"gzip, deflate, br":    true,
		"br;q=1.0, gzip;q=0.8": true,
		"gzip;q=0":             false,
		"gzip; q=0.0":          false,
		"*":                    true,
		"identity":             false,
	}
	for header, want := range cases {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Accept-Encoding", header)
		if got := acceptsGzip(req); got != want {
			t.Errorf("acceptsGzip(%q) = %v, want %v", header, got, want)
		}
	}
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestReviewIndexLookup(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "Computing Science", "Brian_Fraser.json"))
	writeFile(t, filepath.Join(root, "Applied Sciences", "Brian_Fraser.json"))
	writeFile(t, filepath.Join(root, "Mathematics", "Jamie_Mulholland.json"))

	idx, err := buildInstructorReviewIndex(root)
	if err != nil {
		t.Fatal(err)
	}
	got := idx.lookup([]string{normalizeName("Brian Fraser"), "Brian_Fraser", "BrianFraser"})
	want := []string{
		filepath.Join(root, "Applied Sciences", "Brian_Fraser.json"),
		filepath.Join(root, "Computing Science", "Brian_Fraser.json"),
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("lookup = %v, want %v", got, want)
	}
	if got := idx.lookup([]string{"nobody"}); len(got) != 0 {
		t.Errorf("lookup(nobody) = %v, want none", got)
	}

	courseRoot := t.TempDir()
	writeFile(t, filepath.Join(courseRoot, "CMPT225.json"))
	writeFile(t, filepath.Join(courseRoot, "MATH151.json"))
	cidx, err := buildCourseReviewIndex(courseRoot)
	if err != nil {
		t.Fatal(err)
	}
	want = []string{filepath.Join(courseRoot, "CMPT225.json")}
	if got := cidx.lookup([]string{"CMPT225", "cmpt225"}); !reflect.DeepEqual(got, want) {
		t.Errorf("course lookup = %v, want %v", got, want)
	}
}
