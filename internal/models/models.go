package models

type Segment struct {
	Slug    string `json:"segment_slug"`
	DaysTTL int    `json:"days_ttl,omitempty"`
}

type History struct {
	User       int64
	Segment    Segment
	Action     bool
	ActionTime string
}
