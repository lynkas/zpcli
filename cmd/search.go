package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"
	"zpcli/store"

	"os"

	"github.com/mattn/go-runewidth"
	"github.com/spf13/cobra"
)

var (
	searchSeriesCount int
	searchPage        int
	searchSortBy      string
)

type SearchResponse struct {
	List []struct {
		VodID      int    `json:"vod_id"`
		VodName    string `json:"vod_name"`
		TypeName   string `json:"type_name"`
		VodTime    string `json:"vod_time"`
		VodRemarks string `json:"vod_remarks"`
	} `json:"list"`
}

func formatRelativeTime(vodTime string) string {
	t, err := time.Parse("2006-01-02 15:04:05", vodTime)
	if err != nil {
		return vodTime
	}
	diff := time.Since(t)
	if diff < 0 { // Future date? Just show as is
		return "now"
	}
	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 0 {
			mins := int(diff.Minutes())
			if mins == 0 {
				return "now"
			}
			return fmt.Sprintf("%dm", mins)
		}
		return fmt.Sprintf("%dh", hours)
	}
	days := int(diff.Hours() / 24)
	if days < 30 {
		return fmt.Sprintf("%dd", days)
	}
	months := days / 30
	return fmt.Sprintf("%dmo", months)
}

type QueryTarget struct {
	SeriesIndex int
	DomainIndex int
	Domain      *store.Domain
}

type SearchResult struct {
	Target QueryTarget
	Resp   *SearchResponse
	Err    error
}

var searchCmd = &cobra.Command{
	Use:     "search [keyword]",
	Aliases: []string{"s"},
	Short:   "Search content across top configured domains (alias: s)",
	Args:    cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ShowSearch(os.Stdout, strings.Join(args, " "), searchSeriesCount, searchPage, searchSortBy)
	},
}

func ShowSearch(w io.Writer, keyword string, seriesCount int, page int, sortBy string) {
	s, err := store.Load()
	if err != nil {
		fmt.Fprintf(w, "Error loading store: %v\n", err)
		return
	}

	// Rank series and pick top N
	type seriesRank struct {
		sIdx     int
		dIdx     int
		minScore int
	}

	var ranks []seriesRank
	for i, series := range s.Series {
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

	limit := seriesCount
	if limit > len(ranks) {
		limit = len(ranks)
	}

	if limit == 0 {
		fmt.Fprintln(w, "No valid domains to search.")
		return
	}

	var targets []QueryTarget
	for i := 0; i < limit; i++ {
		r := ranks[i]
		targets = append(targets, QueryTarget{
			SeriesIndex: r.sIdx,
			DomainIndex: r.dIdx,
			Domain:      s.Series[r.sIdx].Domains[r.dIdx],
		})
	}

	// Execute searches concurrently
	var wg sync.WaitGroup
	results := make(chan SearchResult, len(targets))

	client := &http.Client{Timeout: 10 * time.Second}

	for _, target := range targets {
		wg.Add(1)
		go func(t QueryTarget) {
			defer wg.Done()

			baseEndpoint := buildEndpointURL(t.Domain.URL)
			reqURL := fmt.Sprintf("%s?ac=list&wd=%s&pg=%d", baseEndpoint, url.QueryEscape(keyword), page)

			resp, err := client.Get(reqURL)
			if err != nil {
				results <- SearchResult{Target: t, Err: err}
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				results <- SearchResult{Target: t, Err: fmt.Errorf("HTTP %d", resp.StatusCode)}
				return
			}

			var searchResp SearchResponse
			if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
				results <- SearchResult{Target: t, Err: err}
				return
			}

			results <- SearchResult{Target: t, Resp: &searchResp}
		}(target)
	}

	wg.Wait()
	close(results)

	// Check for failures and update scores
	var allResults []SearchResult
	resultsCopy := make(chan SearchResult, len(targets))
	for res := range results {
		allResults = append(allResults, res)
		resultsCopy <- res
	}
	close(resultsCopy)

	hasFailures := false
	for _, res := range allResults {
		if res.Err != nil {
			res.Target.Domain.FailureScore++
			hasFailures = true
		}
	}
	if hasFailures {
		s.Save()
	}

	// Print results
	type row struct {
		domainID string
		score    string
		vodID    string
		typeName string
		vodName  string
		vodTime  string
		remarks  string
		rawTime  string // Used for sorting
	}
	var rows []row

	headers := row{
		domainID: "DOMAIN",
		score:    "SCORE",
		vodID:    "ID",
		typeName: "TYPE",
		vodTime:  "TIME",
		vodName:  "TITLE",
		remarks:  "REMARKS",
	}

	maxDomainIDWidth := runewidth.StringWidth(headers.domainID)
	maxScoreWidth := runewidth.StringWidth(headers.score)
	maxVodIDWidth := runewidth.StringWidth(headers.vodID)
	maxTypeNameWidth := runewidth.StringWidth(headers.typeName)
	maxTimeWidth := runewidth.StringWidth(headers.vodTime)
	maxTitleWidth := runewidth.StringWidth(headers.vodName)

	overlaps := make(map[string]int)

	for res := range resultsCopy {
		domainID := fmt.Sprintf("%d.%d", res.Target.SeriesIndex+1, res.Target.DomainIndex+1)
		score := fmt.Sprintf("%d", res.Target.Domain.FailureScore)
		if res.Err != nil || len(res.Resp.List) == 0 {
			continue
		}

		if runewidth.StringWidth(domainID) > maxDomainIDWidth {
			maxDomainIDWidth = runewidth.StringWidth(domainID)
		}
		if runewidth.StringWidth(score) > maxScoreWidth {
			maxScoreWidth = runewidth.StringWidth(score)
		}

		for _, item := range res.Resp.List {
			vID := fmt.Sprintf("%d", item.VodID)
			relTime := formatRelativeTime(item.VodTime)
			if runewidth.StringWidth(vID) > maxVodIDWidth {
				maxVodIDWidth = runewidth.StringWidth(vID)
			}
			if runewidth.StringWidth(item.TypeName) > maxTypeNameWidth {
				maxTypeNameWidth = runewidth.StringWidth(item.TypeName)
			}
			if runewidth.StringWidth(relTime) > maxTimeWidth {
				maxTimeWidth = runewidth.StringWidth(relTime)
			}
			if runewidth.StringWidth(item.VodName) > maxTitleWidth {
				maxTitleWidth = runewidth.StringWidth(item.VodName)
			}

			rows = append(rows, row{
				domainID: domainID,
				score:    score,
				vodID:    vID,
				typeName: item.TypeName,
				vodName:  item.VodName,
				vodTime:  relTime,
				remarks:  item.VodRemarks,
				rawTime:  item.VodTime,
			})
			overlaps[item.VodName]++
		}
	}

	if len(rows) == 0 {
		fmt.Fprintln(w, "No results found.")
		return
	}

	// Sort rows
	sort.Slice(rows, func(i, j int) bool {
		if sortBy == "overlap" {
			oi := overlaps[rows[i].vodName]
			oj := overlaps[rows[j].vodName]
			if oi != oj {
				return oi > oj
			}
		}
		return rows[i].rawTime > rows[j].rawTime
	})

	printRow := func(r row) {
		dPad := strings.Repeat(" ", maxDomainIDWidth-runewidth.StringWidth(r.domainID))
		sPad := strings.Repeat(" ", maxScoreWidth-runewidth.StringWidth(r.score))
		vPad := strings.Repeat(" ", maxVodIDWidth-runewidth.StringWidth(r.vodID))
		tPad := strings.Repeat(" ", maxTypeNameWidth-runewidth.StringWidth(r.typeName))
		tmPad := strings.Repeat(" ", maxTimeWidth-runewidth.StringWidth(r.vodTime))
		titlePad := strings.Repeat(" ", maxTitleWidth-runewidth.StringWidth(r.vodName))

		fmt.Fprintf(w, "%s%s   %s%s   %s%s   %s%s   %s%s   %s%s   %s\n",
			r.domainID, dPad,
			r.score, sPad,
			r.vodID, vPad,
			r.typeName, tPad,
			r.vodTime, tmPad,
			r.vodName, titlePad,
			r.remarks)
	}

	printRow(headers)
	for _, r := range rows {
		printRow(r)
	}
}

func init() {
	searchCmd.Flags().IntVarP(&searchSeriesCount, "series", "s", 3, "Number of series to search concurrently")
	searchCmd.Flags().IntVarP(&searchPage, "page", "p", 1, "Page number for search results")
	searchCmd.Flags().StringVar(&searchSortBy, "sort", "time", "Order results by 'time' or 'overlap'")
	rootCmd.AddCommand(searchCmd)
}
