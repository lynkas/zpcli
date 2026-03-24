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
	if data == nil {
		return nil, fmt.Errorf("store data is required")
	}

	seriesIndex, domainIndex, err := ParseDomainID(siteID)
	if err != nil {
		return nil, err
	}

	if seriesIndex < 0 || seriesIndex >= len(data.Series) || domainIndex < 0 || domainIndex >= len(data.Series[seriesIndex].Domains) {
		return nil, fmt.Errorf("invalid siteId: %s", siteID)
	}

	domainConfig := data.Series[seriesIndex].Domains[domainIndex]
	reqURL := fmt.Sprintf("%s?ac=detail&ids=%s", BuildEndpointURL(domainConfig.URL), vodID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.Client.Do(req)
	if err != nil {
		return &domain.DetailResult{
			SeriesIndex: seriesIndex,
			DomainIndex: domainIndex,
			DomainURL:   domainConfig.URL,
			Err:         err,
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &domain.DetailResult{
			SeriesIndex: seriesIndex,
			DomainIndex: domainIndex,
			DomainURL:   domainConfig.URL,
			Err:         fmt.Errorf("HTTP %d", resp.StatusCode),
		}, nil
	}

	var payload detailResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return &domain.DetailResult{
			SeriesIndex: seriesIndex,
			DomainIndex: domainIndex,
			DomainURL:   domainConfig.URL,
			Err:         err,
		}, nil
	}

	if len(payload.List) == 0 {
		return &domain.DetailResult{
			SeriesIndex: seriesIndex,
			DomainIndex: domainIndex,
			DomainURL:   domainConfig.URL,
		}, nil
	}

	item := payload.List[0]
	return &domain.DetailResult{
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
	}, nil
}

func ParseDomainID(id string) (int, int, error) {
	parts := strings.Split(id, ".")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid siteId format (expected x.y)")
	}
	seriesIndex, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	domainIndex, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
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
	if item == nil {
		return "", false
	}
	for _, player := range item.Players {
		for i, episode := range player.Episodes {
			index := strconv.Itoa(i + 1)
			if target == index || strings.Contains(episode.Name, target) {
				return episode.URL, true
			}
		}
	}
	return "", false
}

func parsePlayers(playFrom string, playURL string) []domain.DetailPlayer {
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

	return players
}
