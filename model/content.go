package model

import "time"

type Content struct {
	ID               int64      `json:"id"`
	ExecutionID      int64      `json:"execution_id"`
	Title            *string    `json:"title"`
	ShortDescription *string    `json:"short_description"`
	Message          *string    `json:"message"`
	Status           *string    `json:"status"`
Category         *string    `json:"category"`
	SubCategory      *string    `json:"sub_category"`
	ImageURL         *string    `json:"image_url"`
	ImagePrompt      *string    `json:"image_prompt"`
	Slug             *string    `json:"slug"`
	Created          *time.Time `json:"created"`
	LastUpdated      *time.Time `json:"last_updated"`
}

type SectionItem struct {
	ID               int64  `json:"id"`
	Title            string `json:"title"`
	Slug             string `json:"slug"`
	ShortDescription string `json:"short_description"`
	Message          string `json:"message"`
}

type Section struct {
	Name  string        `json:"name"`
	Items []SectionItem `json:"items"`
}

type SectionsResponse struct {
	Sections []Section `json:"sections"`
}
