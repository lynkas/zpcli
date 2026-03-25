package cmd

import (
	"fmt"
	"os"
	"strings"
	"zpcli/internal/logx"

	"github.com/spf13/cobra"
)

var (
	detailShowAll bool
	outputJSON    bool
	logFormat     string
	verboseCount  int
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
	Long: `Manage your favorite video sites and search content across them quickly.

TOP-LEVEL COMMAND GUIDE
1. SITE MANAGEMENT
   Command group:
     ` + "`zpcli site`" + `

   What it does:
     Manage the local site configuration, including creating series,
     adding domains, listing configured sites, removing entries,
     validating config data, and checking config health.

   Supported actions:
     - ` + "`zpcli site add <domain>`" + `
       Required parameters:
         - ` + "`domain`" + `
       Optional parameters:
         - none
       Behavior:
         - creates a new series and adds the domain to it

     - ` + "`zpcli site add <seriesId> <domain>`" + `
       Required parameters:
         - ` + "`seriesId`" + `
         - ` + "`domain`" + `
       Optional parameters:
         - none
       Behavior:
         - adds a domain to an existing series

     - ` + "`zpcli site ls`" + `
       Required parameters:
         - none
       Optional parameters:
         - ` + "`--json`" + `
       Behavior:
         - lists all series, domain IDs, URLs, and failure counts

     - ` + "`zpcli site rm <seriesId>`" + `
       Required parameters:
         - ` + "`seriesId`" + `
       Optional parameters:
         - none
       Behavior:
         - removes the entire series

     - ` + "`zpcli site rm <seriesId.domainId>`" + `
       Required parameters:
         - ` + "`seriesId.domainId`" + `, such as ` + "`1.2`" + `
       Optional parameters:
         - none
       Behavior:
         - removes one domain from a series

     - ` + "`zpcli site validate`" + `
       Required parameters:
         - none
       Optional parameters:
         - ` + "`--json`" + `
       Behavior:
         - validates the stored config and reports problems

     - ` + "`zpcli site health`" + `
       Required parameters:
         - none
       Optional parameters:
         - ` + "`--json`" + `
       Behavior:
         - shows totals, warnings, errors, and config path information

   Parameter notes:
     - ` + "`domain`" + ` may be a bare host like ` + "`example.com`" + ` or a full API endpoint URL
     - ` + "`seriesId`" + ` must refer to an existing series when used with ` + "`site add <seriesId> <domain>`" + `

2. SEARCH
   Command:
     ` + "`zpcli search <keyword>`" + `

   What it does:
     Search videos across configured sites.

   Required parameters:
     - ` + "`keyword`" + `: one or more words to search for

   Optional parameters:
     - ` + "`--series <n>`" + `: number of series to search concurrently
     - ` + "`--page <n>`" + `: result page number
     - ` + "`--sort <time|overlap>`" + `: result ordering
     - ` + "`--json`" + `: return JSON output

   Examples:
     - ` + "`zpcli search movie`" + `
     - ` + "`zpcli search --series 5 --page 2 movie`" + `
     - ` + "`zpcli search --sort overlap movie`" + `

3. DETAIL
   Commands:
     - ` + "`zpcli detail <siteId> <vodId>`" + `
     - ` + "`zpcli <siteId> <vodId> [episode]`" + `  (shortcut form)

   What it does:
     Fetch metadata, episode information, and optionally a specific episode URL.

   Required parameters:
     - ` + "`siteId`" + `: configured site ID such as ` + "`1.1`" + `
     - ` + "`vodId`" + `: video ID on that site

   Optional parameters:
     - ` + "`episode`" + `: only in shortcut form; fetch one matching episode link
     - ` + "`--all`" + `: show more detail
     - ` + "`--json`" + `: return JSON output

   Examples:
     - ` + "`zpcli detail 1.1 12345`" + `
     - ` + "`zpcli detail 1.1 12345 --all`" + `
     - ` + "`zpcli 1.1 12345`" + `
     - ` + "`zpcli 1.1 12345 3`" + `

4. DOCTOR
   Command:
     ` + "`zpcli doctor`" + `

   What it does:
     Run local environment and configuration diagnostics.

   Required parameters:
     - none

   Optional parameters:
     - ` + "`--json`" + `

   Behavior:
     - checks whether ` + "`go`" + ` and ` + "`gofmt`" + ` are available
     - checks config path resolution and config loading
     - checks whether a health report can be generated

5. MCP
   Command:
     ` + "`zpcli mcp`" + `

   What it does:
     Start the MCP server for AI / MCP clients.

   Required parameters:
     - none

   Optional parameters:
     - ` + "`--port <port>`" + `: run as SSE / HTTP instead of stdio

   Supported forms:
     - ` + "`zpcli mcp`" + `: start in stdio mode
     - ` + "`zpcli mcp --port 8080`" + `: start in SSE / HTTP mode

GLOBAL FLAGS
  - ` + "`--json`" + `: optional, machine-readable JSON output where supported
  - ` + "`--all`" + `: optional, show more detail for detail output
  - ` + "`-v`" + `: optional, enable logs; repeat as ` + "`-vv`" + ` for more detail`,
	Example: ``,
	Args:    cobra.ArbitraryArgs,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		logx.Init(verboseCount, logFormat)
		if logx.Enabled() {
			logx.Logger("cmd.root").Info("command start",
				"command", cmd.CommandPath(),
				"args", args,
				"output_json", outputJSON,
				"detail_show_all", detailShowAll,
				"verbosity", verboseCount,
				"log_format", logFormat,
			)
		}
	},
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
	rootCmd.PersistentFlags().CountVarP(&verboseCount, "verbose", "v", "Enable logs; use -vv for more detail")
	rootCmd.PersistentFlags().StringVar(&logFormat, "log-format", "text", "Log format: text or json")
}
