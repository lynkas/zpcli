package cmd

import (
	"fmt"
	"os"
	"zpcli/internal/service"
	"zpcli/store"

	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate the current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		runValidate()
	},
}

func runValidate() {
	data, err := store.Load()
	if err != nil {
		if outputJSON {
			writeCommandError(os.Stdout, fmt.Sprintf("Error loading store: %v", err))
			return
		}
		fmt.Printf("Error loading store: %v\n", err)
		return
	}

	healthService := service.NewHealthService()
	issues := healthService.ValidateStore(data)
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
	rootCmd.AddCommand(validateCmd)
}
