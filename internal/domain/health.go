package domain

type ValidationIssue struct {
	Level   string
	Scope   string
	SiteID  string
	Message string
}

type HealthReport struct {
	ConfigPath    string
	Version       int
	SeriesCount   int
	DomainCount   int
	InvalidCount  int
	WarningCount  int
	Issues        []ValidationIssue
}
