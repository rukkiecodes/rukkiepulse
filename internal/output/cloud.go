package output

import (
	"fmt"
	"strings"

	"github.com/rukkiecodes/rukkiepulse/internal/cloud"
)

// PrintCloudServices prints all registered services with their connection status.
func PrintCloudServices(services []cloud.ServiceStatus) {
	title := sProjectName.Render("RukkiePulse") + "  " + sGray.Render("connected services")
	fmt.Printf("\n  %s\n\n", title)

	div := sDivider.Render(strings.Repeat("─", 64))
	fmt.Printf("  %s\n", div)

	if len(services) == 0 {
		fmt.Printf("  %s\n", sGray.Render("No services registered yet — visit the dashboard to add one."))
		fmt.Printf("  %s\n\n", sDim.Render("https://rukkiepulse-dashboard.netlify.app"))
		return
	}

	live, recent, inactive, never := 0, 0, 0, 0

	for _, svc := range services {
		dot, label, lastSeen := cloud.ConnectionStatus(svc.LastUsedAt)

		name := sServiceName.Render(svc.Name)
		lang := sGray.Render(svc.Language)

		statusText := ""
		switch dot {
		case "🟢":
			statusText = sGreen.Render(label)
			live++
		case "🟡":
			statusText = sYellow.Render(label)
			recent++
		case "🔴":
			statusText = sRed.Render(label)
			inactive++
		default:
			statusText = sGray.Render(label)
			never++
		}

		seenPart := ""
		if lastSeen != "" {
			seenPart = "  " + sDim.Render(lastSeen)
		}

		keysPart := ""
		if svc.ActiveKeys > 0 {
			keysPart = "  " + sGray.Render(fmt.Sprintf("%d key(s)", svc.ActiveKeys))
		}

		fmt.Printf("  %s  %s  %s  %s%s%s\n", dot, name, lang, statusText, keysPart, seenPart)
	}

	fmt.Printf("  %s\n", div)

	parts := []string{}
	if live > 0 {
		parts = append(parts, sSummaryOk.Render(fmt.Sprintf("🟢 %d live", live)))
	}
	if recent > 0 {
		parts = append(parts, sSummaryWarn.Render(fmt.Sprintf("🟡 %d recent", recent)))
	}
	if inactive > 0 {
		parts = append(parts, sSummaryDown.Render(fmt.Sprintf("🔴 %d inactive", inactive)))
	}
	if never > 0 {
		parts = append(parts, sGray.Render(fmt.Sprintf("⚫ %d never connected", never)))
	}

	fmt.Printf("\n  %s\n\n", strings.Join(parts, sGray.Render("  ·  ")))
}
