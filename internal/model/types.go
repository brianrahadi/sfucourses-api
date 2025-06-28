package model

// CourseOutline represents the general information about a course
// @Description Course outline information including general details and offerings
type CourseOutline struct {
	Dept           string           `json:"dept" example:"CMPT" description:"Department code"`
	Number         string           `json:"number" example:"225" description:"Course number"`
	Title          string           `json:"title" example:"Data Structures and Programming" description:"Course title"`
	Units          string           `json:"units" example:"3" description:"Number of credit units"`
	Description    string           `json:"description" example:"Introduction to a variety of practical and important data structures and methods..." description:"Course description"`
	Notes          string           `json:"notes" example:"" description:"Additional course notes"`
	Designation    string           `json:"designation" example:"Quantitative" description:"Course designations (e.g., Quantitative)"`
	DeliveryMethod string           `json:"deliveryMethod" example:"In Person" description:"Method of course delivery"`
	Prerequisites  string           `json:"prerequisites" example:"(MACM 101 and (CMPT 125, CMPT 129 or CMPT 135)) or (ENSC 251 and ENSC 252), all with a minimum grade of C-." description:"Course prerequisites"`
	Corequisites   string           `json:"corequisites" example:"" description:"Course corequisites"`
	DegreeLevel    string           `json:"degreeLevel" example:"UGRD" description:"Degree level of the course (e.g., UGRD for Undergraduate)"`
	Offerings      []CourseOffering `json:"offerings" description:"List of term offerings for this course"`
}

// CourseWithSectionDetails represents a course with its section details
// @Description Course with detailed section information
type CourseWithSectionDetails struct {
	Dept           string          `json:"dept" example:"CMPT" description:"Department code"`
	Number         string          `json:"number" example:"225" description:"Course number"`
	Title          string          `json:"title" example:"Data Structure and Algorithms" description:"Course title"`
	Units          string          `json:"units" example:"3" description:"Number of credit units"`
	Term           string          `json:"term" example:"Fall 2024" description:"Academic term"`
	SectionDetails []SectionDetail `json:"sections" description:"List of section details"`
}

// CourseOutlineWithSectionDetails represents a course outline with section details
// @Description Course outline with detailed section information
type CourseOutlineWithSectionDetails struct {
	Dept           string          `json:"dept" example:"CMPT" description:"Department code"`
	Number         string          `json:"number" example:"225" description:"Course number"`
	Title          string          `json:"title" example:"Data Structure and Algorithms" description:"Course title"`
	Units          string          `json:"units" example:"3" description:"Number of credit units"`
	Description    string          `json:"description" example:"Introduction to the study of data structures..." description:"Course description"`
	Designation    string          `json:"designation" example:"W,Q" description:"Course designations (W, Q, B, etc.)"`
	DeliveryMethod string          `json:"deliveryMethod" example:"In Person" description:"Method of course delivery"`
	Prerequisites  string          `json:"prerequisites" example:"CMPT 125 and MACM 101" description:"Course prerequisites"`
	Corequisites   string          `json:"corequisites" example:"MACM 201" description:"Course corequisites"`
	Term           string          `json:"term" example:"Fall 2024" description:"Academic term"`
	SectionDetails []SectionDetail `json:"sections" description:"List of section details"`
}

// CourseOffering represents a course offering for a specific term
// @Description Course offering information for a specific term
type CourseOffering struct {
	Instructors []string `json:"instructors" example:"John Doe,Jane Smith" description:"List of instructor names"`
	Term        string   `json:"term" example:"Fall 2024" description:"Academic term of the offering"`
}

// SectionDetail represents detailed information about a course section
// @Description Detailed information about a course section
type SectionDetail struct {
	Section        string            `json:"section" example:"D100" description:"Section code"`
	DeliveryMethod string            `json:"deliveryMethod" example:"In Person" description:"Method of section delivery"`
	ClassNumber    string            `json:"classNumber" example:"6327" description:"Class number for registration"`
	Instructors    []Instructor      `json:"instructors" description:"List of section instructors"`
	Schedules      []SectionSchedule `json:"schedules" description:"List of section schedules"`
}

// Instructor represents an instructor's information
// @Description Instructor information
type Instructor struct {
	Name  string `json:"name" example:"John Doe" description:"Instructor's full name"`
	Email string `json:"email" example:"john_doe@sfu.ca" description:"Instructor's email address"`
}

// InstructorResponse represents instructor information
// @Description Instructor information
type InstructorResponse struct {
	Name      string               `json:"name" example:"John Doe" description:"Instructor's full name"`
	Offerings []InstructorOffering `json:"offerings" description:"List of course offerings"`
}

// InstructorOffering represents an instructor's offering of a course
// @Description Instructor offering information
type InstructorOffering struct {
	Dept   string `json:"dept" example:"CMPT" description:"Department code"`
	Number string `json:"number" example:"225" description:"Course number"`
	Title  string `json:"title" example:"Data Structures and Algorithms" description:"Course title"`
	Term   string `json:"term" example:"Fall 2024" description:"Academic term"`
}

// SectionSchedule represents a section's schedule information
// @Description Schedule information for a section
type SectionSchedule struct {
	StartDate   string `json:"startDate" example:"2024-09-03" description:"Start date of the class"`
	EndDate     string `json:"endDate" example:"2024-12-06" description:"End date of the class"`
	Campus      string `json:"campus" example:"Burnaby" description:"Campus location"`
	Days        string `json:"days" example:"Mo,We,Fr" description:"Days of the week"`
	StartTime   string `json:"startTime" example:"10:30" description:"Start time of the class"`
	EndTime     string `json:"endTime" example:"11:20" description:"End time of the class"`
	SectionCode string `json:"sectionCode" example:"LEC" description:"Section code type (LEC, TUT, LAB)"`
}

// SectionDetailRaw represents raw section detail from sfucourses API
// @Description Raw section detail information from sfucourses API
type SectionDetailRaw struct {
	Info           SectionInfo       `json:"info" description:"Basic section information"`
	Instructor     []Instructor      `json:"instructor" description:"List of instructors (singular for parsing)"`
	CourseSchedule []SectionSchedule `json:"courseSchedule" description:"List of section schedules"`
}

// SectionInfo represents basic section information
// @Description Basic information about a section
type SectionInfo struct {
	Dept           string `json:"dept" example:"CMPT" description:"Department code"`
	Number         string `json:"number" example:"225" description:"Course number"`
	Section        string `json:"section" example:"D100" description:"Section code"`
	Title          string `json:"title" example:"Data Structures and Algorithms" description:"Course title"`
	Units          string `json:"units" example:"3" description:"Number of credit units"`
	Term           string `json:"term" example:"Fall 2024" description:"Academic term"`
	DeliveryMethod string `json:"deliveryMethod" example:"In Person" description:"Method of course delivery"`
	ClassNumber    string `json:"classNumber" example:"6327" description:"Class number for registration"`
}

// AllCourseOutlinesResponse represents a paginated response of course outlines
// @Description Paginated response containing course outlines
type AllCourseOutlinesResponse struct {
	Data       []CourseOutline `json:"data" description:"List of course outlines"`
	TotalCount int             `json:"total_count" example:"3412" description:"Total count of outlines matching the query"`
	NextURL    string          `json:"next_url,omitempty" example:"/v1/rest/outlines?limit=100&offset=50" description:"URL for the next page of results"`
}
