package store

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	. "github.com/brianrahadi/sfucourses-api/internal/model"
	"github.com/samber/lo"
)

type OutlineStore struct {
}

//go:embed json/outlines.json
var data []byte

func (s *OutlineStore) GetAll(ctx context.Context) ([]CourseOutline, error) {
	var outlines []CourseOutline
	if err := json.Unmarshal(data, &outlines); err != nil {
		return []CourseOutline{}, fmt.Errorf("error parsing JSON: %v", err)
	}

	return outlines, nil
}

func (s *OutlineStore) GetByDept(ctx context.Context, dept string) ([]CourseOutline, error) {
	outlines, err := s.GetAll(ctx)
	if err != nil {
		return []CourseOutline{}, err
	}
	dept = strings.ToUpper(dept)
	outlines = lo.Filter(outlines, func(outline CourseOutline, _ int) bool {
		return outline.Dept == dept
	})

	if len(outlines) == 0 {
		return outlines, ErrNotFound
	}

	return outlines, nil
}

func (s *OutlineStore) GetByDeptAndNumber(ctx context.Context, dept string, number string) (CourseOutline, error) {
	outlines, err := s.GetAll(ctx)
	if err != nil {
		return CourseOutline{}, err
	}
	dept = strings.ToUpper(dept)
	number = strings.ToUpper(number)

	index := sort.Search(len(outlines), func(i int) bool {
		if outlines[i].Dept > dept {
			return true
		}
		if outlines[i].Dept == dept {
			return outlines[i].Number >= number
		}
		return false
	})

	if index < len(outlines) && outlines[index].Dept == dept && outlines[index].Number == number {
		return outlines[index], nil
	}

	return CourseOutline{}, ErrNotFound
}
