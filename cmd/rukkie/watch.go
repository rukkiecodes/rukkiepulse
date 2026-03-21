package main

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/rukkiecodes/rukkiepulse/internal/auth"
	"github.com/rukkiecodes/rukkiepulse/internal/config"
	"github.com/rukkiecodes/rukkiepulse/internal/output"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Live-updating dashboard of all services",
	RunE:  runWatch,
}

var watchIntervalFlag time.Duration

func init() {
	watchCmd.Flags().DurationVar(&watchIntervalFlag, "interval", 10*time.Second, "refresh interval (e.g. 5s, 30s, 1m)")
	rootCmd.AddCommand(watchCmd)
}

func runWatch(cmd *cobra.Command, args []string) error {
	if err := auth.RequireAuth(); err != nil {
		output.PrintError(err.Error())
		return nil
	}

	cfg, err := config.Load("rukkie.yaml")
	if err != nil {
		// No local rukkie.yaml — watch cloud services instead
		m := output.NewCloudWatchModel(watchIntervalFlag)
		p := tea.NewProgram(m, tea.WithAltScreen())
		_, err = p.Run()
		return err
	}

	services, err := cfg.GetServices(envFlag)
	if err != nil {
		output.PrintError(err.Error())
		return nil
	}

	m := output.NewWatchModel(cfg, services, envFlag, watchIntervalFlag)
	p := tea.NewProgram(m, tea.WithAltScreen())
	_, err = p.Run()
	return err
}
