package model

// CourseInfo represents the general information about a course
type CourseInfo struct {
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
	TermsOffered   []string `json:"termsOffered"`
}

// func CourseInfoFromDict(data map[string]interface{}) CourseInfo {
// 	return CourseInfo{
// 		Dept:           stringValue(data["dept"]),
// 		Number:         stringValue(data["number"]),
// 		Title:          stringValue(data["title"]),
// 		Description:    stringValue(data["description"]),
// 		Prerequisites:  stringValue(data["prerequisites"]),
// 		Corequisites:   stringValue(data["corequisites"]),
// 		Notes:          stringValue(data["notes"]),
// 		DeliveryMethod: stringValue(data["deliveryMethod"]),
// 		Units:          stringValue(data["units"]),
// 	}
// }

// // Section represents a complete course section with its schedule
// type Section struct {
// 	Info           SectionInfo      `json:"info"`
// 	CourseSchedule []CourseSchedule `json:"courseSchedule"`
// }

// func SectionFromDict(data map[string]interface{}) Section {
// 	var schedules []CourseSchedule
// 	if courseSchedule, ok := data["courseSchedule"].([]interface{}); ok {
// 		for _, schedule := range courseSchedule {
// 			if scheduleMap, ok := schedule.(map[string]interface{}); ok {
// 				schedules = append(schedules, CourseScheduleFromDict(scheduleMap))
// 			}
// 		}
// 	}

// 	return Section{
// 		Info:           SectionInfoFromDict(data),
// 		CourseSchedule: schedules,
// 	}
// }

// UnmarshalJSON implements the json.Unmarshaler interface for CourseInfo
// func (c *CourseInfo) UnmarshalJSON(data []byte) error {
// 	var rawData map[string]interface{}
// 	if err := json.Unmarshal(data, &rawData); err != nil {
// 		return err
// 	}
// 	*c = CourseInfoFromDict(rawData)
// 	return nil
// }
