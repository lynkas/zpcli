package cmd

import (
	"fmt"
	"strconv"
	"zpcli/store"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [domain] OR add [seriesId] [domain]",
	Short: "Create a series, add a domain to a series, or add an endpoint",
	Args:  cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		if err := AddSite(args...); err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}
	},
}

func AddSite(args ...string) error {
	s, err := store.Load()
	if err != nil {
		return fmt.Errorf("loading store: %v", err)
	}

	if len(args) == 1 {
		domain := args[0]
		err = s.CreateSeries(domain)
		if err != nil {
			return err
		}
		fmt.Printf("Successfully created new series for domain %s\n", domain)
	} else {
		seriesID, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid series ID: %v", err)
		}
		domain := args[1]

		err = s.AddDomainToSeries(seriesID-1, domain)
		if err != nil {
			return err
		}
		fmt.Printf("Successfully added domain %s to series %d\n", domain, seriesID)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(addCmd)
}
