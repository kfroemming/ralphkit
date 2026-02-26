package ui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
	warningStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // yellow
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // red
	headerStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("14")) // cyan bold
	dimStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))             // dim
)

// Quiet suppresses UI chrome when true.
var Quiet bool

func Success(msg string) {
	if Quiet {
		return
	}
	fmt.Fprintln(os.Stderr, successStyle.Render(msg))
}

func Warn(msg string) {
	if Quiet {
		return
	}
	fmt.Fprintln(os.Stderr, warningStyle.Render(msg))
}

func Error(msg string) {
	fmt.Fprintln(os.Stderr, errorStyle.Render(msg))
}

func Header(msg string) {
	if Quiet {
		return
	}
	fmt.Fprintln(os.Stderr, headerStyle.Render(msg))
}

func Dim(msg string) {
	if Quiet {
		return
	}
	fmt.Fprintln(os.Stderr, dimStyle.Render(msg))
}

func IterationHeader(current, max int, elapsed time.Duration) {
	if Quiet {
		return
	}
	line := fmt.Sprintf("=== Iteration %d/%d === [elapsed: %s]", current, max, formatDuration(elapsed))
	fmt.Fprintln(os.Stderr, headerStyle.Render(line))
}

func Celebration(iterations int, elapsed time.Duration) {
	if Quiet {
		fmt.Println("ALL_DONE")
		return
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("10")).
		Padding(1, 3)

	content := fmt.Sprintf(
		"%s\n\n%s\n%s",
		successStyle.Render("Ralph loop complete!"),
		fmt.Sprintf("Iterations: %d", iterations),
		fmt.Sprintf("Total time: %s", formatDuration(elapsed)),
	)
	fmt.Fprintln(os.Stderr, box.Render(content))
}

func MaxIterationsWarning(max int) {
	if Quiet {
		return
	}
	msg := fmt.Sprintf("Reached max iterations (%d). PRD may not be fully complete.\nRe-run with --max-iterations %d to continue.", max, max*2)
	fmt.Fprintln(os.Stderr, warningStyle.Render(msg))
}

func StatusLine(label, value string) {
	if Quiet {
		return
	}
	fmt.Fprintf(os.Stderr, "  %s %s\n", dimStyle.Render(label+":"), value)
}

func PrintLastLines(output string, n int) {
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	start := 0
	if len(lines) > n {
		start = len(lines) - n
	}
	for _, line := range lines[start:] {
		fmt.Println(line)
	}
}

// FormatStatus returns a color-coded status string.
func FormatStatus(status string) string {
	switch status {
	case "running":
		return successStyle.Render("running")
	case "complete":
		return successStyle.Render("complete")
	case "stopped":
		return warningStyle.Render("stopped")
	default:
		return status
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m >= 60 {
		h := m / 60
		m = m % 60
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	return fmt.Sprintf("%dm%ds", m, s)
}
