package cmd

import (
	"fmt"
	"strconv"
	"strings"
	"zpcli/store"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "rm [id]",
	Short: "Remove a series or a single domain",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		s, err := store.Load()
		if err != nil {
			fmt.Printf("Error loading store: %v\n", err)
			return
		}

		id := args[0]
		if strings.Contains(id, ".") {
			if err := s.RemoveDomain(id); err != nil {
				fmt.Printf("Error removing domain: %v\n", err)
			} else {
				fmt.Printf("Successfully removed domain %s\n", id)
			}
		} else {
			seriesID, err := strconv.Atoi(id)
			if err != nil {
				fmt.Printf("Invalid id format: %v\n", err)
				return
			}
			if err := s.RemoveSeries(seriesID); err != nil {
				fmt.Printf("Error removing series: %v\n", err)
			} else {
				fmt.Printf("Successfully removed series %d\n", seriesID)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
