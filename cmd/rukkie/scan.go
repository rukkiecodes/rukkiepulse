package main

import (
	"github.com/rukkiecodes/rukkiepulse/internal/config"
	"github.com/rukkiecodes/rukkiepulse/internal/engine"
	"github.com/rukkiecodes/rukkiepulse/internal/output"
	"github.com/spf13/cobra"
)

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
	rootCmd.AddCommand(scanCmd)
	rootCmd.AddCommand(statusCmd)
}

func runScan(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load("rukkie.yaml")
	if err != nil {
		output.PrintError(err.Error())
		return nil
	}

	services, err := cfg.GetServices(envFlag)
	if err != nil {
		output.PrintError(err.Error())
		return nil
	}

	output.PrintScanHeader(cfg.Project, envFlag)

	results := engine.Run(services)
	output.PrintResults(results)

	return nil
}
