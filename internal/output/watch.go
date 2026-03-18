package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/rukkiecodes/rukkiepulse/internal/config"
	"github.com/rukkiecodes/rukkiepulse/internal/engine"
)

// ── messages ────────────────────────────────────────────────────────────────

type scanDoneMsg []engine.ServiceResult
type tickMsg struct{}

// ── model ────────────────────────────────────────────────────────────────────

type WatchModel struct {
	cfg      *config.Config
	services []config.Service
	results  []engine.ServiceResult
	spinner  spinner.Model
	scanning bool
	lastRun  time.Time
	interval time.Duration
	env      string
	width    int
}

func NewWatchModel(cfg *config.Config, services []config.Service, env string, interval time.Duration) WatchModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(cBlue)

	return WatchModel{
		cfg:      cfg,
		services: services,
		env:      env,
		interval: interval,
		scanning: true,
		width:    80,
		spinner:  s,
	}
}

func (m WatchModel) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		doScan(m.services),
	)
}

func (m WatchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

	case scanDoneMsg:
		m.results = []engine.ServiceResult(msg)
		m.scanning = false
		m.lastRun = time.Now()
		return m, tickAfter(m.interval)

	case tickMsg:
		m.scanning = true
		return m, tea.Batch(m.spinner.Tick, doScan(m.services))
	}

	return m, nil
}

func (m WatchModel) View() string {
	w := m.width
	if w < 40 {
		w = 40
	}

	var b strings.Builder

	// ── header ──────────────────────────────────────────────────────────────
	project := sProjectName.Render(m.cfg.Project)
	badge := sEnvBadge.Render(m.env)

	statusLine := ""
	if m.scanning {
		statusLine = m.spinner.View() + " " + sBlue.Render("scanning…")
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
		Render(fmt.Sprintf("%s  %s\n%s  %s%s", project, badge, statusLine, interval, quit))

	b.WriteString(header + "\n")

	if len(m.results) == 0 {
		b.WriteString(sGray.Render("\n  Waiting for first scan…\n"))
		return b.String()
	}

	// ── column headers ───────────────────────────────────────────────────────
	colHeader := lipgloss.NewStyle().
		Foreground(cGray).
		Bold(true).
		Padding(0, 2).
		Render(fmt.Sprintf(
			"  %-3s  %-28s  %-10s  %-20s  %s",
			"", "Service", "Latency", "Endpoints", "Status",
		))
	b.WriteString(colHeader + "\n")

	div := sDivider.Render(strings.Repeat("─", min(w-4, 72)))
	b.WriteString("  " + div + "\n")

	// ── rows ─────────────────────────────────────────────────────────────────
	for _, r := range m.results {
		b.WriteString(watchRow(r) + "\n")
	}

	b.WriteString("  " + div + "\n")

	// ── summary ──────────────────────────────────────────────────────────────
	ok, degraded, down := 0, 0, 0
	for _, r := range m.results {
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

	summary := lipgloss.NewStyle().Padding(1, 2).
		Render(strings.Join(parts, sGray.Render("  ·  ")))
	b.WriteString(summary)

	return b.String()
}

func watchRow(r engine.ServiceResult) string {
	icon := statusIcon(r)
	name := lipgloss.NewStyle().Width(28).Foreground(cWhite).Render(r.Name)
	lat := lipgloss.NewStyle().Width(10).Render(formatLatency(r.Health.Latency))
	ep := endpointSummary(r)

	errPart := ""
	if r.Health.Error != "" && r.Health.Status != "ok" {
		errPart = "  " + sDim.Render("← "+truncate(r.Health.Error, 35))
	}

	return fmt.Sprintf("  %s  %s  %s  %s%s", icon, name, lat, ep, errPart)
}

// ── commands ─────────────────────────────────────────────────────────────────

func doScan(services []config.Service) tea.Cmd {
	return func() tea.Msg {
		return scanDoneMsg(engine.Run(services))
	}
}

func tickAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
