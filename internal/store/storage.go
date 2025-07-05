package store

import (
	"context"
	"errors"
	"log"

	"github.com/brianrahadi/sfucourses-api/internal/model"
)

var (
	ErrNotFound = errors.New("resource not found")
)

type Storage struct {
	Outlines interface {
		Get(context.Context, string, string) ([]model.CourseOutline, error)
		ForceReload() error
	}

	Sections interface {
		Get(context.Context, string, string, string, string) ([]model.CourseWithSectionDetails, error)
		ForceReload() error
	}

	SectionsWithOutlines interface {
		Get(context.Context, string, string, string, string) ([]model.CourseOutlineWithSectionDetails, error)
		ForceReload() error
	}

	Instructors interface {
		Get(context.Context, string, string, string) ([]model.InstructorResponse, error)
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
