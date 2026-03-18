package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rukkiecodes/rukkiepulse/internal/auth"
	"github.com/rukkiecodes/rukkiepulse/internal/config"
	"github.com/rukkiecodes/rukkiepulse/internal/engine"
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
		output.PrintError(err.Error())
		return nil
	}

	services, err := cfg.GetServices(envFlag)
	if err != nil {
		output.PrintError(err.Error())
		return nil
	}

	// handle Ctrl+C gracefully
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	ticker := time.NewTicker(watchIntervalFlag)
	defer ticker.Stop()

	// run immediately, then on each tick
	renderDashboard(cfg, envFlag, services, watchIntervalFlag)

	for {
		select {
		case <-ticker.C:
			renderDashboard(cfg, envFlag, services, watchIntervalFlag)
		case <-quit:
			// move cursor below dashboard before exiting
			fmt.Print("\n\033[?25h") // restore cursor
			fmt.Println("  Stopped.")
			return nil
		}
	}
}

func renderDashboard(cfg *config.Config, env string, services []config.Service, interval time.Duration) {
	results := engine.Run(services)

	// move cursor to top-left and clear screen
	fmt.Print("\033[H\033[2J")
	// hide cursor while rendering
	fmt.Print("\033[?25l")

	now := time.Now().Format("15:04:05")

	fmt.Printf("\n\033[1m%s\033[0m  [%s]", cfg.Project, env)
	fmt.Printf("   \033[90mRefreshing every %s  (Ctrl+C to stop)\033[0m\n", interval)
	fmt.Printf("  Last updated: \033[36m%s\033[0m\n\n", now)

	// header row
	fmt.Printf("  \033[1m%-24s  %-12s  %-10s  %-20s\033[0m\n",
		"Service", "Status", "Latency", "Endpoints")
	fmt.Printf("  %s\n", strings.Repeat("─", 72))

	for _, r := range results {
		printDashboardRow(r)
	}

	fmt.Printf("  %s\n", strings.Repeat("─", 72))

	// summary counts
	ok, degraded, down := 0, 0, 0
	for _, r := range results {
		switch r.Health.Status {
		case "ok":
			ok++
		case "degraded":
			degraded++
		default:
			down++
		}
	}
	fmt.Printf("\n  \033[32m%d ok\033[0m", ok)
	if degraded > 0 {
		fmt.Printf("   \033[33m%d degraded\033[0m", degraded)
	}
	if down > 0 {
		fmt.Printf("   \033[31m%d down\033[0m", down)
	}
	fmt.Println()

	// restore cursor
	fmt.Print("\033[?25h")
}

func printDashboardRow(r engine.ServiceResult) {
	icon, statusText, color := dashboardStatus(r)
	latency := dashboardLatency(r)

	epSummary := ""
	if total := len(r.Endpoints); total > 0 {
		pass, _ := r.PassingEndpoints()
		epColor := "\033[32m"
		if pass < total {
			epColor = "\033[33m"
		}
		if pass == 0 {
			epColor = "\033[31m"
		}
		epSummary = fmt.Sprintf("%s%d/%d pass\033[0m", epColor, pass, total)
	} else {
		epSummary = "\033[90m—\033[0m"
	}

	fmt.Printf("  %-24s  %s %s%-10s\033[0m  %-10s  %s",
		r.Name,
		icon,
		color, statusText,
		latency,
		epSummary,
	)

	if r.Health.Error != "" {
		fmt.Printf("   \033[90m← %s\033[0m", truncateStr(r.Health.Error, 35))
	}
	fmt.Println()
}

func dashboardStatus(r engine.ServiceResult) (icon, text, color string) {
	switch r.Health.Status {
	case "ok":
		if r.Health.Latency >= 500*time.Millisecond {
			return "🟡", "slow", "\033[33m"
		}
		return "🟢", "ok", "\033[32m"
	case "degraded":
		return "🟡", "degraded", "\033[33m"
	default:
		return "🔴", "down", "\033[31m"
	}
}

func dashboardLatency(r engine.ServiceResult) string {
	if r.Health.Status == "down" {
		return "—"
	}
	d := r.Health.Latency
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
