package cmd

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
	"zpcli/store"

	"os"

	"zpcli/internal/service"

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
	writeSearchResults(w, s, results, sortBy)
}

func init() {
	searchCmd.Flags().IntVarP(&searchSeriesCount, "series", "s", 3, "Number of series to search concurrently")
	searchCmd.Flags().IntVarP(&searchPage, "page", "p", 1, "Page number for search results")
	searchCmd.Flags().StringVar(&searchSortBy, "sort", "time", "Order results by 'time' or 'overlap'")
	rootCmd.AddCommand(searchCmd)
}
