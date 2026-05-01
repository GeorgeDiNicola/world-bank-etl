package model

type Topic struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type Indicator struct {
	ID     string  `json:"id"`
	Name   string  `json:"name"`
	Topics []Topic `json:"topics"`
}

type PageMetadata struct {
	Page  int `json:"page"`
	Pages int `json:"pages"`
	Total int `json:"total"`
}
