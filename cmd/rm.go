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
		if err := RemoveSite(args[0]); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	},
}

func RemoveSite(id string) error {
	s, err := store.Load()
	if err != nil {
		return fmt.Errorf("loading store: %v", err)
	}

	if strings.Contains(id, ".") {
		if err := s.RemoveDomain(id); err != nil {
			return err
		}
		fmt.Printf("Successfully removed domain %s\n", id)
	} else {
		seriesID, err := strconv.Atoi(id)
		if err != nil {
			return fmt.Errorf("invalid id format: %v", err)
		}
		if err := s.RemoveSeries(seriesID - 1); err != nil {
			return err
		}
		fmt.Printf("Successfully removed series %d\n", seriesID)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
