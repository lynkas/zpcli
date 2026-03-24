package cmd

import (
	"fmt"
	"io"
	"sort"
	"strings"
	"zpcli/internal/domain"
	"zpcli/internal/service"
	"zpcli/store"

	"github.com/mattn/go-runewidth"
)

func writeSearchResults(w io.Writer, data *store.StoreData, results []domain.SearchResult, sortBy string) {
	type row struct {
		domainID string
		score    string
		vodID    string
		typeName string
		vodName  string
		vodTime  string
		remarks  string
		rawTime  string
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
		score := fmt.Sprintf("%d", data.Series[res.SeriesIndex].Domains[res.DomainIndex].FailureScore)
		if res.Err != nil || len(res.Items) == 0 {
			continue
		}

		if width := runewidth.StringWidth(domainID); width > maxDomainIDWidth {
			maxDomainIDWidth = width
		}
		if width := runewidth.StringWidth(score); width > maxScoreWidth {
			maxScoreWidth = width
		}

		for _, item := range res.Items {
			vID := fmt.Sprintf("%d", item.VodID)
			relTime := formatRelativeTime(item.VodTime)
			if width := runewidth.StringWidth(vID); width > maxVodIDWidth {
				maxVodIDWidth = width
			}
			if width := runewidth.StringWidth(item.TypeName); width > maxTypeNameWidth {
				maxTypeNameWidth = width
			}
			if width := runewidth.StringWidth(relTime); width > maxTimeWidth {
				maxTimeWidth = width
			}
			if width := runewidth.StringWidth(item.VodName); width > maxTitleWidth {
				maxTitleWidth = width
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

func writeDetailResult(w io.Writer, item *domain.DetailItem, showAll bool) {
	fmt.Fprintf(w, "Name:    %s\n", item.VodName)
	if item.VodSub != "" {
		fmt.Fprintf(w, "Sub:     %s\n", item.VodSub)
	}
	if showAll {
		fmt.Fprintf(w, "Type:    %s\n", item.TypeName)
		fmt.Fprintf(w, "Tag:     %s\n", item.VodTag)
		fmt.Fprintf(w, "Class:   %s\n", item.VodClass)
		fmt.Fprintf(w, "Actor:   %s\n", item.VodActor)
		fmt.Fprintf(w, "Director:%s\n", item.VodDirector)
		fmt.Fprintf(w, "Area:    %s\n", item.VodArea)
		fmt.Fprintf(w, "Lang:    %s\n", item.VodLang)
		fmt.Fprintf(w, "Year:    %s\n", item.VodYear)
		fmt.Fprintf(w, "PubDate: %s\n", item.VodPubDate)
		fmt.Fprintf(w, "Total:   %d\n", item.VodTotal)
		fmt.Fprintf(w, "Hits:    %d\n", item.VodHits)
		fmt.Fprintf(w, "Score:   %s (Douban: %s)\n", item.VodScore, item.VodDoubanScore)
		fmt.Fprintf(w, "Time:    %s\n", item.VodTime)
		fmt.Fprintf(w, "Remarks: %s\n", item.VodRemarks)
	}
	fmt.Fprintf(w, "\nContent:\n%s\n", item.VodContent)
	fmt.Fprintf(w, "\nPlay URLs:\n")

	for _, player := range item.Players {
		fmt.Fprintf(w, "\n[%s]\n", player.Name)
		for _, episode := range player.Episodes {
			fmt.Fprintf(w, "  %s: %s\n", episode.Name, episode.URL)
		}
	}
}

func writeEpisodeMatch(w io.Writer, item *domain.DetailItem, targetEp string) {
	if episodeURL, ok := service.FindEpisodeURL(item, targetEp); ok {
		fmt.Fprintln(w, episodeURL)
		return
	}
	fmt.Fprintf(w, "Episode %s not found.\n", targetEp)
}
