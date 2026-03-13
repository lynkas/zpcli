package cmd

import (
	"encoding/json"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
	"zpcli/store"

	"github.com/spf13/cobra"
)

type DetailResponse struct {
	List []struct {
		VodName        string `json:"vod_name"`
		VodSub         string `json:"vod_sub"`
		TypeName       string `json:"type_name"`
		VodTag         string `json:"vod_tag"`
		VodClass       string `json:"vod_class"`
		VodActor       string `json:"vod_actor"`
		VodDirector    string `json:"vod_director"`
		VodArea        string `json:"vod_area"`
		VodLang        string `json:"vod_lang"`
		VodYear        string `json:"vod_year"`
		VodHits        int    `json:"vod_hits"`
		VodScore       string `json:"vod_score"`
		VodDoubanScore string `json:"vod_douban_score"`
		VodTime        string `json:"vod_time"`
		VodPubDate     string `json:"vod_pubdate"`
		VodTotal       int    `json:"vod_total"`
		VodContent     string `json:"vod_content"`
		VodPlayURL     string `json:"vod_play_url"`
		VodPlayFrom    string `json:"vod_play_from"`
		VodRemarks     string `json:"vod_remarks"`
	} `json:"list"`
}

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

	sIdx, dIdx, err := parseDomainID(siteIDStr)
	if err != nil {
		fmt.Fprintf(w, "Error parsing siteId: %v\n", err)
		return
	}

	s, err := store.Load()
	if err != nil {
		fmt.Fprintf(w, "Error loading store: %v\n", err)
		return
	}

	if sIdx < 0 || sIdx >= len(s.Series) || dIdx < 0 || dIdx >= len(s.Series[sIdx].Domains) {
		fmt.Fprintf(w, "Invalid siteId: %s\n", siteIDStr)
		return
	}

	domain := s.Series[sIdx].Domains[dIdx]
	baseEndpoint := buildEndpointURL(domain.URL)
	reqURL := fmt.Sprintf("%s?ac=detail&ids=%s", baseEndpoint, vodIDStr)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(reqURL)
	if err != nil {
		fmt.Fprintf(w, "Error fetching detail: %v\n", err)
		domain.FailureScore++
		s.Save()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(w, "Error fetching detail: HTTP %d\n", resp.StatusCode)
		domain.FailureScore++
		s.Save()
		return
	}

	var detailResp DetailResponse
	if err := json.NewDecoder(resp.Body).Decode(&detailResp); err != nil {
		fmt.Fprintf(w, "Error decoding response: %v\n", err)
		domain.FailureScore++
		s.Save()
		return
	}

	if len(detailResp.List) == 0 {
		fmt.Fprintln(w, "No detail found.")
		return
	}

	v := detailResp.List[0]

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
		fmt.Fprintf(w, "\nContent:\n%s\n", stripHTML(v.VodContent))

		fmt.Fprintf(w, "\nPlay URLs:\n")
	}

	players := strings.Split(v.VodPlayFrom, "$$$")
	groups := strings.Split(v.VodPlayURL, "$$$")

	for i, group := range groups {
		playerName := fmt.Sprintf("Player %d", i+1)
		if i < len(players) {
			playerName = players[i]
		}
		if targetEp == "" {
			fmt.Fprintf(w, "\n[%s]\n", playerName)
		}

		episodes := strings.Split(group, "#")
		for j, ep := range episodes {
			parts := strings.Split(ep, "$")
			name := ep
			u := ep
			if len(parts) == 2 {
				name = parts[0]
				u = parts[1]
			}

			// If target episode specified, check if it matches name or index (1-based)
			if targetEp != "" {
				idxStr := strconv.Itoa(j + 1)
				if targetEp == idxStr || strings.Contains(name, targetEp) {
					fmt.Fprintln(w, u)
					return
				}
				continue
			}

			// Detail view
			fmt.Fprintf(w, "  %s: %s\n", name, u)
		}
	}

	if targetEp != "" {
		fmt.Fprintf(w, "Episode %s not found.\n", targetEp)
	}
}

func stripHTML(s string) string {
	s = html.UnescapeString(s)
	r := strings.NewReplacer(
		"<p>", "",
		"</p>", "\n",
		"<br>", "\n",
		"<br/>", "\n",
		"<br />", "\n",
	)
	s = r.Replace(s)
	// Simple regex-less tag removal
	var result strings.Builder
	inTag := false
	for _, r := range s {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			continue
		}
		if !inTag {
			result.WriteRune(r)
		}
	}
	return strings.TrimSpace(result.String())
}

func parseDomainID(id string) (int, int, error) {
	parts := strings.Split(id, ".")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid siteId format (expected x.y)")
	}
	sIdx, err := strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, err
	}
	dIdx, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return sIdx - 1, dIdx - 1, nil
}

func init() {
	rootCmd.AddCommand(detailCmd)
}
