package cmd

import (
	"fmt"
	"os"
	"zpcli/internal/logx"
	"zpcli/internal/service"
	"zpcli/store"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "rm [id]",
	Short: "Remove a series or a single domain",
	Long: `Remove a configured site entry.

Supported forms:
  1. ` + "`zpcli site rm <seriesId>`" + `
     Required:
       - ` + "`seriesId`" + `
     Optional:
       - none
     Behavior:
       - removes the entire series and all domains inside it

  2. ` + "`zpcli site rm <seriesId.domainId>`" + `
     Required:
       - a combined ID such as ` + "`1.2`" + `
     Optional:
       - none
     Behavior:
       - removes one domain from a series
       - if that was the last domain, the series is removed as well`,
	Example: ``,
	Args:    cobra.ExactArgs(1),
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
	logger := logx.Logger("cmd.site.rm")
	logger.Info("remove site command start", "id", id, "output_json", outputJSON)
	s, err := store.Load()
	if err != nil {
		logger.Error("load store failed", "error", err)
		return fmt.Errorf("loading store: %v", err)
	}

	siteService := service.NewSiteService()
	message, err := siteService.RemoveSite(s, id)
	if err != nil {
		logger.Error("site service remove failed", "id", id, "error", err)
		return err
	}
	logger.Info("remove site command complete", "message", message)
	if outputJSON {
		writeJSON(os.Stdout, commandStatus{Status: "ok", Message: message})
		return nil
	}
	fmt.Println(message)
	return nil
}

func init() {
	siteCmd.AddCommand(removeCmd)
}
