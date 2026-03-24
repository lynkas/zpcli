package cmd

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
	"zpcli/internal/service"
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

	searchService := service.NewSearchService(nil)
	results, err := searchService.Search(context.Background(), s, keyword, seriesCount, page)
	if err != nil {
		fmt.Fprintf(w, "Error searching sites: %v\n", err)
		return
	}

	if len(results) == 0 {
		fmt.Fprintln(w, "No valid domains to search.")
		return
	}

	hasFailures := false
	for _, res := range results {
		if res.Err != nil {
			s.Series[res.SeriesIndex].Domains[res.DomainIndex].FailureScore++
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

	for _, res := range results {
		domainID := fmt.Sprintf("%d.%d", res.SeriesIndex+1, res.DomainIndex+1)
		score := fmt.Sprintf("%d", s.Series[res.SeriesIndex].Domains[res.DomainIndex].FailureScore)
		if res.Err != nil || len(res.Items) == 0 {
			continue
		}

		if runewidth.StringWidth(domainID) > maxDomainIDWidth {
			maxDomainIDWidth = runewidth.StringWidth(domainID)
		}
		if runewidth.StringWidth(score) > maxScoreWidth {
			maxScoreWidth = runewidth.StringWidth(score)
		}

		for _, item := range res.Items {
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
