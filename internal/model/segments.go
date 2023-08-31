package model

import "time"

type Segment struct {
	Id   int    `json:"id"`
	Slug string `json:"slug"`
}

type UserSegment struct {
	Id         int       `json:"id,omitempty"`
	User       *User     `json:"user,omitempty"`
	Segment    *Segment  `json:"segment,omitempty"`
	ActiveFrom time.Time `json:"activeFrom"`
	ActiveTo   time.Time `json:"activeTo"`
}

type SegmentAutoInsert struct {
	Id         int       `json:"id"`
	Segment    *Segment  `json:"segment"`
	Chance     float64   `json:"chance"`
	ActiveFrom time.Time `json:"activeFrom"`
	ActiveTo   time.Time `json:"activeTo"`
}
