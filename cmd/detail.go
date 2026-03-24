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

	if targetEp == "" {
		fmt.Fprintf(w, "Name:    %s\n", v.VodName)
		if v.VodSub != "" {
			fmt.Fprintf(w, "Sub:     %s\n", v.VodSub)
		}
		if showAll {
			fmt.Fprintf(w, "Type:    %s\n", v.TypeName)
			fmt.Fprintf(w, "Tag:     %s\n", v.VodTag)
			fmt.Fprintf(w, "Class:   %s\n", v.VodClass)
			fmt.Fprintf(w, "Actor:   %s\n", v.VodActor)
			fmt.Fprintf(w, "Director:%s\n", v.VodDirector)
			fmt.Fprintf(w, "Area:    %s\n", v.VodArea)
			fmt.Fprintf(w, "Lang:    %s\n", v.VodLang)
			fmt.Fprintf(w, "Year:    %s\n", v.VodYear)
			fmt.Fprintf(w, "PubDate: %s\n", v.VodPubDate)
			fmt.Fprintf(w, "Total:   %d\n", v.VodTotal)
			fmt.Fprintf(w, "Hits:    %d\n", v.VodHits)
			fmt.Fprintf(w, "Score:   %s (Douban: %s)\n", v.VodScore, v.VodDoubanScore)
			fmt.Fprintf(w, "Time:    %s\n", v.VodTime)
			fmt.Fprintf(w, "Remarks: %s\n", v.VodRemarks)
		}
		fmt.Fprintf(w, "\nContent:\n%s\n", v.VodContent)

		fmt.Fprintf(w, "\nPlay URLs:\n")
	}

	if targetEp != "" {
		if episodeURL, ok := service.FindEpisodeURL(v, targetEp); ok {
			fmt.Fprintln(w, episodeURL)
			return
		}
		fmt.Fprintf(w, "Episode %s not found.\n", targetEp)
		return
	}

	for _, player := range v.Players {
		fmt.Fprintf(w, "\n[%s]\n", player.Name)
		for _, episode := range player.Episodes {
			fmt.Fprintf(w, "  %s: %s\n", episode.Name, episode.URL)
		}
	}
}

func init() {
	rootCmd.AddCommand(detailCmd)
}
