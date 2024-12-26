package store

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	. "github.com/brianrahadi/sfucourses-api/internal/model"
)

type OutlineStore struct {
}

//go:embed json/outlines.json
var data []byte

func (s *OutlineStore) GetAll(ctx context.Context) ([]CourseOutline, error) {
	var outlines []CourseOutline
	if err := json.Unmarshal(data, &outlines); err != nil {
		return outlines, fmt.Errorf("error parsing JSON: %v", err)
	}

	return outlines, nil
}
