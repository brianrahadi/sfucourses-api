package store

import (
	"context"
	"errors"
	"log"

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

	Courses interface {
		GetByTerm(context.Context, string, string) ([]CourseWithSectionDetails, error)
		GetByTermAndDept(context.Context, string, string, string) ([]CourseWithSectionDetails, error)
		GetByTermAndDeptAndNumber(context.Context, string, string, string, string) (CourseWithSectionDetails, error)
	}
}

func NewStorage() Storage {
	outlines, err := NewOutlineStore()
	if err != nil {
		log.Fatal("Error loading outlines store")
		outlines = &OutlineStore{}
	}
	courses, err := NewCourseStore()
	if err != nil {
		log.Fatal("Error loading courses store")
		courses = &CoursesStore{}
	}
	return Storage{
		Outlines: outlines,
		Courses:  courses,
	}
}
