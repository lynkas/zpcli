package cmd

import (
	"context"
	"fmt"
	"io"
	"os"
	"zpcli/internal/service"
	"zpcli/store"

	"github.com/spf13/cobra"
)

var detailCmd = &cobra.Command{
	Use:     "detail [siteId] [vodId]",
	Aliases: []string{"d"},
	Short:   "Show details of a specific video (alias: d)",
	Args:    cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		ShowDetail(os.Stdout, detailShowAll, args...)
	},
}

func ShowDetail(w io.Writer, showAll bool, args ...string) {
	siteIDStr := args[0]
	vodIDStr := args[1]
	targetEp := ""
	if len(args) == 3 {
		targetEp = args[2]
	}

	s, err := store.Load()
	if err != nil {
		fmt.Fprintf(w, "Error loading store: %v\n", err)
		return
	}

	detailService := service.NewDetailService(nil)
	result, err := detailService.GetDetail(context.Background(), s, siteIDStr, vodIDStr)
	if err != nil {
		fmt.Fprintf(w, "Error fetching detail: %v\n", err)
		return
	}

	if result.Err != nil {
		fmt.Fprintf(w, "Error fetching detail: %v\n", result.Err)
		s.Series[result.SeriesIndex].Domains[result.DomainIndex].FailureScore++
		s.Save()
		return
	}

	if result.Item == nil {
		fmt.Fprintln(w, "No detail found.")
		return
	}

	v := result.Item

	if targetEp != "" {
		writeEpisodeMatch(w, v, targetEp)
		return
	}
	writeDetailResult(w, v, showAll)
}

func init() {
	rootCmd.AddCommand(detailCmd)
}
