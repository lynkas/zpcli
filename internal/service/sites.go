package service

import (
	"fmt"
	"strconv"
	"strings"
	"zpcli/internal/domain"
	"zpcli/store"
)

type SiteService struct{}

func NewSiteService() *SiteService {
	return &SiteService{}
}

func (s *SiteService) ListSites(data *store.StoreData) []domain.SiteSeries {
	if data == nil {
		return nil
	}

	seriesList := make([]domain.SiteSeries, 0, len(data.Series))
	for i, series := range data.Series {
		siteSeries := domain.SiteSeries{
			SeriesID: i + 1,
			Domains:  make([]domain.SiteRecord, 0, len(series.Domains)),
		}
		for j, dom := range series.Domains {
			siteSeries.Domains = append(siteSeries.Domains, domain.SiteRecord{
				SeriesID:     i + 1,
				DomainID:     j + 1,
				URL:          dom.URL,
				FailureScore: dom.FailureScore,
			})
		}
		seriesList = append(seriesList, siteSeries)
	}

	return seriesList
}

func (s *SiteService) AddSite(data *store.StoreData, args ...string) (string, error) {
	if data == nil {
		return "", fmt.Errorf("store data is required")
	}
	if len(args) == 0 || len(args) > 2 {
		return "", fmt.Errorf("expected 1 or 2 arguments")
	}

	if len(args) == 1 {
		domain := args[0]
		if err := data.CreateSeries(domain); err != nil {
			return "", err
		}
		return fmt.Sprintf("Successfully created new series for domain %s", domain), nil
	}

	seriesID, err := strconv.Atoi(args[0])
	if err != nil {
		return "", fmt.Errorf("invalid series ID: %v", err)
	}
	domain := args[1]
	if err := data.AddDomainToSeries(seriesID-1, domain); err != nil {
		return "", err
	}
	return fmt.Sprintf("Successfully added domain %s to series %d", domain, seriesID), nil
}

func (s *SiteService) RemoveSite(data *store.StoreData, id string) (string, error) {
	if data == nil {
		return "", fmt.Errorf("store data is required")
	}

	if strings.Contains(id, ".") {
		if err := data.RemoveDomain(id); err != nil {
			return "", err
		}
		return fmt.Sprintf("Successfully removed domain %s", id), nil
	}

	seriesID, err := strconv.Atoi(id)
	if err != nil {
		return "", fmt.Errorf("invalid id format: %v", err)
	}
	if err := data.RemoveSeries(seriesID - 1); err != nil {
		return "", err
	}
	return fmt.Sprintf("Successfully removed series %d", seriesID), nil
}
