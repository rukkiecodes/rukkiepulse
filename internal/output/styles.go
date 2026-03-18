package output

import "github.com/charmbracelet/lipgloss"

var (
	// palette
	cGreen  = lipgloss.Color("#39d353")
	cRed    = lipgloss.Color("#f85149")
	cYellow = lipgloss.Color("#d29922")
	cBlue   = lipgloss.Color("#58a6ff")
	cGray   = lipgloss.Color("#8b949e")
	cWhite  = lipgloss.Color("#e6edf3")
	cBorder = lipgloss.Color("#30363d")
	cBg2    = lipgloss.Color("#161b22")

	// base text styles
	sGreen  = lipgloss.NewStyle().Foreground(cGreen)
	sRed    = lipgloss.NewStyle().Foreground(cRed)
	sYellow = lipgloss.NewStyle().Foreground(cYellow)
	sBlue   = lipgloss.NewStyle().Foreground(cBlue)
	sGray   = lipgloss.NewStyle().Foreground(cGray)
	sWhite  = lipgloss.NewStyle().Foreground(cWhite)
	sBold   = lipgloss.NewStyle().Bold(true)
	sDim    = lipgloss.NewStyle().Foreground(cGray).Faint(true)

	// header: project name
	sProjectName = lipgloss.NewStyle().
			Bold(true).
			Foreground(cGreen)

	// header: env badge
	sEnvBadge = lipgloss.NewStyle().
			Foreground(cGray).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cBorder).
			Padding(0, 1)

	// divider line
	sDivider = lipgloss.NewStyle().Foreground(cBorder)

	// service name column (fixed width)
	sServiceName = lipgloss.NewStyle().
			Width(28).
			Foreground(cWhite)

	// latency column (fixed width)
	sLatency = lipgloss.NewStyle().Width(10)

	// endpoint summary
	sEpGood = lipgloss.NewStyle().Foreground(cGreen)
	sEpWarn = lipgloss.NewStyle().Foreground(cYellow)
	sEpFail = lipgloss.NewStyle().Foreground(cRed)

	// inspect box
	sInspectBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cBorder).
			Padding(1, 3).
			Margin(1, 2)

	sInspectTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(cGreen).
			MarginBottom(1)

	sLabel = lipgloss.NewStyle().
		Width(14).
		Foreground(cGray)

	// error block
	sErrorBlock = lipgloss.NewStyle().
			Foreground(cRed).
			Border(lipgloss.NormalBorder(), false, false, false, true).
			BorderForeground(cRed).
			PaddingLeft(1).
			Margin(0, 2)

	// summary bar
	sSummaryOk   = lipgloss.NewStyle().Foreground(cGreen).Bold(true)
	sSummaryWarn = lipgloss.NewStyle().Foreground(cYellow).Bold(true)
	sSummaryDown = lipgloss.NewStyle().Foreground(cRed).Bold(true)

	// login
	sLoginPrompt = lipgloss.NewStyle().
			Foreground(cBlue).
			Bold(true)

	sLoginSuccess = lipgloss.NewStyle().
			Foreground(cGreen).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cGreen).
			Padding(0, 2).
			Margin(1, 2)

	sLoginError = lipgloss.NewStyle().
			Foreground(cRed).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(cRed).
			Padding(0, 2).
			Margin(1, 2)
)
