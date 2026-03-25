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

var legacyListCmd = &cobra.Command{
	Use:    listCmd.Use,
	Short:  listCmd.Short,
	Args:   listCmd.Args,
	Hidden: true,
	Run:    listCmd.Run,
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
	siteCmd.AddCommand(listCmd)
	rootCmd.AddCommand(legacyListCmd)
}
