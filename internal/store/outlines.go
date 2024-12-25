package store

import (
	"context"
	"errors"

	. "github.com/brianrahadi/sfucourses-api/internal/model"
	utils "github.com/brianrahadi/sfucourses-api/internal/utils"
)

type OutlineStore struct {
}

func (s *OutlineStore) GetAll(ctx context.Context) ([]CourseOutline, error) {
	outlines, err := utils.ReadCoursesFromJSON[[]CourseOutline]("./json/outlines/outline.json")
	if err != nil {
		return []CourseOutline{}, errors.New("not found")
	}

	return outlines, nil
}
