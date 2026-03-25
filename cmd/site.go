package cmd

import "github.com/spf13/cobra"

var siteCmd = &cobra.Command{
	Use:   "site",
	Short: "Manage configured sites",
	Long: `Manage site configuration stored in the local config file.

Functions:
  - ` + "`zpcli site add <domain>`" + `: create a new series with one site
  - ` + "`zpcli site add <seriesId> <domain>`" + `: add one site to an existing series
  - ` + "`zpcli site ls`" + `: list all series and all configured domains
  - ` + "`zpcli site rm <seriesId>`" + `: remove an entire series
  - ` + "`zpcli site rm <seriesId.domainId>`" + `: remove one domain from a series
  - ` + "`zpcli site validate`" + `: validate the stored configuration
  - ` + "`zpcli site health`" + `: show counts, warnings, and errors for the config

Parameters:
  - ` + "`seriesId`" + `: required for commands that target one series, such as ` + "`site add <seriesId> <domain>`" + `
  - ` + "`domainId`" + `: required only when removing one domain from a series
  - ` + "`domain`" + `: required when adding a site; may be a host or full API URL

Optional flags:
  - ` + "`--json`" + `: return JSON instead of text`,
	Example: ``,
}

func init() {
	rootCmd.AddCommand(siteCmd)
}
