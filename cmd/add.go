package cmd

import (
	"fmt"
	"os"
	"zpcli/internal/service"
	"zpcli/store"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add [domain] OR add [seriesId] [domain]",
	Short: "Create a series, add a domain to a series, or add an endpoint",
	Long: `Add a new site configuration.

Supported forms:
  1. ` + "`zpcli site add <domain>`" + `
     Required:
       - ` + "`domain`" + `
     Optional:
       - none
     Behavior:
       - creates a new series and puts the domain into that new series

  2. ` + "`zpcli site add <seriesId> <domain>`" + `
     Required:
       - ` + "`seriesId`" + `: target series number, such as ` + "`1`" + `
       - ` + "`domain`" + `
     Optional:
       - none
     Behavior:
       - appends the domain to an existing series

Parameter details:
  - ` + "`domain`" + ` may be a bare host like ` + "`example.com`" + ` or a full endpoint URL
  - ` + "`seriesId`" + ` must refer to an existing series`,
	Example: ``,
	Args:    cobra.MaximumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 {
			cmd.Help()
			return
		}
		if err := AddSite(args...); err != nil {
			if outputJSON {
				writeCommandError(os.Stdout, fmt.Sprintf("Error: %v", err))
				return
			}
			fmt.Printf("Error: %v\n", err)
			return
		}
	},
}

var legacyAddCmd = &cobra.Command{
	Use:    addCmd.Use,
	Short:  addCmd.Short,
	Args:   addCmd.Args,
	Hidden: true,
	Run:    addCmd.Run,
}

func AddSite(args ...string) error {
	s, err := store.Load()
	if err != nil {
		return fmt.Errorf("loading store: %v", err)
	}

	siteService := service.NewSiteService()
	message, err := siteService.AddSite(s, args...)
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
	siteCmd.AddCommand(addCmd)
	rootCmd.AddCommand(legacyAddCmd)
}
