package store

import (
	"context"
	"errors"

	. "github.com/brianrahadi/sfucourses-api/internal/model"
)

var (
	ErrNotFound = errors.New("resource not found")
)

type Storage struct {
	Outlines interface {
		GetAll(context.Context) ([]CourseOutline, error)
		GetByDept(context.Context, string) ([]CourseOutline, error)
		GetByDeptAndNumber(context.Context, string, string) (CourseOutline, error)
	}

	// Schedules interface {
	// 	GetByTerm(context.Context, year string, term string) ([]CourseWithSectionDetails)
	// }
}

func NewStorage() Storage {
	return Storage{
		Outlines: &OutlineStore{},
	}
}
