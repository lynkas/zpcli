package service

import (
	"fmt"
	"strconv"
	"strings"
	"zpcli/internal/domain"
	"zpcli/internal/logx"
	"zpcli/store"
)

type SiteService struct{}

func NewSiteService() *SiteService {
	return &SiteService{}
}

func (s *SiteService) ListSites(data *store.StoreData) []domain.SiteSeries {
	logger := logx.Logger("service.sites")
	logger.Info("list sites start", "has_data", data != nil)
	if data == nil {
		logger.Warn("list sites called with nil data")
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

	logger.Info("list sites complete", "series_count", len(seriesList))
	logger.Debug("list sites result", "series", seriesList)
	return seriesList
}

func (s *SiteService) AddSite(data *store.StoreData, args ...string) (string, error) {
	logger := logx.Logger("service.sites")
	logger.Info("add site start", "args", args)
	if data == nil {
		logger.Error("add site missing store data")
		return "", fmt.Errorf("store data is required")
	}
	if len(args) == 0 || len(args) > 2 {
		logger.Warn("add site invalid arg count", "arg_count", len(args))
		return "", fmt.Errorf("expected 1 or 2 arguments")
	}

	if len(args) == 1 {
		domain := args[0]
		if err := data.CreateSeries(domain); err != nil {
			logger.Error("add site create series failed", "domain", domain, "error", err)
			return "", err
		}
		message := fmt.Sprintf("Successfully created new series for domain %s", domain)
		logger.Info("add site complete", "message", message)
		return message, nil
	}

	seriesID, err := strconv.Atoi(args[0])
	if err != nil {
		logger.Warn("add site invalid series id", "series_id", args[0], "error", err)
		return "", fmt.Errorf("invalid series ID: %v", err)
	}
	domain := args[1]
	if err := data.AddDomainToSeries(seriesID-1, domain); err != nil {
		logger.Error("add site append domain failed", "series_id", seriesID, "domain", domain, "error", err)
		return "", err
	}
	message := fmt.Sprintf("Successfully added domain %s to series %d", domain, seriesID)
	logger.Info("add site complete", "message", message)
	return message, nil
}

func (s *SiteService) RemoveSite(data *store.StoreData, id string) (string, error) {
	logger := logx.Logger("service.sites")
	logger.Info("remove site start", "id", id)
	if data == nil {
		logger.Error("remove site missing store data")
		return "", fmt.Errorf("store data is required")
	}

	if strings.Contains(id, ".") {
		if err := data.RemoveDomain(id); err != nil {
			logger.Error("remove domain failed", "id", id, "error", err)
			return "", err
		}
		message := fmt.Sprintf("Successfully removed domain %s", id)
		logger.Info("remove site complete", "message", message)
		return message, nil
	}

	seriesID, err := strconv.Atoi(id)
	if err != nil {
		logger.Warn("remove site invalid id", "id", id, "error", err)
		return "", fmt.Errorf("invalid id format: %v", err)
	}
	if err := data.RemoveSeries(seriesID - 1); err != nil {
		logger.Error("remove series failed", "series_id", seriesID, "error", err)
		return "", err
	}
	message := fmt.Sprintf("Successfully removed series %d", seriesID)
	logger.Info("remove site complete", "message", message)
	return message, nil
}
