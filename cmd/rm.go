package cmd

import (
	"fmt"
	"os"
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
			if outputJSON {
				writeCommandError(os.Stdout, fmt.Sprintf("Error: %v", err))
				return
			}
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
	if outputJSON {
		writeJSON(os.Stdout, commandStatus{Status: "ok", Message: message})
		return nil
	}
	fmt.Println(message)
	return nil
}

func init() {
	rootCmd.AddCommand(removeCmd)
}
