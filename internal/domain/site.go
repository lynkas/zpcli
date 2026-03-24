package domain

type SiteRecord struct {
	SeriesID     int
	DomainID     int
	URL          string
	FailureScore int
}

type SiteSeries struct {
	SeriesID int
	Domains  []SiteRecord
}
