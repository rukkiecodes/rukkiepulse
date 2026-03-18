package trace

import (
	"fmt"
	"strings"
	"time"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorGray   = "\033[90m"
	bold        = "\033[1m"
	barWidth    = 40 // characters for the ASCII bar area
)

// Render prints a full trace waterfall to stdout.
func Render(t *Trace) {
	if len(t.Spans) == 0 {
		fmt.Println("  (no spans in trace)")
		return
	}

	total := t.TotalDuration()
	statusIcon := "✅"
	statusColor := colorGreen
	if t.HasErrors() {
		statusIcon = "❌"
		statusColor = colorRed
	}

	fmt.Printf("\n%sTrace:%s  %s\n", bold, colorReset, t.TraceID)
	fmt.Printf("%sTotal:%s  %s%s  %s%s\n\n",
		bold, colorReset,
		statusColor, formatDuration(total), statusIcon, colorReset,
	)

	// build children map
	children := make(map[string][]int, len(t.Spans))
	roots := []int{}
	spanIndex := make(map[string]int, len(t.Spans))
	for i, s := range t.Spans {
		spanIndex[s.SpanID] = i
	}
	for i, s := range t.Spans {
		if s.ParentSpanID == "" {
			roots = append(roots, i)
		} else if _, exists := spanIndex[s.ParentSpanID]; exists {
			children[s.ParentSpanID] = append(children[s.ParentSpanID], i)
		} else {
			roots = append(roots, i)
		}
	}

	// find root start time for offset calculations
	rootStart := t.Spans[0].StartTime
	for _, s := range t.Spans {
		if s.StartTime.Before(rootStart) {
			rootStart = s.StartTime
		}
	}

	fmt.Printf("  %-40s  %-*s  %s\n", "Operation", barWidth, "Timeline", "Duration")
	fmt.Printf("  %s  %s  %s\n", strings.Repeat("─", 40), strings.Repeat("─", barWidth), strings.Repeat("─", 10))

	for _, idx := range roots {
		renderSpan(t.Spans, children, idx, 0, rootStart, total)
	}

	// root cause
	if rc := findRootCause(t.Spans); rc != nil {
		fmt.Printf("\n%sRoot Cause:%s\n", bold, colorReset)
		fmt.Printf("  %s❌ %s%s\n", colorRed, rc.OperationName, colorReset)
		if rc.ErrorMessage != "" {
			fmt.Printf("     %s%s%s\n", colorGray, rc.ErrorMessage, colorReset)
		}
	}

	fmt.Println()
}

func renderSpan(
	spans []Span,
	children map[string][]int,
	idx, depth int,
	rootStart time.Time,
	total time.Duration,
) {
	s := spans[idx]
	indent := strings.Repeat("  ", depth)

	label := indent + s.OperationName
	if s.Service != "" {
		label = indent + fmt.Sprintf("%s%s%s", colorGray, s.Service+": ", colorReset) + s.OperationName
	}

	offset := s.StartTime.Sub(rootStart)
	bar := buildBar(offset, s.Duration, total)

	durationStr := formatDuration(s.Duration)
	statusMark := ""
	lineColor := colorReset
	if s.HasError {
		statusMark = " ❌"
		lineColor = colorRed
	}

	fmt.Printf("  %-40s  %s%s%s  %s%s%s%s\n",
		truncate(label, 40),
		lineColor, bar, colorReset,
		lineColor, durationStr, statusMark, colorReset,
	)

	if s.HasError && s.ErrorMessage != "" {
		errIndent := strings.Repeat("  ", depth+1)
		fmt.Printf("  %s%s└─ %s%s\n", errIndent, colorRed, s.ErrorMessage, colorReset)
	}

	for _, childIdx := range children[s.SpanID] {
		renderSpan(spans, children, childIdx, depth+1, rootStart, total)
	}
}

func buildBar(offset, duration, total time.Duration) string {
	if total == 0 {
		return strings.Repeat("█", barWidth)
	}

	startPos := int(float64(offset) / float64(total) * float64(barWidth))
	width := int(float64(duration) / float64(total) * float64(barWidth))

	if width < 1 {
		width = 1
	}
	if startPos+width > barWidth {
		width = barWidth - startPos
	}
	if startPos < 0 {
		startPos = 0
	}

	bar := strings.Repeat(" ", startPos) +
		strings.Repeat("█", width) +
		strings.Repeat(" ", barWidth-startPos-width)
	return bar
}

// findRootCause finds the deepest error span with no error children.
func findRootCause(spans []Span) *Span {
	errorSpanIDs := make(map[string]bool)
	for _, s := range spans {
		if s.HasError {
			errorSpanIDs[s.SpanID] = true
		}
	}

	// build parent→children map
	children := make(map[string][]string)
	for _, s := range spans {
		if s.ParentSpanID != "" {
			children[s.ParentSpanID] = append(children[s.ParentSpanID], s.SpanID)
		}
	}

	// find error spans with no error children
	for i := range spans {
		s := &spans[i]
		if !s.HasError {
			continue
		}
		hasErrorChild := false
		for _, childID := range children[s.SpanID] {
			if errorSpanIDs[childID] {
				hasErrorChild = true
				break
			}
		}
		if !hasErrorChild {
			return s
		}
	}
	return nil
}

func formatDuration(d time.Duration) string {
	if d == 0 {
		return "—"
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%dµs", d.Microseconds())
	}
	return fmt.Sprintf("%dms", d.Milliseconds())
}

func truncate(s string, max int) string {
	// strip ANSI codes for length calculation (rough)
	visible := stripANSI(s)
	if len(visible) <= max {
		return s + strings.Repeat(" ", max-len(visible))
	}
	return s[:max-1] + "…"
}

// RenderFlame prints a horizontal flame graph for a trace.
func RenderFlame(t *Trace) {
	if len(t.Spans) == 0 {
		fmt.Println("  (no spans in trace)")
		return
	}

	total := t.TotalDuration()
	flameWidth := 50

	fmt.Printf("\n%sFlame Graph%s — %s  %s\n\n",
		bold, colorReset,
		t.TraceID,
		formatDuration(total),
	)

	// find root start for offset
	rootStart := t.Spans[0].StartTime
	for _, s := range t.Spans {
		if s.StartTime.Before(rootStart) {
			rootStart = s.StartTime
		}
	}

	// build children map to know depth
	spanIndex := make(map[string]int, len(t.Spans))
	for i, s := range t.Spans {
		spanIndex[s.SpanID] = i
	}

	depth := make(map[string]int, len(t.Spans))
	for _, s := range t.Spans {
		d := 0
		cur := s.ParentSpanID
		for cur != "" {
			if idx, ok := spanIndex[cur]; ok {
				cur = t.Spans[idx].ParentSpanID
				d++
			} else {
				break
			}
		}
		depth[s.SpanID] = d
	}

	for _, s := range t.Spans {
		offset := s.StartTime.Sub(rootStart)
		startPos := int(float64(offset) / float64(total) * float64(flameWidth))
		width := int(float64(s.Duration) / float64(total) * float64(flameWidth))
		if width < 1 {
			width = 1
		}
		if startPos+width > flameWidth {
			width = flameWidth - startPos
		}

		bar := strings.Repeat(" ", startPos) + strings.Repeat("█", width)

		color := colorGreen
		errMark := ""
		if s.HasError {
			color = colorRed
			errMark = " ❌"
		}

		indent := strings.Repeat("  ", depth[s.SpanID])
		label := fmt.Sprintf("%s%s", indent, s.OperationName)

		fmt.Printf("  %-35s  %s%-*s%s  %s%s\n",
			truncate(label, 35),
			color, flameWidth, bar, colorReset,
			formatDuration(s.Duration),
			errMark,
		)
	}
	fmt.Println()
}

func stripANSI(s string) string {
	result := strings.Builder{}
	inEscape := false
	for _, c := range s {
		if c == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if c == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(c)
	}
	return result.String()
}
