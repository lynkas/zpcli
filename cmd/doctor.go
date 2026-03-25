package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"zpcli/internal/logx"
	"zpcli/internal/service"
	"zpcli/store"

	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run environment and configuration diagnostics",
	Long: `Run local diagnostics for the current environment.

Supported forms:
  1. ` + "`zpcli doctor`" + `
     Required:
       - no positional arguments
     Optional:
       - ` + "`--json`" + `
     Behavior:
       - reports Go / gofmt availability
       - reports config path resolution and load status
       - reports whether a health report can be generated`,
	Example: ``,
	Run: func(cmd *cobra.Command, args []string) {
		runDoctor()
	},
}

func runDoctor() {
	logger := logx.Logger("cmd.doctor")
	logger.Info("doctor command start", "output_json", outputJSON)
	configPath, configErr := store.ConfigFilePath()
	data, loadErr := store.Load()

	healthService := service.NewHealthService()
	var healthReport interface{}
	if loadErr == nil {
		healthReport, _ = healthService.BuildHealthReport(data)
	}

	goPath, goFound := lookPath("go")
	goFmtPath, goFmtFound := lookPath("gofmt")
	logger.Info("doctor probes complete",
		"config_path", configPath,
		"config_error", configErr,
		"load_error", loadErr,
		"go_found", goFound,
		"go_path", goPath,
		"gofmt_found", goFmtFound,
		"gofmt_path", goFmtPath,
	)
	logger.Debug("doctor health report", "health_report", healthReport)

	if outputJSON {
		writeJSON(os.Stdout, map[string]interface{}{
			"status":      "ok",
			"os":          runtime.GOOS,
			"arch":        runtime.GOARCH,
			"go_found":    goFound,
			"go_path":     goPath,
			"gofmt_found": goFmtFound,
			"gofmt_path":  goFmtPath,
			"config_path": configPath,
			"config_error": func() string {
				if configErr != nil {
					return configErr.Error()
				}
				return ""
			}(),
			"load_error": func() string {
				if loadErr != nil {
					return loadErr.Error()
				}
				return ""
			}(),
			"health_report": healthReport,
		})
		return
	}

	fmt.Printf("OS:          %s\n", runtime.GOOS)
	fmt.Printf("Arch:        %s\n", runtime.GOARCH)
	fmt.Printf("Go found:    %t\n", goFound)
	if goFound {
		fmt.Printf("Go path:     %s\n", goPath)
	}
	fmt.Printf("Gofmt found: %t\n", goFmtFound)
	if goFmtFound {
		fmt.Printf("Gofmt path:  %s\n", goFmtPath)
	}

	if configErr != nil {
		fmt.Printf("Config path error: %v\n", configErr)
	} else {
		fmt.Printf("Config path: %s\n", configPath)
	}

	if loadErr != nil {
		fmt.Printf("Config load error: %v\n", loadErr)
		return
	}
	report, err := healthService.BuildHealthReport(data)
	if err != nil {
		fmt.Printf("Health report error: %v\n", err)
		return
	}
	fmt.Printf("Series:      %d\n", report.SeriesCount)
	fmt.Printf("Domains:     %d\n", report.DomainCount)
	fmt.Printf("Errors:      %d\n", report.InvalidCount)
	fmt.Printf("Warnings:    %d\n", report.WarningCount)
}

func lookPath(name string) (string, bool) {
	path, err := exec.LookPath(name)
	if err != nil {
		return "", false
	}
	return path, true
}

func init() {
	doctorCmd.SetOut(os.Stdout)
	rootCmd.AddCommand(doctorCmd)
}
