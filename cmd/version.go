package cmd

import (
	"fmt"
	"os"
	"zpcli/internal/buildinfo"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version, commit, and build date",
	Long: `Show build metadata for the current zpcli binary.

Fields:
  - version: semantic version from the VERSION file during release builds
  - commit: git commit hash injected at build time
  - build date: UTC timestamp injected at build time`,
	Run: func(cmd *cobra.Command, args []string) {
		if outputJSON {
			writeJSON(os.Stdout, map[string]string{
				"version":    buildinfo.Version,
				"commit":     buildinfo.Commit,
				"build_date": buildinfo.BuildDate,
			})
			return
		}

		fmt.Printf("zpcli v%s\n", buildinfo.Version)
		fmt.Printf("commit: %s\n", buildinfo.Commit)
		fmt.Printf("built:  %s\n", buildinfo.BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
