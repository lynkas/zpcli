package cmd

import (
	"fmt"
	"io"
	"os"
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

	for i, series := range s.Series {
		seriesId := i + 1
		fmt.Fprintf(w, "Series %d:\n", seriesId)

		for j, dom := range series.Domains {
			domainId := fmt.Sprintf("%d.%d", seriesId, j+1)
			fmt.Fprintf(w, "  [%s] URL: %s [Failures: %d]\n", domainId, dom.URL, dom.FailureScore)
		}
	}
}

func init() {
	rootCmd.AddCommand(listCmd)
}
