package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var (
	detailShowAll bool
	outputJSON    bool
)

func Execute() {
	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.SetHelpCommand(&cobra.Command{Hidden: true})

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "zpcli [command] or zpcli [siteId] [vodId] [episode?]",
	Short: "ZPCLI is a fast video CMS query tool",
	Long:  `Manage your favorite video sites and search content across them quickly.`,
	Args:  cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if (len(args) == 2 || len(args) == 3) && strings.Contains(args[0], ".") {
			ShowDetail(os.Stdout, detailShowAll, args...)
			return
		}
		cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&detailShowAll, "all", "a", false, "Show all provided information")
	rootCmd.PersistentFlags().BoolVar(&outputJSON, "json", false, "Output machine-readable JSON")
}
