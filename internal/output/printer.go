package output

import (
	"fmt"
	"time"

	"github.com/rukkiecodes/rukkiepulse/internal/engine"
	"github.com/rukkiecodes/rukkiepulse/internal/probe"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	bold        = "\033[1m"
)

// PrintScanHeader prints the project + environment header.
func PrintScanHeader(project, env string) {
	fmt.Printf("\n%s%s%s  [%s]\n\n", bold, project, colorReset, env)
}

// PrintResults prints the scan summary table.
func PrintResults(results []engine.ServiceResult) {
	for _, r := range results {
		printServiceRow(r)
	}
	fmt.Println()
}

func printServiceRow(r engine.ServiceResult) {
	icon, color := healthDisplay(r)
	latency := formatLatency(r.Health.Latency)

	errMsg := ""
	if r.Health.Error != "" && r.Health.Status == "down" {
		errMsg = fmt.Sprintf("  %s(%s)%s", colorRed, r.Health.Error, colorReset)
	}

	endpointSummary := ""
	if total := len(r.Endpoints); total > 0 {
		pass, _ := r.PassingEndpoints()
		epColor := colorGreen
		if pass < total {
			epColor = colorYellow
		}
		if pass == 0 {
			epColor = colorRed
		}
		endpointSummary = fmt.Sprintf("  %s%d/%d endpoints ok%s", epColor, pass, total, colorReset)
	}

	fmt.Printf("  %s %-25s %s%-8s%s%s%s\n",
		icon,
		r.Name,
		color, latency, colorReset,
		endpointSummary,
		errMsg,
	)
}

func healthDisplay(r engine.ServiceResult) (icon, color string) {
	switch r.Health.Status {
	case "ok":
		if r.Health.Latency >= 500*time.Millisecond {
			return "🟡", colorYellow
		}
		return "🟢", colorGreen
	case "degraded":
		return "🟡", colorYellow
	default:
		return "🔴", colorRed
	}
}

// PrintInspect prints the full detail view for one service.
func PrintInspect(r engine.ServiceResult) {
	fmt.Printf("\n%sService:%s %s\n", bold, colorReset, r.Name)
	fmt.Printf("%sURL:%s     %s\n", bold, colorReset, r.URL)
	fmt.Printf("%sType:%s    %s\n\n", bold, colorReset, r.Type)

	icon, color := healthDisplay(r)
	latency := formatLatency(r.Health.Latency)
	fmt.Printf("%sHealth:%s  %s %s%s%s  %s\n",
		bold, colorReset,
		icon,
		color, r.Health.Status, colorReset,
		latency,
	)

	if r.Health.Error != "" {
		fmt.Printf("         %s%s%s\n", colorRed, r.Health.Error, colorReset)
	}

	if len(r.Health.Dependencies) > 0 {
		fmt.Printf("\n%sDependencies:%s\n", bold, colorReset)
		for dep, status := range r.Health.Dependencies {
			c := colorGreen
			if status != "connected" && status != "ok" {
				c = colorRed
			}
			fmt.Printf("  %-10s %s%s%s\n", dep+":", c, status, colorReset)
		}
	}

	if len(r.Endpoints) > 0 {
		fmt.Printf("\n%sEndpoints:%s\n", bold, colorReset)
		for _, ep := range r.Endpoints {
			printEndpointRow(ep)
		}
	}

	fmt.Println()
}

func printEndpointRow(ep probe.EndpointResult) {
	icon := "🟢"
	color := colorGreen
	if ep.Status == "fail" {
		icon = "🔴"
		color = colorRed
	}

	code := ""
	if ep.Code > 0 {
		code = fmt.Sprintf("%d", ep.Code)
	}

	errMsg := ""
	if ep.Error != "" {
		errMsg = fmt.Sprintf("  %s%s%s", colorRed, ep.Error, colorReset)
	}

	kind := ep.Kind
	if kind == "GRAPHQL" {
		fmt.Printf("  %s  %-8s %-30s %s%-6s%s %s%s\n",
			icon, "GRAPHQL", ep.Path,
			color, formatLatency(ep.Latency), colorReset,
			code, errMsg,
		)
	} else {
		fmt.Printf("  %s  %-7s %-30s %s%-6s%s %s%s\n",
			icon, ep.Method, ep.Path,
			color, formatLatency(ep.Latency), colorReset,
			code, errMsg,
		)
	}
}

// PrintAllClear prints a message when --errors-only finds nothing wrong.
func PrintAllClear() {
	fmt.Printf("  %s✅ All services are healthy%s\n\n", colorGreen, colorReset)
}

// PrintError prints a top-level error message.
func PrintError(msg string) {
	fmt.Printf("\n%s%s error:%s %s\n\n", bold, colorRed, colorReset, msg)
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
