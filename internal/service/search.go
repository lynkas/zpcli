package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"
	"zpcli/internal/domain"
	"zpcli/internal/logx"
	"zpcli/store"
)

type SearchService struct {
	Client *http.Client
}

type searchResponse struct {
	List []struct {
		VodID      int    `json:"vod_id"`
		VodName    string `json:"vod_name"`
		TypeName   string `json:"type_name"`
		VodTime    string `json:"vod_time"`
		VodRemarks string `json:"vod_remarks"`
	} `json:"list"`
}

func NewSearchService(client *http.Client) *SearchService {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &SearchService{Client: client}
}

func (s *SearchService) Search(ctx context.Context, data *store.StoreData, keyword string, seriesCount int, page int) ([]domain.SearchResult, error) {
	logger := logx.Logger("service.search")
	logger.Info("search start", "keyword", keyword, "series_count", seriesCount, "page", page, "has_data", data != nil)
	if data == nil {
		logger.Error("search missing store data")
		return nil, fmt.Errorf("store data is required")
	}

	type target struct {
		SeriesIndex int
		DomainIndex int
		Domain      *store.Domain
	}

	type seriesRank struct {
		sIdx     int
		dIdx     int
		minScore int
	}

	var ranks []seriesRank
	for i, series := range data.Series {
		if len(series.Domains) == 0 {
			continue
		}
		bestDIdx := 0
		minScore := series.Domains[0].FailureScore
		for j, dom := range series.Domains {
			if dom.FailureScore < minScore {
				minScore = dom.FailureScore
				bestDIdx = j
			}
		}
		ranks = append(ranks, seriesRank{sIdx: i, dIdx: bestDIdx, minScore: minScore})
	}

	sort.Slice(ranks, func(i, j int) bool {
		return ranks[i].minScore < ranks[j].minScore
	})
	logger.Debug("search ranked series", "ranks", ranks)

	limit := seriesCount
	if limit > len(ranks) {
		limit = len(ranks)
	}
	if limit == 0 {
		logger.Info("search complete with no targets", "keyword", keyword)
		return nil, nil
	}

	targets := make([]target, 0, limit)
	for i := 0; i < limit; i++ {
		r := ranks[i]
		targets = append(targets, target{
			SeriesIndex: r.sIdx,
			DomainIndex: r.dIdx,
			Domain:      data.Series[r.sIdx].Domains[r.dIdx],
		})
	}

	results := make(chan domain.SearchResult, len(targets))
	var wg sync.WaitGroup

	for _, t := range targets {
		wg.Add(1)
		go func(t target) {
			defer wg.Done()

			reqURL := fmt.Sprintf("%s?ac=list&wd=%s&pg=%d", BuildEndpointURL(t.Domain.URL), url.QueryEscape(keyword), page)
			logger.Info("search http request", "series_index", t.SeriesIndex, "domain_index", t.DomainIndex, "domain_url", t.Domain.URL, "request_url", reqURL)
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
			if err != nil {
				logger.Error("search build request failed", "request_url", reqURL, "error", err)
				results <- domain.SearchResult{
					SeriesIndex: t.SeriesIndex,
					DomainIndex: t.DomainIndex,
					DomainURL:   t.Domain.URL,
					Score:       t.Domain.FailureScore,
					Err:         err,
				}
				return
			}

			resp, err := s.Client.Do(req)
			if err != nil {
				logger.Error("search http request failed", "request_url", reqURL, "error", err)
				results <- domain.SearchResult{
					SeriesIndex: t.SeriesIndex,
					DomainIndex: t.DomainIndex,
					DomainURL:   t.Domain.URL,
					Score:       t.Domain.FailureScore,
					Err:         err,
				}
				return
			}
			defer resp.Body.Close()
			logger.Info("search http response", "request_url", reqURL, "status_code", resp.StatusCode)

			if resp.StatusCode != http.StatusOK {
				logger.Warn("search non-200 response", "request_url", reqURL, "status_code", resp.StatusCode)
				results <- domain.SearchResult{
					SeriesIndex: t.SeriesIndex,
					DomainIndex: t.DomainIndex,
					DomainURL:   t.Domain.URL,
					Score:       t.Domain.FailureScore,
					Err:         fmt.Errorf("HTTP %d", resp.StatusCode),
				}
				return
			}

			var payload searchResponse
			if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
				logger.Error("search decode response failed", "request_url", reqURL, "error", err)
				results <- domain.SearchResult{
					SeriesIndex: t.SeriesIndex,
					DomainIndex: t.DomainIndex,
					DomainURL:   t.Domain.URL,
					Score:       t.Domain.FailureScore,
					Err:         err,
				}
				return
			}
			logger.Debug("search decoded response", "request_url", reqURL, "payload", payload)

			items := make([]domain.SearchItem, 0, len(payload.List))
			for _, item := range payload.List {
				items = append(items, domain.SearchItem{
					VodID:      item.VodID,
					VodName:    item.VodName,
					TypeName:   item.TypeName,
					VodTime:    item.VodTime,
					VodRemarks: item.VodRemarks,
				})
			}

			logger.Info("search target complete", "request_url", reqURL, "item_count", len(items))
			logger.Debug("search target result", "request_url", reqURL, "items", items)
			results <- domain.SearchResult{
				SeriesIndex: t.SeriesIndex,
				DomainIndex: t.DomainIndex,
				DomainURL:   t.Domain.URL,
				Score:       t.Domain.FailureScore,
				Items:       items,
			}
		}(t)
	}

	wg.Wait()
	close(results)

	collected := make([]domain.SearchResult, 0, len(targets))
	for result := range results {
		collected = append(collected, result)
	}
	logger.Info("search complete", "keyword", keyword, "result_count", len(collected))
	logger.Debug("search results", "results", collected)

	return collected, nil
}
