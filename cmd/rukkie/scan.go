package main

import (
	"github.com/rukkiecodes/rukkiepulse/internal/auth"
	"github.com/rukkiecodes/rukkiepulse/internal/cloud"
	"github.com/rukkiecodes/rukkiepulse/internal/config"
	"github.com/rukkiecodes/rukkiepulse/internal/engine"
	"github.com/rukkiecodes/rukkiepulse/internal/output"
	"github.com/spf13/cobra"
)

var errorsOnlyFlag bool

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan all services and check their health",
	RunE:  runScan,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current health status of all services",
	RunE:  runScan,
}

func init() {
	scanCmd.Flags().BoolVar(&errorsOnlyFlag, "errors-only", false, "show only services with issues")
	statusCmd.Flags().BoolVar(&errorsOnlyFlag, "errors-only", false, "show only services with issues")
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(statusCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	if err := auth.RequireAuth(); err != nil {
		output.PrintError(err.Error())
		return nil
	}

	cfg, err := config.Load("rukkie.yaml")
	if err != nil {
		// No local rukkie.yaml — fall back to cloud services
		return runCloudScan()
	}

	services, err := cfg.GetServices(envFlag)
	if err != nil {
		output.PrintError(err.Error())
		return nil
	}

	output.PrintScanHeader(cfg.Project, envFlag)

	results := engine.Run(services)

	if errorsOnlyFlag {
		filtered := results[:0]
		for _, r := range results {
			if r.Health.Status != "ok" {
				filtered = append(filtered, r)
			} else {
				pass, total := r.PassingEndpoints()
				if total > 0 && pass < total {
					filtered = append(filtered, r)
				}
			}
		}
		results = filtered
		if len(results) == 0 {
			output.PrintAllClear()
			return nil
		}
	}

	output.PrintResults(results)
	return nil
}

func runCloudScan() error {
	services, err := cloud.FetchServices()
	if err != nil {
		output.PrintError("Could not reach RukkiePulse cloud: " + err.Error())
		return nil
	}
	output.PrintCloudServices(services)
	return nil
}
