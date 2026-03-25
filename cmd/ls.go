package cmd

import (
	"fmt"
	"io"
	"os"
	"zpcli/internal/logx"
	"zpcli/internal/service"
	"zpcli/store"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "ls",
	Short: "List all domains and endpoints",
	Long: `List all configured series, domains, and endpoint failure counts.

Supported forms:
  1. ` + "`zpcli site ls`" + `
     Required:
       - no positional arguments
     Optional:
       - ` + "`--json`" + `
     Behavior:
       - prints every series, every domain ID, and failure counts

Output:
  - text mode shows ` + "`Series N`" + ` sections and domain IDs like ` + "`1.2`" + `
  - JSON mode returns a structured site list`,
	Example: ``,
	Run: func(cmd *cobra.Command, args []string) {
		ShowList(os.Stdout)
	},
}

func ShowList(w io.Writer) {
	logger := logx.Logger("cmd.site.ls")
	logger.Info("list sites command start", "output_json", outputJSON)
	s, err := store.Load()
	if err != nil {
		logger.Error("load store failed", "error", err)
		if outputJSON {
			writeCommandError(w, fmt.Sprintf("Error loading store: %v", err))
			return
		}
		fmt.Fprintf(w, "Error loading store: %v\n", err)
		return
	}

	if len(s.Series) == 0 {
		logger.Info("list sites no configured series")
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
	logger.Info("list sites command complete", "series_count", len(seriesList))
	logger.Debug("list sites output", "sites", seriesList)
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
	siteCmd.AddCommand(listCmd)
}
