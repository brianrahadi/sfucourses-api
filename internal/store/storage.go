package store

import (
	"context"
	"database/sql"
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
	}

	Sections interface {
		GetByTerm(context.Context, string, string) ([]model.CourseWithSectionDetails, error)
		GetByTermAndDept(context.Context, string, string, string) ([]model.CourseWithSectionDetails, error)
		GetByTermAndDeptAndNumber(context.Context, string, string, string, string) (model.CourseWithSectionDetails, error)
	}

	SectionsWithOutlines interface {
		GetByTerm(context.Context, string, string) ([]model.CourseOutlineWithSectionDetails, error)
		GetByTermAndDept(context.Context, string, string, string) ([]model.CourseOutlineWithSectionDetails, error)
		GetByTermAndDeptAndNumber(context.Context, string, string, string, string) (model.CourseOutlineWithSectionDetails, error)
	}
}

func NewStorage(db *sql.DB) Storage {
	outlines, err := NewOutlineStore(db)
	if err != nil {
		log.Fatal("Error loading outlines store")
		outlines = &OutlineStore{db: db}
	}
	sections, err := NewSectionStore(db)
	if err != nil {
		log.Fatal("Error loading sections store")
		sections = &SectionsStore{db: db}
	}

	sectionsWithOutline, err := NewSectionsWithOutlineStore(db)
	if err != nil {
		log.Fatal("Error loading sections store")
		sectionsWithOutline = &SectionsWithOutlineStore{db: db}
	}
	return Storage{
		Outlines:             outlines,
		Sections:             sections,
		SectionsWithOutlines: sectionsWithOutline,
	}
}
