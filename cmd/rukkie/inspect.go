package main

import (
	"fmt"

	"github.com/rukkiecodes/rukkiepulse/internal/auth"
	"github.com/rukkiecodes/rukkiepulse/internal/config"
	"github.com/rukkiecodes/rukkiepulse/internal/engine"
	"github.com/rukkiecodes/rukkiepulse/internal/output"
	"github.com/spf13/cobra"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <service>",
	Short: "Deep-dive into a single service (health + all endpoints)",
	Args:  cobra.ExactArgs(1),
	RunE:  runInspect,
}

func init() {
	rootCmd.AddCommand(inspectCmd)
}

func runInspect(cmd *cobra.Command, args []string) error {
	if err := auth.RequireAuth(); err != nil {
		output.PrintError(err.Error())
		return nil
	}

	serviceName := args[0]

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

	var target *config.Service
	for i := range services {
		if services[i].Name == serviceName {
			target = &services[i]
			break
		}
	}

	if target == nil {
		output.PrintError(fmt.Sprintf("service %q not found in environment %q", serviceName, envFlag))
		return nil
	}

	results := engine.Run([]config.Service{*target})
	output.PrintInspect(results[0])

	return nil
}
