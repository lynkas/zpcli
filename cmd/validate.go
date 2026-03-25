package cmd

import (
	"fmt"
	"os"
	"zpcli/internal/logx"
	"zpcli/internal/service"
	"zpcli/store"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the current configuration",
	Long: `Validate the current site configuration and report structural or data issues.

Supported forms:
  1. ` + "`zpcli site validate`" + `
     Required:
       - no positional arguments
     Optional:
       - ` + "`--json`" + `
     Behavior:
       - checks the stored configuration and prints any issues found`,
	Example: ``,
	Run: func(cmd *cobra.Command, args []string) {
		runValidate()
	},
}

func runValidate() {
	logger := logx.Logger("cmd.site.validate")
	logger.Info("validate command start", "output_json", outputJSON)
	data, err := store.Load()
	if err != nil {
		logger.Error("load store failed", "error", err)
		if outputJSON {
			writeCommandError(os.Stdout, fmt.Sprintf("Error loading store: %v", err))
			return
		}
		fmt.Printf("Error loading store: %v\n", err)
		return
	}

	healthService := service.NewHealthService()
	issues := healthService.ValidateStore(data)
	logger.Info("validate command complete", "issue_count", len(issues))
	logger.Debug("validate issues", "issues", issues)
	if outputJSON {
		writeJSON(os.Stdout, map[string]interface{}{
			"status": "ok",
			"issues": issues,
		})
		return
	}
	if len(issues) == 0 {
		fmt.Println("Configuration is valid.")
		return
	}

	fmt.Printf("Found %d issue(s):\n", len(issues))
	for _, issue := range issues {
		label := issue.Scope
		if issue.SiteID != "" {
			label = fmt.Sprintf("%s %s", issue.Scope, issue.SiteID)
		}
		fmt.Printf("  [%s] %s: %s\n", issue.Level, label, issue.Message)
	}
}

func init() {
	validateCmd.SetOut(os.Stdout)
	siteCmd.AddCommand(validateCmd)
}
