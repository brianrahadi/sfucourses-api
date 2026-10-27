package prereqsync

import (
	"encoding/json"
	"os"
	"time"

	"github.com/brianrahadi/sfucourses-api/internal/prereq"
)

// MetaPath is served as lastPrereqUpdate on /health.
const MetaPath = "./internal/store/json/prereq_meta.json"

// Meta records when requirement data last changed and how well it parsed.
type Meta struct {
	LastUpdated      string   `json:"lastUpdated"`
	Terms            []string `json:"terms"`
	Courses          int      `json:"courses"`
	WithPrereqText   int      `json:"withPrereqText"`
	WithParsedPrereq int      `json:"withParsedPrereq"`
	PrereqEdges      int      `json:"prereqEdges"`
	CoreqEdges       int      `json:"coreqEdges"`
	Unparsed         int      `json:"unparsed"`
}

// NewMeta builds metadata from parse stats.
func NewMeta(now time.Time, terms []string, st prereq.Stats) Meta {
	return Meta{
		LastUpdated:      now.UTC().Format(time.RFC3339),
		Terms:            terms,
		Courses:          st.Courses,
		WithPrereqText:   st.WithPrereqText,
		WithParsedPrereq: st.WithParsedPrereq,
		PrereqEdges:      st.PrereqEdges,
		CoreqEdges:       st.CoreqEdges,
		Unparsed:         len(st.Unparsed),
	}
}

// WriteMeta writes m as indented JSON.
func WriteMeta(path string, m Meta) error {
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(b, '\n'), 0o644)
}
