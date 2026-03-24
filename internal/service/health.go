package service

import (
	"fmt"
	"net/url"
	"strings"
	"zpcli/internal/domain"
	"zpcli/store"
)

type HealthService struct{}

func NewHealthService() *HealthService {
	return &HealthService{}
}

func (s *HealthService) BuildHealthReport(data *store.StoreData) (*domain.HealthReport, error) {
	configPath, err := store.ConfigFilePath()
	if err != nil {
		return nil, err
	}

	report := &domain.HealthReport{
		ConfigPath: configPath,
	}
	if data == nil {
		return report, nil
	}

	report.Version = data.Version
	report.SeriesCount = len(data.Series)
	for _, series := range data.Series {
		if series == nil {
			continue
		}
		report.DomainCount += len(series.Domains)
	}

	report.Issues = s.ValidateStore(data)
	for _, issue := range report.Issues {
		if issue.Level == "error" {
			report.InvalidCount++
		}
		if issue.Level == "warning" {
			report.WarningCount++
		}
	}

	return report, nil
}

func (s *HealthService) ValidateStore(data *store.StoreData) []domain.ValidationIssue {
	if data == nil {
		return []domain.ValidationIssue{{
			Level:   "error",
			Scope:   "config",
			Message: "store data is nil",
		}}
	}

	var issues []domain.ValidationIssue
	seen := make(map[string]string)

	for i, series := range data.Series {
		seriesID := i + 1
		if series == nil {
			issues = append(issues, domain.ValidationIssue{
				Level:   "error",
				Scope:   "series",
				SiteID:  fmt.Sprintf("%d", seriesID),
				Message: "series entry is nil",
			})
			continue
		}
		if len(series.Domains) == 0 {
			issues = append(issues, domain.ValidationIssue{
				Level:   "warning",
				Scope:   "series",
				SiteID:  fmt.Sprintf("%d", seriesID),
				Message: "series has no configured domains",
			})
		}

		for j, dom := range series.Domains {
			siteID := fmt.Sprintf("%d.%d", seriesID, j+1)
			if dom == nil {
				issues = append(issues, domain.ValidationIssue{
					Level:   "error",
					Scope:   "domain",
					SiteID:  siteID,
					Message: "domain entry is nil",
				})
				continue
			}

			normalizedURL := strings.TrimSpace(dom.URL)
			if normalizedURL == "" {
				issues = append(issues, domain.ValidationIssue{
					Level:   "error",
					Scope:   "domain",
					SiteID:  siteID,
					Message: "domain URL is empty",
				})
				continue
			}

			endpointURL := BuildEndpointURL(normalizedURL)
			parsed, err := url.Parse(endpointURL)
			if err != nil || parsed.Host == "" {
				issues = append(issues, domain.ValidationIssue{
					Level:   "error",
					Scope:   "domain",
					SiteID:  siteID,
					Message: fmt.Sprintf("invalid domain URL: %s", dom.URL),
				})
			}

			if previous, ok := seen[normalizedURL]; ok {
				issues = append(issues, domain.ValidationIssue{
					Level:   "warning",
					Scope:   "domain",
					SiteID:  siteID,
					Message: fmt.Sprintf("duplicate domain URL also used by %s", previous),
				})
			} else {
				seen[normalizedURL] = siteID
			}
		}
	}

	return issues
}
