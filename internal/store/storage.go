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
