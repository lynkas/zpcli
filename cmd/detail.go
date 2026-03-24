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
		if outputJSON {
			writeCommandError(w, fmt.Sprintf("Error loading store: %v", err))
			return
		}
		fmt.Fprintf(w, "Error loading store: %v\n", err)
		return
	}

	detailService := service.NewDetailService(nil)
	result, err := detailService.GetDetail(context.Background(), s, siteIDStr, vodIDStr)
	if err != nil {
		if outputJSON {
			writeCommandError(w, fmt.Sprintf("Error fetching detail: %v", err))
			return
		}
		fmt.Fprintf(w, "Error fetching detail: %v\n", err)
		return
	}

	if result.Err != nil {
		s.Series[result.SeriesIndex].Domains[result.DomainIndex].FailureScore++
		s.Save()
		if outputJSON {
			writeCommandError(w, fmt.Sprintf("Error fetching detail: %v", result.Err))
			return
		}
		fmt.Fprintf(w, "Error fetching detail: %v\n", result.Err)
		return
	}

	if result.Item == nil {
		if outputJSON {
			writeJSON(w, map[string]interface{}{
				"status":  "ok",
				"message": "No detail found.",
			})
			return
		}
		fmt.Fprintln(w, "No detail found.")
		return
	}

	v := result.Item

	if targetEp != "" {
		if outputJSON {
			if episodeURL, ok := service.FindEpisodeURL(v, targetEp); ok {
			writeJSON(w, map[string]interface{}{
				"status":      "ok",
					"site_id":     siteIDStr,
					"vod_id":      vodIDStr,
					"episode":     targetEp,
					"episode_url": episodeURL,
				})
				return
			}
			writeCommandError(w, fmt.Sprintf("Episode %s not found.", targetEp))
			return
		}
		writeEpisodeMatch(w, v, targetEp)
		return
	}
	if outputJSON {
		writeJSON(w, map[string]interface{}{
			"status":  "ok",
			"site_id": siteIDStr,
			"vod_id":  vodIDStr,
			"detail":  v,
		})
		return
	}
	writeDetailResult(w, v, showAll)
}

func init() {
	rootCmd.AddCommand(detailCmd)
}
