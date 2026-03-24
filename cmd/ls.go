package cmd

import (
	"fmt"
	"io"
	"os"
	"zpcli/internal/service"
	"zpcli/store"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all domains and endpoints",
	Run: func(cmd *cobra.Command, args []string) {
		ShowList(os.Stdout)
	},
}

func ShowList(w io.Writer) {
	s, err := store.Load()
	if err != nil {
		fmt.Fprintf(w, "Error loading store: %v\n", err)
		return
	}

	if len(s.Series) == 0 {
		fmt.Fprintln(w, "No series configured.")
		return
	}

	siteService := service.NewSiteService()
	for _, series := range siteService.ListSites(s) {
		fmt.Fprintf(w, "Series %d:\n", series.SeriesID)
		for _, dom := range series.Domains {
			domainID := fmt.Sprintf("%d.%d", dom.SeriesID, dom.DomainID)
			fmt.Fprintf(w, "  [%s] URL: %s [Failures: %d]\n", domainID, dom.URL, dom.FailureScore)
		}
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
}
