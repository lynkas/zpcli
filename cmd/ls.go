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
		if outputJSON {
			writeCommandError(w, fmt.Sprintf("Error loading store: %v", err))
			return
		}
		fmt.Fprintf(w, "Error loading store: %v\n", err)
		return
	}

	if len(s.Series) == 0 {
		if outputJSON {
			writeJSON(w, map[string]interface{}{
				"status": "ok",
				"sites":  []interface{}{},
			})
			return
		}
		fmt.Fprintln(w, "No series configured.")
		return
	}

	siteService := service.NewSiteService()
	seriesList := siteService.ListSites(s)
	if outputJSON {
		writeJSON(w, map[string]interface{}{
			"status": "ok",
			"sites":  seriesList,
		})
		return
	}
	for _, series := range seriesList {
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
