package cmd

import (
	"fmt"
	"zpcli/internal/service"
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

	siteService := service.NewSiteService()
	message, err := siteService.RemoveSite(s, id)
	if err != nil {
		return err
	}
	fmt.Println(message)
	return nil
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
