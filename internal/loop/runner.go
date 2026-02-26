package loop

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/kfroemming/ralphkit/internal/detect"
	"github.com/kfroemming/ralphkit/internal/session"
	"github.com/kfroemming/ralphkit/internal/ui"
)

// Config holds all parameters for a Ralph loop run.
type Config struct {
	PRDContent    string
	Model         string
	MaxIterations int
	SkipTests     bool
	WorkDir       string
	SessionName   string
	DangerouslySkipPermissions bool
	Quiet         bool
}

// completionMarkers are strings that signal the agent considers the PRD complete.
var completionMarkers = []string{"ALL_DONE", "PRD_COMPLETE"}

// Run executes the Ralph loop.
func Run(ctx context.Context, cfg Config) error {
	startTime := time.Now()

	logPath, err := session.LogPath(cfg.SessionName)
	if err != nil {
		return fmt.Errorf("failed to get log path: %w", err)
	}
	logFile, err := os.Create(logPath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer logFile.Close()

	state := &session.State{
		Name:          cfg.SessionName,
		Status:        "running",
		PID:           os.Getpid(),
		Iterations:    0,
		MaxIterations: cfg.MaxIterations,
		Model:         cfg.Model,
		WorkDir:       cfg.WorkDir,
		PRDFile:       "",
		StartTime:     startTime,
		LogFile:       logPath,
	}
	if err := session.Save(state); err != nil {
		return fmt.Errorf("failed to save session state: %w", err)
	}

	var testResults string

	for i := 1; i <= cfg.MaxIterations; i++ {
		select {
		case <-ctx.Done():
			ui.Warn("\nInterrupted. Saving session state...")
			now := time.Now()
			state.Status = "stopped"
			state.EndTime = &now
			_ = session.Save(state)
			return nil
		default:
		}

		state.Iterations = i
		_ = session.Save(state)

		elapsed := time.Since(startTime)
		ui.IterationHeader(i, cfg.MaxIterations, elapsed)

		prompt := buildPrompt(cfg.PRDContent, testResults, i)

		output, err := runClaude(ctx, cfg, prompt, logFile)
		if err != nil {
			if ctx.Err() != nil {
				now := time.Now()
				state.Status = "stopped"
				state.EndTime = &now
				_ = session.Save(state)
				ui.Warn("Session stopped.")
				return nil
			}
			ui.Error(fmt.Sprintf("Claude exited with error: %v", err))
			// Continue to next iteration rather than failing entirely.
		}

		ui.PrintLastLines(output, 10)

		reportPRDProgress(cfg.PRDContent)

		if isComplete(output) {
			now := time.Now()
			state.Status = "complete"
			state.EndTime = &now
			_ = session.Save(state)
			ui.Celebration(i, time.Since(startTime))
			return nil
		}

		// Run tests if enabled.
		if !cfg.SkipTests {
			testResults = runTests(ctx, cfg.WorkDir, logFile)
			if testResults != "" {
				ui.Dim("Test results captured for next iteration.")
			}
		}
	}

	now := time.Now()
	state.Status = "stopped"
	state.EndTime = &now
	_ = session.Save(state)
	ui.MaxIterationsWarning(cfg.MaxIterations)
	return nil
}

// BuildPrompt is the exported version of buildPrompt for use in dry-run mode.
func BuildPrompt(prd, testResults string, iteration int) string {
	return buildPrompt(prd, testResults, iteration)
}

func buildPrompt(prd, testResults string, iteration int) string {
	var b strings.Builder
	b.WriteString("You are working on a coding task. Here is the specification:\n\n")
	b.WriteString(prd)
	b.WriteString("\n\nComplete all items in the specification. When you have completed EVERYTHING, output the exact string ALL_DONE on its own line.")

	if testResults != "" {
		b.WriteString("\n\nIf tests were run, here are the results:\n")
		b.WriteString(testResults)
	}

	b.WriteString(fmt.Sprintf("\n\nCurrent iteration: %d. Items remaining: continue until all done.", iteration))
	return b.String()
}

func runClaude(ctx context.Context, cfg Config, prompt string, logWriter io.Writer) (string, error) {
	args := []string{"-p", prompt, "--model", cfg.Model}
	if cfg.DangerouslySkipPermissions {
		args = append([]string{"--dangerously-skip-permissions"}, args...)
	}

	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = cfg.WorkDir

	var outputBuf bytes.Buffer
	multiOut := io.MultiWriter(&outputBuf, logWriter)
	if !cfg.Quiet {
		multiOut = io.MultiWriter(&outputBuf, logWriter, os.Stdout)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start claude: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Fprintln(multiOut, line)
	}

	err = cmd.Wait()
	return outputBuf.String(), err
}

func runTests(ctx context.Context, workDir string, logWriter io.Writer) string {
	pt := detect.Detect(workDir)
	bin, args := detect.TestCommand(pt)
	if bin == "" {
		return ""
	}

	ui.Dim(fmt.Sprintf("Running tests: %s %s", bin, strings.Join(args, " ")))

	cmd := exec.CommandContext(ctx, bin, args...)
	cmd.Dir = workDir

	var buf bytes.Buffer
	cmd.Stdout = io.MultiWriter(&buf, logWriter)
	cmd.Stderr = io.MultiWriter(&buf, logWriter)

	err := cmd.Run()
	result := buf.String()
	if err != nil {
		result += fmt.Sprintf("\n(tests exited with error: %v)", err)
	}
	return result
}

// reportPRDProgress scans the PRD for markdown checkboxes and prints progress.
func reportPRDProgress(prd string) {
	complete := 0
	incomplete := 0
	for _, line := range strings.Split(prd, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- [x]") || strings.HasPrefix(trimmed, "- [X]") {
			complete++
		} else if strings.HasPrefix(trimmed, "- [ ]") {
			incomplete++
		}
	}
	total := complete + incomplete
	if total == 0 {
		return
	}
	pct := 0
	if total > 0 {
		pct = complete * 100 / total
	}
	ui.Dim(fmt.Sprintf("Progress: %d/%d items complete (%d%%)", complete, total, pct))
}

func isComplete(output string) bool {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		for _, marker := range completionMarkers {
			if trimmed == marker {
				return true
			}
		}
	}
	// Also check for natural language completion signals.
	lower := strings.ToLower(output)
	return strings.Contains(lower, "all items complete") || strings.Contains(lower, "all tasks complete")
}
