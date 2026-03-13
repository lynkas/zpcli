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
		s, err := store.Load()
		if err != nil {
			fmt.Printf("Error loading store: %v\n", err)
			return
		}

		if len(args) == 1 {
			domain := args[0]
			err = s.CreateSeries(domain)
			if err != nil {
				fmt.Printf("Error creating series: %v\n", err)
				return
			}
			fmt.Printf("Successfully created new series for domain %s\n", domain)
		} else {
			seriesID, err := strconv.Atoi(args[0])
			if err != nil {
				fmt.Printf("Invalid series ID: %v\n", err)
				return
			}
			domain := args[1]

			err = s.AddDomainToSeries(seriesID-1, domain)
			if err != nil {
				fmt.Printf("Error adding domain to series: %v\n", err)
				return
			}
			fmt.Printf("Successfully added domain %s to series %d\n", domain, seriesID)
		}
	},
}

func init() {
	rootCmd.AddCommand(addCmd)
}
