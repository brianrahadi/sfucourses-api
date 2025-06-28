package store

import (
	"context"
	"errors"
	"log"

	"github.com/brianrahadi/sfucourses-api/internal/model"
	"github.com/samber/mo"
)

var (
	ErrNotFound = errors.New("resource not found")
)

type Storage struct {
	Outlines interface {
		GetAll(context.Context, mo.Option[int], mo.Option[int]) ([]model.CourseOutline, int, error)
		GetByDept(context.Context, string) ([]model.CourseOutline, error)
		GetByDeptAndNumber(context.Context, string, string) (model.CourseOutline, error)
		ForceReload() error
	}

	Sections interface {
		GetByTerm(context.Context, string, string) ([]model.CourseWithSectionDetails, error)
		GetByTermAndDept(context.Context, string, string, string) ([]model.CourseWithSectionDetails, error)
		GetByTermAndDeptAndNumber(context.Context, string, string, string, string) (model.CourseWithSectionDetails, error)
		ForceReload() error
	}

	SectionsWithOutlines interface {
		GetByTerm(context.Context, string, string) ([]model.CourseOutlineWithSectionDetails, error)
		GetByTermAndDept(context.Context, string, string, string) ([]model.CourseOutlineWithSectionDetails, error)
		GetByTermAndDeptAndNumber(context.Context, string, string, string, string) (model.CourseOutlineWithSectionDetails, error)
		ForceReload() error
	}

	Instructors interface {
		GetAll(context.Context) ([]model.InstructorResponse, error)
		GetByDept(context.Context, string) ([]model.InstructorResponse, error)
		GetByDeptAndNumber(context.Context, string, string) ([]model.InstructorResponse, error)
		ForceReload() error
	}
}

func NewStorage() Storage {
	outlines, err := NewOutlineStore()

	if err != nil {
		log.Fatal("Error loading outlines store")
		outlines = &OutlineStore{}
	}
	sections, err := NewSectionStore()
	if err != nil {
		log.Fatal("Error loading sections store")
		sections = &SectionsStore{}
	}

	sectionsWithOutline, err := NewSectionsWithOutlineStore()
	if err != nil {
		log.Fatal("Error loading sections store")
		sectionsWithOutline = &SectionsWithOutlineStore{}
	}

	instructors, err := NewInstructorStore()
	if err != nil {
		log.Fatal("Error loading instructors store")
		instructors = &InstructorStore{}
	}

	return Storage{
		Outlines:             outlines,
		Sections:             sections,
		SectionsWithOutlines: sectionsWithOutline,
		Instructors:          instructors,
	}
}
