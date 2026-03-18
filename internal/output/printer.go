package output

import (
	"fmt"
	"time"

	"github.com/rukkiecodes/rukkiepulse/internal/health"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	bold        = "\033[1m"
)

func PrintScanHeader(project, env string) {
	fmt.Printf("\n%s%s%s  [%s]\n\n", bold, project, colorReset, env)
}

func PrintResults(results []health.Result) {
	for _, r := range results {
		printResult(r)
	}
	fmt.Println()
}

func printResult(r health.Result) {
	icon, color := statusDisplay(r)

	latency := ""
	if r.Status != "down" || r.Latency > 0 {
		latency = formatLatency(r.Latency)
	}

	errMsg := ""
	if r.Error != "" {
		errMsg = fmt.Sprintf("  %s(%s)%s", colorRed, r.Error, colorReset)
	}

	fmt.Printf("  %s %-25s %s%s%s%s\n",
		icon,
		r.Service,
		color,
		latency,
		colorReset,
		errMsg,
	)
}

func statusDisplay(r health.Result) (icon, color string) {
	switch r.Status {
	case "ok":
		if r.Latency >= 500*time.Millisecond {
			return "🟡", colorYellow
		}
		return "🟢", colorGreen
	case "degraded":
		return "🟡", colorYellow
	default:
		return "🔴", colorRed
	}
}

func formatLatency(d time.Duration) string {
	if d == 0 {
		return "—"
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

func PrintError(msg string) {
	fmt.Printf("\n%s%s error:%s %s\n\n", bold, colorRed, colorReset, msg)
}
