package model

type Topic struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type Source struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type Indicator struct {
	ID                 string  `json:"id"`
	Name               string  `json:"name"`
	Source             Source  `json:"source"`
	SourceOrganization string  `json:"sourceOrganization"`
	Topics             []Topic `json:"topics"`
}

type PageMetadata struct {
	Page  int `json:"page"`
	Pages int `json:"pages"`
	Total int `json:"total"`
}

type Observation struct {
	Indicator struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	} `json:"indicator"`
	Country struct {
		ID    string `json:"id"`
		Value string `json:"value"`
	} `json:"country"`
	CountryISO3 string   `json:"countryiso3code"`
	Date        string   `json:"date"`
	Value       *float64 `json:"value"` // Pointer to handle null values
	Unit        string   `json:"unit"`
	ObsStatus   string   `json:"obs_status"`
	Decimal     int      `json:"decimal"`
}
