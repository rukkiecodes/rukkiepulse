package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rukkiecodes/rukkiepulse/internal/cloud"
)

// ── messages ─────────────────────────────────────────────────────────────────

type cloudScanDoneMsg []cloud.ServiceStatus
type cloudTickMsg struct{}

// ── model ────────────────────────────────────────────────────────────────────

type CloudWatchModel struct {
	services []cloud.ServiceStatus
	spinner  spinner.Model
	scanning bool
	lastRun  time.Time
	interval time.Duration
	width    int
	err      string
}

func NewCloudWatchModel(interval time.Duration) CloudWatchModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(cBlue)

	return CloudWatchModel{
		interval: interval,
		scanning: true,
		width:    80,
		spinner:  s,
	}
}

func (m CloudWatchModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, doCloudScan())
}

func (m CloudWatchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width

	case tea.KeyMsg:
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

	case spinner.TickMsg:
		if m.scanning {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case cloudScanDoneMsg:
		m.services = []cloud.ServiceStatus(msg)
		m.scanning = false
		m.lastRun = time.Now()
		m.err = ""
		return m, tickCloudAfter(m.interval)

	case cloudTickMsg:
		m.scanning = true
		return m, tea.Batch(m.spinner.Tick, doCloudScan())
	}

	return m, nil
}

func (m CloudWatchModel) View() string {
	w := m.width
	if w < 40 {
		w = 40
	}

	var b strings.Builder

	// ── header ────────────────────────────────────────────────────────────────
	title := sProjectName.Render("RukkiePulse") + "  " + sGray.Render("connected services")

	statusLine := ""
	if m.scanning {
		statusLine = m.spinner.View() + " " + sBlue.Render("fetching…")
	} else {
		statusLine = sGray.Render("Updated " + m.lastRun.Format("15:04:05"))
	}

	interval := sDim.Render(fmt.Sprintf("↻ %s", m.interval))
	quit := sDim.Render("  q to stop")

	header := lipgloss.NewStyle().
		Width(w).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(cBorder).
		Padding(1, 2).
		Render(fmt.Sprintf("%s\n%s  %s%s", title, statusLine, interval, quit))

	b.WriteString(header + "\n")

	if m.err != "" {
		b.WriteString(sRed.Render("\n  "+m.err) + "\n")
		return b.String()
	}

	if len(m.services) == 0 {
		b.WriteString(sGray.Render("\n  No services registered yet.\n"))
		return b.String()
	}

	div := sDivider.Render(strings.Repeat("─", min(w-4, 72)))
	b.WriteString("  " + div + "\n")

	live, recent, inactive, never := 0, 0, 0, 0

	for _, svc := range m.services {
		dot, label, lastSeen := cloud.ConnectionStatus(svc.LastUsedAt)

		name := lipgloss.NewStyle().Width(28).Foreground(cWhite).Render(svc.Name)
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

		b.WriteString(fmt.Sprintf("  %s  %s  %s  %s%s\n", dot, name, lang, statusText, seenPart))
	}

	b.WriteString("  " + div + "\n")

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

	summary := lipgloss.NewStyle().Padding(1, 2).
		Render(strings.Join(parts, sGray.Render("  ·  ")))
	b.WriteString(summary)

	return b.String()
}

// ── commands ─────────────────────────────────────────────────────────────────

func doCloudScan() tea.Cmd {
	return func() tea.Msg {
		svcs, _ := cloud.FetchServices()
		return cloudScanDoneMsg(svcs)
	}
}

func tickCloudAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return cloudTickMsg{}
	})
}
