package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/rukkiecodes/rukkiepulse/internal/engine"
	"github.com/rukkiecodes/rukkiepulse/internal/probe"
)

// PrintScanHeader prints the project + environment header.
func PrintScanHeader(project, env string) {
	name := sProjectName.Render(project)
	badge := sEnvBadge.Render(env)
	fmt.Printf("\n  %s  %s\n\n", name, badge)
}

// PrintResults prints the scan summary table.
func PrintResults(results []engine.ServiceResult) {
	div := sDivider.Render(strings.Repeat("─", 64))
	fmt.Printf("  %s\n", div)

	for _, r := range results {
		printServiceRow(r)
	}

	fmt.Printf("  %s\n", div)
	printSummary(results)
	fmt.Println()
}

func printServiceRow(r engine.ServiceResult) {
	icon := statusIcon(r)
	latency := sLatency.Render(formatLatency(r.Health.Latency))
	name := sServiceName.Render(r.Name)

	epSummary := endpointSummary(r)

	errPart := ""
	if r.Health.Error != "" && r.Health.Status != "ok" {
		errPart = "  " + sDim.Render("← "+truncate(r.Health.Error, 40))
	}

	fmt.Printf("  %s  %s  %s  %s%s\n", icon, name, latency, epSummary, errPart)
}

func endpointSummary(r engine.ServiceResult) string {
	total := len(r.Endpoints)
	if total == 0 {
		return sGray.Render("—")
	}
	pass, _ := r.PassingEndpoints()
	text := fmt.Sprintf("%d/%d endpoints ok", pass, total)
	switch {
	case pass == total:
		return sEpGood.Render(text)
	case pass == 0:
		return sEpFail.Render(text)
	default:
		return sEpWarn.Render(text)
	}
}

func printSummary(results []engine.ServiceResult) {
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

	parts := []string{}
	if ok > 0 {
		parts = append(parts, sSummaryOk.Render(fmt.Sprintf("✓ %d ok", ok)))
	}
	if degraded > 0 {
		parts = append(parts, sSummaryWarn.Render(fmt.Sprintf("⚡ %d degraded", degraded)))
	}
	if down > 0 {
		parts = append(parts, sSummaryDown.Render(fmt.Sprintf("✗ %d down", down)))
	}

	fmt.Printf("\n  %s\n", strings.Join(parts, sGray.Render("  ·  ")))
}

// PrintInspect prints the full detail view for one service.
func PrintInspect(r engine.ServiceResult) {
	var b strings.Builder

	b.WriteString(sInspectTitle.Render(r.Name) + "\n")
	b.WriteString(sLabel.Render("URL") + sWhite.Render(r.URL) + "\n")
	b.WriteString(sLabel.Render("Type") + sWhite.Render(r.Type) + "\n\n")

	icon := statusIcon(r)
	b.WriteString(sLabel.Render("Health") + icon + "  " + healthText(r) + "  " + sGray.Render(formatLatency(r.Health.Latency)) + "\n")

	if r.Health.Error != "" {
		b.WriteString("\n" + sRed.Render(r.Health.Error) + "\n")
	}

	if len(r.Health.Dependencies) > 0 {
		b.WriteString("\n" + sBold.Render("Dependencies") + "\n")
		for dep, status := range r.Health.Dependencies {
			c := sGreen
			if status != "connected" && status != "ok" {
				c = sRed
			}
			b.WriteString(sGray.Render("  "+dep+":") + "  " + c.Render(status) + "\n")
		}
	}

	if len(r.Endpoints) > 0 {
		b.WriteString("\n" + sBold.Render("Endpoints") + "\n")
		div := sDivider.Render(strings.Repeat("─", 52))
		b.WriteString(div + "\n")
		for _, ep := range r.Endpoints {
			b.WriteString(inspectEndpointRow(ep) + "\n")
		}
		b.WriteString(div + "\n")
	}

	box := sInspectBox.Render(b.String())
	fmt.Println(box)
}

func inspectEndpointRow(ep probe.EndpointResult) string {
	icon := "🟢"
	codeStyle := sGreen
	if ep.Status == "fail" {
		icon = "🔴"
		codeStyle = sRed
	}

	method := lipgloss.NewStyle().Width(8).Foreground(cBlue).Render(ep.Method)
	if ep.Kind == "GRAPHQL" {
		method = lipgloss.NewStyle().Width(8).Foreground(cYellow).Render("GRAPHQL")
	}

	path := lipgloss.NewStyle().Width(30).Render(ep.Path)
	latency := lipgloss.NewStyle().Width(8).Foreground(cGray).Render(formatLatency(ep.Latency))
	code := ""
	if ep.Code > 0 {
		code = codeStyle.Render(fmt.Sprintf("%d", ep.Code))
	}
	errPart := ""
	if ep.Error != "" {
		errPart = "  " + sDim.Render(ep.Error)
	}

	return fmt.Sprintf("  %s  %s%s%s  %s%s", icon, method, path, latency, code, errPart)
}

// PrintAllClear prints a message when --errors-only finds nothing wrong.
func PrintAllClear() {
	fmt.Println(sLoginSuccess.Render("✓  All services are healthy"))
}

// PrintError prints a top-level error message.
func PrintError(msg string) {
	fmt.Println(sLoginError.Render("✗  " + msg))
}

// PrintLoginSuccess prints the login success message.
func PrintLoginSuccess() {
	fmt.Println(sLoginSuccess.Render("✓  Logged in to RukkiePulse"))
}

// PrintLogout prints the logout message.
func PrintLogout() {
	fmt.Println(sGray.Render("\n  Logged out.\n"))
}

// PrintPasswordPrompt prints the styled password prompt.
func PrintPasswordPrompt() {
	fmt.Print(sLoginPrompt.Render("  Password: "))
}

func statusIcon(r engine.ServiceResult) string {
	switch r.Health.Status {
	case "ok":
		if r.Health.Latency >= 500*time.Millisecond {
			return "🟡"
		}
		return "🟢"
	case "degraded":
		return "🟡"
	default:
		return "🔴"
	}
}

func healthText(r engine.ServiceResult) string {
	switch r.Health.Status {
	case "ok":
		if r.Health.Latency >= 500*time.Millisecond {
			return sYellow.Render("slow")
		}
		return sGreen.Render("ok")
	case "degraded":
		return sYellow.Render("degraded")
	default:
		return sRed.Render("down")
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

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
