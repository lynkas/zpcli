package service

import (
	"context"
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strconv"
	"strings"
	"time"
	"zpcli/internal/domain"
	"zpcli/internal/logx"
	"zpcli/store"
)

type DetailService struct {
	Client *http.Client
}

type detailResponse struct {
	List []struct {
		VodName        string `json:"vod_name"`
		VodSub         string `json:"vod_sub"`
		TypeName       string `json:"type_name"`
		VodTag         string `json:"vod_tag"`
		VodClass       string `json:"vod_class"`
		VodActor       string `json:"vod_actor"`
		VodDirector    string `json:"vod_director"`
		VodArea        string `json:"vod_area"`
		VodLang        string `json:"vod_lang"`
		VodYear        string `json:"vod_year"`
		VodHits        int    `json:"vod_hits"`
		VodScore       string `json:"vod_score"`
		VodDoubanScore string `json:"vod_douban_score"`
		VodTime        string `json:"vod_time"`
		VodPubDate     string `json:"vod_pubdate"`
		VodTotal       int    `json:"vod_total"`
		VodContent     string `json:"vod_content"`
		VodPlayURL     string `json:"vod_play_url"`
		VodPlayFrom    string `json:"vod_play_from"`
		VodRemarks     string `json:"vod_remarks"`
	} `json:"list"`
}

func NewDetailService(client *http.Client) *DetailService {
	if client == nil {
		client = &http.Client{Timeout: 10 * time.Second}
	}
	return &DetailService{Client: client}
}

func (s *DetailService) GetDetail(ctx context.Context, data *store.StoreData, siteID string, vodID string) (*domain.DetailResult, error) {
	logger := logx.Logger("service.detail")
	logger.Info("get detail start", "site_id", siteID, "vod_id", vodID, "has_data", data != nil)
	if data == nil {
		logger.Error("get detail missing store data")
		return nil, fmt.Errorf("store data is required")
	}

	seriesIndex, domainIndex, err := ParseDomainID(siteID)
	if err != nil {
		logger.Error("parse site id failed", "site_id", siteID, "error", err)
		return nil, err
	}

	if seriesIndex < 0 || seriesIndex >= len(data.Series) || domainIndex < 0 || domainIndex >= len(data.Series[seriesIndex].Domains) {
		logger.Warn("get detail invalid site id", "site_id", siteID, "series_index", seriesIndex, "domain_index", domainIndex)
		return nil, fmt.Errorf("invalid siteId: %s", siteID)
	}

	domainConfig := data.Series[seriesIndex].Domains[domainIndex]
	reqURL := fmt.Sprintf("%s?ac=detail&ids=%s", BuildEndpointURL(domainConfig.URL), vodID)
	logger.Info("detail http request", "site_id", siteID, "vod_id", vodID, "domain_url", domainConfig.URL, "request_url", reqURL)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		logger.Error("build detail request failed", "request_url", reqURL, "error", err)
		return nil, err
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		logger.Error("detail http request failed", "request_url", reqURL, "error", err)
		return &domain.DetailResult{
			SeriesIndex: seriesIndex,
			DomainIndex: domainIndex,
			DomainURL:   domainConfig.URL,
			Err:         err,
		}, nil
	}
	defer resp.Body.Close()
	logger.Info("detail http response", "request_url", reqURL, "status_code", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		logger.Warn("detail non-200 response", "request_url", reqURL, "status_code", resp.StatusCode)
		return &domain.DetailResult{
			SeriesIndex: seriesIndex,
			DomainIndex: domainIndex,
			DomainURL:   domainConfig.URL,
			Err:         fmt.Errorf("HTTP %d", resp.StatusCode),
		}, nil
	}

	var payload detailResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		logger.Error("decode detail response failed", "request_url", reqURL, "error", err)
		return &domain.DetailResult{
			SeriesIndex: seriesIndex,
			DomainIndex: domainIndex,
			DomainURL:   domainConfig.URL,
			Err:         err,
		}, nil
	}
	logger.Debug("detail decoded response", "request_url", reqURL, "payload", payload)

	if len(payload.List) == 0 {
		logger.Info("detail no items found", "site_id", siteID, "vod_id", vodID)
		return &domain.DetailResult{
			SeriesIndex: seriesIndex,
			DomainIndex: domainIndex,
			DomainURL:   domainConfig.URL,
		}, nil
	}

	item := payload.List[0]
	result := &domain.DetailResult{
		SeriesIndex: seriesIndex,
		DomainIndex: domainIndex,
		DomainURL:   domainConfig.URL,
		Item: &domain.DetailItem{
			VodName:        item.VodName,
			VodSub:         item.VodSub,
			TypeName:       item.TypeName,
			VodTag:         item.VodTag,
			VodClass:       item.VodClass,
			VodActor:       item.VodActor,
			VodDirector:    item.VodDirector,
			VodArea:        item.VodArea,
			VodLang:        item.VodLang,
			VodYear:        item.VodYear,
			VodHits:        item.VodHits,
			VodScore:       item.VodScore,
			VodDoubanScore: item.VodDoubanScore,
			VodTime:        item.VodTime,
			VodPubDate:     item.VodPubDate,
			VodTotal:       item.VodTotal,
			VodContent:     StripHTML(item.VodContent),
			VodRemarks:     item.VodRemarks,
			Players:        parsePlayers(item.VodPlayFrom, item.VodPlayURL),
		},
	}
	logger.Info("get detail complete", "site_id", siteID, "vod_id", vodID, "player_count", len(result.Item.Players))
	logger.Debug("get detail result", "result", result)
	return result, nil
}

func ParseDomainID(id string) (int, int, error) {
	logger := logx.Logger("service.detail")
	parts := strings.Split(id, ".")
	if len(parts) != 2 {
		logger.Warn("parse domain id invalid format", "id", id)
		return 0, 0, fmt.Errorf("invalid siteId format (expected x.y)")
	}
	seriesIndex, err := strconv.Atoi(parts[0])
	if err != nil {
		logger.Warn("parse domain id invalid series", "id", id, "error", err)
		return 0, 0, err
	}
	domainIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		logger.Warn("parse domain id invalid domain", "id", id, "error", err)
		return 0, 0, err
	}
	logger.Debug("parse domain id complete", "id", id, "series_index", seriesIndex-1, "domain_index", domainIndex-1)
	return seriesIndex - 1, domainIndex - 1, nil
}

func StripHTML(s string) string {
	s = html.UnescapeString(s)
	r := strings.NewReplacer(
		"<p>", "",
		"</p>", "\n",
		"<br>", "\n",
		"<br/>", "\n",
		"<br />", "\n",
	)
	s = r.Replace(s)

	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

func FindEpisodeURL(item *domain.DetailItem, target string) (string, bool) {
	logger := logx.Logger("service.detail")
	logger.Info("find episode url start", "target", target, "has_item", item != nil)
	if item == nil {
		logger.Warn("find episode url missing detail item", "target", target)
		return "", false
	}
	for _, player := range item.Players {
		for i, episode := range player.Episodes {
			index := strconv.Itoa(i + 1)
			if target == index || strings.Contains(episode.Name, target) {
				logger.Info("find episode url complete", "target", target, "player", player.Name, "episode_name", episode.Name, "url", episode.URL)
				return episode.URL, true
			}
		}
	}
	logger.Info("find episode url no match", "target", target)
	return "", false
}

func parsePlayers(playFrom string, playURL string) []domain.DetailPlayer {
	logger := logx.Logger("service.detail")
	playerNames := strings.Split(playFrom, "$$$")
	groups := strings.Split(playURL, "$$$")
	players := make([]domain.DetailPlayer, 0, len(groups))

	for i, group := range groups {
		playerName := fmt.Sprintf("Player %d", i+1)
		if i < len(playerNames) && playerNames[i] != "" {
			playerName = playerNames[i]
		}

		episodes := strings.Split(group, "#")
		player := domain.DetailPlayer{
			Name:     playerName,
			Episodes: make([]domain.DetailEpisode, 0, len(episodes)),
		}

		for _, episode := range episodes {
			if episode == "" {
				continue
			}
			parts := strings.Split(episode, "$")
			name := episode
			url := episode
			if len(parts) == 2 {
				name = parts[0]
				url = parts[1]
			}
			player.Episodes = append(player.Episodes, domain.DetailEpisode{
				Name: name,
				URL:  url,
			})
		}

		players = append(players, player)
	}

	logger.Debug("parse players complete", "play_from", playFrom, "play_url", playURL, "players", players)
	return players
}
