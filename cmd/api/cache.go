package main

import (
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
)

// Responses are cached by clients and CDNs for a few minutes and revalidated
// cheaply with ETags. Data syncs hourly, so this never hides a sync for long.
const (
	cacheControlDefault = "public, max-age=300, s-maxage=3600, stale-while-revalidate=86400"
	cacheControlHealth  = "public, max-age=60"
)

// Bodies smaller than this aren't worth compressing.
const minGzipSize = 1024

// Compressed copies of large bodies are kept so the full catalog endpoints
// (outlines, instructors, review summaries) aren't re-gzipped on every request.
const (
	minCachedGzipSize = 32 * 1024
	maxCachedGzips    = 64
)

type gzipCache struct {
	mu    sync.Mutex
	items map[string][]byte
	order []string
}

var compressedBodies = &gzipCache{items: make(map[string][]byte)}

func (c *gzipCache) get(etag string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	b, ok := c.items[etag]
	return b, ok
}

func (c *gzipCache) put(etag string, b []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, ok := c.items[etag]; ok {
		return
	}
	if len(c.order) >= maxCachedGzips {
		oldest := c.order[0]
		c.order = c.order[1:]
		delete(c.items, oldest)
	}
	c.items[etag] = b
	c.order = append(c.order, etag)
}

func gzipBytes(b []byte) []byte {
	var buf bytes.Buffer
	buf.Grow(len(b) / 4)
	gz, _ := gzip.NewWriterLevel(&buf, gzip.BestSpeed)
	gz.Write(b)
	gz.Close()
	return buf.Bytes()
}

// bufferedResponse captures a handler's response so it can be compressed,
// tagged and possibly answered with 304 Not Modified.
type bufferedResponse struct {
	header http.Header
	status int
	body   bytes.Buffer
}

func (b *bufferedResponse) Header() http.Header { return b.header }

func (b *bufferedResponse) WriteHeader(status int) {
	if b.status == 0 {
		b.status = status
	}
}

func (b *bufferedResponse) Write(p []byte) (int, error) {
	if b.status == 0 {
		b.status = http.StatusOK
	}
	return b.body.Write(p)
}

func acceptsGzip(r *http.Request) bool {
	for _, part := range strings.Split(r.Header.Get("Accept-Encoding"), ",") {
		fields := strings.Split(part, ";")
		enc := strings.ToLower(strings.TrimSpace(fields[0]))
		if enc != "gzip" && enc != "*" {
			continue
		}
		for _, param := range fields[1:] {
			param = strings.ReplaceAll(strings.TrimSpace(param), " ", "")
			if q, ok := strings.CutPrefix(param, "q="); ok {
				return strings.Trim(q, "0.") != ""
			}
		}
		return true
	}
	return false
}

func etagMatches(ifNoneMatch, etag string) bool {
	if ifNoneMatch == "" {
		return false
	}
	for _, candidate := range strings.Split(ifNoneMatch, ",") {
		candidate = strings.TrimSpace(candidate)
		candidate = strings.TrimPrefix(candidate, "W/")
		if candidate == "*" || candidate == etag {
			return true
		}
	}
	return false
}

// serveCacheable runs next and adds Cache-Control, ETag (with 304 support) and
// gzip compression to successful GET/HEAD responses.
func serveCacheable(next http.Handler, w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		next.ServeHTTP(w, r)
		return
	}

	buf := &bufferedResponse{header: w.Header()}
	next.ServeHTTP(buf, r)
	if buf.status == 0 {
		buf.status = http.StatusOK
	}

	h := w.Header()
	body := buf.body.Bytes()

	if buf.status != http.StatusOK {
		w.WriteHeader(buf.status)
		w.Write(body)
		return
	}

	sum := sha256.Sum256(body)
	etag := `"` + base64.RawURLEncoding.EncodeToString(sum[:12]) + `"`
	h.Set("ETag", etag)
	h.Add("Vary", "Accept-Encoding")
	if h.Get("Cache-Control") == "" {
		if r.URL.Path == "/health" {
			h.Set("Cache-Control", cacheControlHealth)
		} else {
			h.Set("Cache-Control", cacheControlDefault)
		}
	}

	if etagMatches(r.Header.Get("If-None-Match"), etag) {
		h.Del("Content-Type")
		w.WriteHeader(http.StatusNotModified)
		return
	}

	if len(body) >= minGzipSize && h.Get("Content-Encoding") == "" && acceptsGzip(r) {
		compressed, ok := compressedBodies.get(etag)
		if !ok {
			compressed = gzipBytes(body)
			if len(body) >= minCachedGzipSize {
				compressedBodies.put(etag, compressed)
			}
		}
		h.Set("Content-Encoding", "gzip")
		body = compressed
	}

	w.WriteHeader(http.StatusOK)
	w.Write(body)
}
