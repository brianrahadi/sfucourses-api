package store

import (
	"context"
	"errors"
	"log"

	. "github.com/brianrahadi/sfucourses-api/internal/model"
	"github.com/samber/mo"
)

var (
	ErrNotFound = errors.New("resource not found")
)

type Storage struct {
	Outlines interface {
		GetAll(context.Context, mo.Option[int], mo.Option[int]) ([]CourseOutline, int, error)
		GetByDept(context.Context, string) ([]CourseOutline, error)
		GetByDeptAndNumber(context.Context, string, string) (CourseOutline, error)
	}

	Sections interface {
		GetByTerm(context.Context, string, string) ([]CourseWithSectionDetails, error)
		GetByTermAndDept(context.Context, string, string, string) ([]CourseWithSectionDetails, error)
		GetByTermAndDeptAndNumber(context.Context, string, string, string, string) (CourseWithSectionDetails, error)
	}

	SectionsWithOutlines interface {
		GetByTerm(context.Context, string, string) ([]CourseOutlineWithSectionDetails, error)
		GetByTermAndDept(context.Context, string, string, string) ([]CourseOutlineWithSectionDetails, error)
		GetByTermAndDeptAndNumber(context.Context, string, string, string, string) (CourseOutlineWithSectionDetails, error)
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
	return Storage{
		Outlines:             outlines,
		Sections:             sections,
		SectionsWithOutlines: sectionsWithOutline,
	}
}
