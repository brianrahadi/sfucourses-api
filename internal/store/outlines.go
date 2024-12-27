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
	"github.com/samber/mo"
)

//go:embed json/outlines.json
var data []byte

type OutlineStore struct {
	cachedOutlines []CourseOutline
}

func NewOutlineStore() (*OutlineStore, error) {
	var outlines []CourseOutline
	if err := json.Unmarshal(data, &outlines); err != nil {
		return nil, fmt.Errorf("error parsing JSON: %v", err)
	}

	return &OutlineStore{cachedOutlines: outlines}, nil
}

func (s *OutlineStore) GetAll(ctx context.Context, limitOpt mo.Option[int], offsetOpt mo.Option[int]) ([]CourseOutline, int, error) {
	totalCount := len(s.cachedOutlines)
	if limitOpt.IsAbsent() && offsetOpt.IsAbsent() {
		return s.cachedOutlines, totalCount, nil
	}
	limit := limitOpt.OrElse(totalCount)
	offset := offsetOpt.OrElse(0)

	if offset >= totalCount {
		return []CourseOutline{}, 0, nil
	}

	end := offset + limit
	if end > totalCount {
		end = totalCount
	}

	return s.cachedOutlines[offset:end], totalCount, nil
}

func (s *OutlineStore) GetByDept(ctx context.Context, dept string) ([]CourseOutline, error) {
	outlines := s.cachedOutlines
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
	outlines := s.cachedOutlines
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
