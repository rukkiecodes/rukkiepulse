package output

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ── palette ──────────────────────────────────────────────────────────────────

var (
	shellBg     = lipgloss.Color("#050f09")
	shellGreen  = lipgloss.Color("#39d353")
	shellOrange = lipgloss.Color("#f7a541")
	shellGray   = lipgloss.Color("#8b949e")
	shellWhite  = lipgloss.Color("#e6edf3")
	shellRed    = lipgloss.Color("#f85149")
	shellYellow = lipgloss.Color("#d29922")
	shellBlue   = lipgloss.Color("#58a6ff")
	shellBorder = lipgloss.Color("#0d2b18")
)

// ── ASCII logo ────────────────────────────────────────────────────────────────

const logo = `
 ██████╗ ██╗   ██╗██╗  ██╗██╗  ██╗██╗███████╗
 ██╔══██╗██║   ██║██║ ██╔╝██║ ██╔╝██║██╔════╝
 ██████╔╝██║   ██║█████╔╝ █████╔╝ ██║█████╗
 ██╔══██╗██║   ██║██╔═██╗ ██╔═██╗ ██║██╔══╝
 ██║  ██║╚██████╔╝██║  ██╗██║  ██╗██║███████╗
 ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝╚═╝  ╚═╝╚═╝╚══════╝`

// ── state ────────────────────────────────────────────────────────────────────

type shellState int

const (
	stateReady    shellState = iota
	stateRunning             // command executing
	statePassword            // awaiting password for login
)

// ── messages ─────────────────────────────────────────────────────────────────

type cmdOutputMsg struct {
	input  string
	output string
	isErr  bool
}

type loginDoneMsg struct {
	output string
	isErr  bool
}

// ── model ────────────────────────────────────────────────────────────────────

type ShellModel struct {
	input   textinput.Model
	vp      viewport.Model
	spinner spinner.Model
	lines   []string // rendered history lines
	state   shellState
	width   int
	height  int
	ready   bool
	cwd     string
	exe     string // path to rukkie binary
}

func NewShellModel() ShellModel {
	ti := textinput.New()
	ti.Focus()
	ti.TextStyle = lipgloss.NewStyle().Foreground(shellWhite).Background(shellBg)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(shellGreen).Background(shellBg)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(shellGreen)

	cwd, _ := os.Getwd()
	exe, _ := os.Executable()

	return ShellModel{
		input:   ti,
		spinner: sp,
		cwd:     cwd,
		exe:     exe,
		lines:   []string{renderWelcome()},
	}
}

// ── init ──────────────────────────────────────────────────────────────────────

func (m ShellModel) Init() tea.Cmd {
	return textinput.Blink
}

// ── update ────────────────────────────────────────────────────────────────────

func (m ShellModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerH := logoHeight() + 2
		promptH := 3
		vpH := m.height - headerH - promptH
		if vpH < 1 {
			vpH = 1
		}
		if !m.ready {
			m.vp = viewport.New(m.width-4, vpH)
			m.ready = true
		} else {
			m.vp.Width = m.width - 4
			m.vp.Height = vpH
		}
		m.vp.SetContent(strings.Join(m.lines, "\n"))
		m.vp.GotoBottom()

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			if m.state == stateReady || m.state == statePassword {
				return m, tea.Quit
			}

		case "enter":
			raw := strings.TrimSpace(m.input.Value())
			m.input.SetValue("")

			if raw == "" {
				return m, nil
			}

			switch m.state {
			case statePassword:
				m.state = stateRunning
				password := raw
				exe := m.exe
				cmds = append(cmds, func() tea.Msg {
					out, err := runRukkieStdin(exe, password, "login")
					isErr := err != nil
					return loginDoneMsg{output: out, isErr: isErr}
				})
				cmds = append(cmds, m.spinner.Tick)
				return m, tea.Batch(cmds...)

			case stateReady:
				lower := strings.ToLower(raw)
				if lower == "exit" || lower == "quit" {
					return m, tea.Quit
				}
				if lower == "clear" {
					m.lines = []string{}
					m.vp.SetContent("")
					return m, nil
				}
				if lower == "login" {
					m.appendLine(colorizeInput(raw))
					m.appendLine(lipgloss.NewStyle().Foreground(shellBlue).Bold(true).Render("  Password: "))
					m.state = statePassword
					m.input.EchoMode = textinput.EchoPassword
					m.input.EchoCharacter = '•'
					m.vp.SetContent(strings.Join(m.lines, "\n"))
					m.vp.GotoBottom()
					return m, nil
				}

				m.appendLine(renderPromptLine(m.cwd) + " " + colorizeInput(raw))
				m.state = stateRunning
				parts := strings.Fields(raw)
				exe := m.exe
				cmds = append(cmds, func() tea.Msg {
					out, err := runRukkie(exe, parts...)
					return cmdOutputMsg{input: raw, output: out, isErr: err != nil}
				})
				cmds = append(cmds, m.spinner.Tick)
				return m, tea.Batch(cmds...)
			}

		default:
			if m.state == stateReady || m.state == statePassword {
				var inputCmd tea.Cmd
				m.input, inputCmd = m.input.Update(msg)
				cmds = append(cmds, inputCmd)
			}
		}

	case spinner.TickMsg:
		if m.state == stateRunning {
			var spCmd tea.Cmd
			m.spinner, spCmd = m.spinner.Update(msg)
			cmds = append(cmds, spCmd)
		}

	case cmdOutputMsg:
		m.state = stateReady
		m.input.EchoMode = textinput.EchoNormal
		if msg.output != "" {
			for _, line := range strings.Split(strings.TrimRight(msg.output, "\n"), "\n") {
				m.appendLine("  " + line)
			}
		}
		m.appendLine("") // blank spacer
		m.vp.SetContent(strings.Join(m.lines, "\n"))
		m.vp.GotoBottom()

	case loginDoneMsg:
		m.state = stateReady
		m.input.EchoMode = textinput.EchoNormal
		m.input.EchoCharacter = 0
		if msg.output != "" {
			for _, line := range strings.Split(strings.TrimRight(msg.output, "\n"), "\n") {
				m.appendLine("  " + line)
			}
		}
		m.appendLine("")
		m.vp.SetContent(strings.Join(m.lines, "\n"))
		m.vp.GotoBottom()
	}

	if m.ready {
		var vpCmd tea.Cmd
		m.vp, vpCmd = m.vp.Update(msg)
		cmds = append(cmds, vpCmd)
	}

	return m, tea.Batch(cmds...)
}

// ── view ──────────────────────────────────────────────────────────────────────

func (m ShellModel) View() string {
	if !m.ready {
		return ""
	}

	bg := lipgloss.NewStyle().Background(shellBg)
	full := lipgloss.NewStyle().
		Background(shellBg).
		Width(m.width).
		Height(m.height)

	// ── header ──────────────────────────────────────────────────────────────
	logoStyled := lipgloss.NewStyle().
		Foreground(shellGreen).
		Background(shellBg).
		Bold(true).
		Render(logo)

	pulse := lipgloss.NewStyle().
		Foreground(shellOrange).
		Background(shellBg).
		Bold(true).
		MarginLeft(2).
		Render("pulse  — CLI Observability for Backend Services")

	divider := lipgloss.NewStyle().
		Foreground(shellBorder).
		Background(shellBg).
		Width(m.width).
		Render(strings.Repeat("─", m.width))

	header := bg.Render(logoStyled+"\n"+pulse+"\n") + divider + "\n"

	// ── output viewport ─────────────────────────────────────────────────────
	vpStyled := lipgloss.NewStyle().
		Background(shellBg).
		Padding(0, 2).
		Render(m.vp.View())

	// ── prompt ───────────────────────────────────────────────────────────────
	var promptPrefix string
	switch m.state {
	case stateRunning:
		promptPrefix = "  " + m.spinner.View() + " " +
			lipgloss.NewStyle().Foreground(shellBlue).Background(shellBg).Render("running…")
	case statePassword:
		promptPrefix = lipgloss.NewStyle().Foreground(shellBlue).Bold(true).Background(shellBg).Render("  Password: ")
	default:
		promptPrefix = renderPromptLine(m.cwd)
	}

	inputLine := promptPrefix + " " + m.input.View()

	promptDiv := lipgloss.NewStyle().
		Foreground(shellBorder).
		Background(shellBg).
		Width(m.width).
		Render(strings.Repeat("─", m.width))

	promptBlock := lipgloss.NewStyle().Background(shellBg).Width(m.width).
		Render("\n" + promptDiv + "\n" + inputLine)

	content := header + vpStyled + promptBlock

	return full.Render(content)
}

// ── helpers ───────────────────────────────────────────────────────────────────

func (m *ShellModel) appendLine(line string) {
	m.lines = append(m.lines, line)
}

func renderPromptLine(cwd string) string {
	parts := strings.Split(cwd, string(os.PathSeparator))
	if len(parts) == 0 {
		parts = []string{cwd}
	}

	var segments []string
	sep := lipgloss.NewStyle().Foreground(shellGray).Background(shellBg).Render(string(os.PathSeparator))

	for i, p := range parts {
		if p == "" {
			continue
		}
		if i == 0 {
			// drive root e.g. "C:"
			segments = append(segments, lipgloss.NewStyle().
				Foreground(shellGreen).Background(shellBg).Bold(true).
				Render(p+string(os.PathSeparator)))
		} else {
			segments = append(segments, lipgloss.NewStyle().
				Foreground(shellOrange).Background(shellBg).
				Render(p))
			if i < len(parts)-1 {
				segments = append(segments, sep)
			}
		}
	}

	path := strings.Join(segments, "")
	arrow := lipgloss.NewStyle().Foreground(shellWhite).Background(shellBg).Bold(true).Render(" ❯")
	return "  " + path + arrow
}

func renderWelcome() string {
	lines := []string{
		lipgloss.NewStyle().Foreground(shellGray).Background(shellBg).Render("  Type a command to get started. Try: scan  status  inspect  trace  help"),
		lipgloss.NewStyle().Foreground(shellGray).Background(shellBg).Render("  Type exit to quit."),
	}
	return strings.Join(lines, "\n")
}

func colorizeInput(raw string) string {
	parts := strings.Fields(raw)
	if len(parts) == 0 {
		return raw
	}
	known := map[string]bool{
		"scan": true, "status": true, "inspect": true,
		"trace": true, "watch": true, "login": true,
		"logout": true, "help": true, "clear": true, "exit": true,
	}

	var out []string
	for i, p := range parts {
		switch {
		case i == 0 && known[p]:
			out = append(out, lipgloss.NewStyle().Foreground(shellGreen).Background(shellBg).Bold(true).Render(p))
		case i == 0:
			out = append(out, lipgloss.NewStyle().Foreground(shellRed).Background(shellBg).Render(p))
		case strings.HasPrefix(p, "--") || strings.HasPrefix(p, "-"):
			out = append(out, lipgloss.NewStyle().Foreground(shellYellow).Background(shellBg).Render(p))
		default:
			out = append(out, lipgloss.NewStyle().Foreground(shellWhite).Background(shellBg).Render(p))
		}
	}
	return strings.Join(out, " ")
}

func runRukkie(exe string, args ...string) (string, error) {
	cmd := exec.Command(exe, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

// runRukkieStdin pipes stdin to the subprocess — used for login so the
// password is passed via stdin rather than a flag.
func runRukkieStdin(exe, stdin string, args ...string) (string, error) {
	cmd := exec.Command(exe, args...)
	cmd.Stdin = strings.NewReader(stdin + "\n")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err
}

func logoHeight() int {
	return strings.Count(logo, "\n") + 3 // logo + pulse line + divider
}

func renderHelp() string {
	cmds := [][]string{
		{"scan", "Health check + probe all services"},
		{"status", "Alias for scan"},
		{"inspect <service>", "Deep dive into one service"},
		{"trace <service>", "Show distributed traces from Jaeger"},
		{"watch", "Live-updating dashboard"},
		{"login", "Authenticate"},
		{"logout", "Clear session"},
		{"clear", "Clear screen"},
		{"exit", "Quit"},
	}
	var b strings.Builder
	b.WriteString(lipgloss.NewStyle().Foreground(shellGreen).Background(shellBg).Bold(true).Render("  Available commands") + "\n\n")
	for _, c := range cmds {
		name := lipgloss.NewStyle().Width(24).Foreground(shellBlue).Background(shellBg).Render(c[0])
		desc := lipgloss.NewStyle().Foreground(shellGray).Background(shellBg).Render(c[1])
		b.WriteString(fmt.Sprintf("  %s %s\n", name, desc))
	}
	return b.String()
}
