package cmd

import (
	"fmt"
	"os"
	"zpcli/internal/service"
	"zpcli/store"

	"github.com/spf13/cobra"
)

var healthCmd = &cobra.Command{
	Use:   "health",
	Short: "Show health information for the current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		showHealth()
	},
}

func showHealth() {
	data, err := store.Load()
	if err != nil {
		fmt.Printf("Error loading store: %v\n", err)
		return
	}

	healthService := service.NewHealthService()
	report, err := healthService.BuildHealthReport(data)
	if err != nil {
		fmt.Printf("Error building health report: %v\n", err)
		return
	}

	fmt.Printf("Config:   %s\n", report.ConfigPath)
	fmt.Printf("Version:  %d\n", report.Version)
	fmt.Printf("Series:   %d\n", report.SeriesCount)
	fmt.Printf("Domains:  %d\n", report.DomainCount)
	fmt.Printf("Errors:   %d\n", report.InvalidCount)
	fmt.Printf("Warnings: %d\n", report.WarningCount)

	if len(report.Issues) == 0 {
		fmt.Println("\nStatus: healthy")
		return
	}

	fmt.Println("\nIssues:")
	for _, issue := range report.Issues {
		label := issue.Scope
		if issue.SiteID != "" {
			label = fmt.Sprintf("%s %s", issue.Scope, issue.SiteID)
		}
		fmt.Printf("  [%s] %s: %s\n", issue.Level, label, issue.Message)
	}
}

func init() {
	healthCmd.SetOut(os.Stdout)
	rootCmd.AddCommand(healthCmd)
}
