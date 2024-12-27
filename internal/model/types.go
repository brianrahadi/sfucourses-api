package model

// CourseOutline represents the general information about a course
type CourseOutline struct {
	Dept           string   `json:"dept"`
	Number         string   `json:"number"`
	Title          string   `json:"title"`
	Units          string   `json:"units"`
	Description    string   `json:"description"`
	Notes          string   `json:"notes"`
	Designation    string   `json:"designation"`
	DeliveryMethod string   `json:"deliveryMethod"`
	Prerequisites  string   `json:"prerequisites"`
	Corequisites   string   `json:"corequisites"`
	DegreeLevel    string   `json:"degreeLevel"`
	Terms          []string `json:"terms"`
}

type SectionInfo struct {
	Dept           string `json:"dept"`           // CMPT
	Number         string `json:"number"`         // 225
	Section        string `json:"section"`        // D100
	Term           string `json:"term"`           // Fall 2024
	OutlinePath    string `json:"outlinePath"`    // 2024/fall/cmpt/225/d100
	DeliveryMethod string `json:"deliveryMethod"` // In Person
	ClassNumber    string `json:"classNumber"`    // 6327
}

type SectionInstructor struct {
	Name  string `json:"name"` // John Doe
	Email string `json:"email"`
}

type SectionSchedule struct {
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Campus      string `json:"campus"`
	Days        string `json:"days"`
	StartTime   string `json:"startTime"`
	EndTime     string `json:"endTime"`
	SectionCode string `json:"sectionCode"`
}

// for json read from sfu courses API
type SectionDetailRaw struct {
	Info           SectionInfo         `json:"info"`
	Instructor     []SectionInstructor `json:"instructor"`     // singular for parsing
	CourseSchedule []SectionSchedule   `json:"courseSchedule"` //
}

type SectionDetail struct {
	Section        string              `json:"section"`        // D100
	OutlinePath    string              `json:"outlinePath"`    // 2024/fall/cmpt/225/d100
	DeliveryMethod string              `json:"deliveryMethod"` // In Person
	ClassNumber    string              `json:"classNumber"`    // 6327
	Instructors    []SectionInstructor `json:"instructors"`
	Schedules      []SectionSchedule   `json:"schedules"`
}

// for json write
type CourseWithSectionDetails struct {
	Dept           string          `json:"dept"`   // CMPT
	Number         string          `json:"number"` // 225
	Term           string          `json:"term"`   // Fall 2024
	SectionDetails []SectionDetail `json:"sections"`
}

type CourseOutlinesResponse struct {
	Data       []CourseOutline `json:"data"`
	TotalCount int             `json:"total_count"`
	NextURL    string          `json:"next_url,omitempty"`
}
