package cmd

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/kfroemming/ralphkit/internal/session"
	"github.com/kfroemming/ralphkit/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tailCmd)
}

var tailCmd = &cobra.Command{
	Use:   "tail [session-name]",
	Short: "Tail the live output of a running session",
	Args:  cobra.ExactArgs(1),
	RunE:  runTail,
}

func runTail(cmd *cobra.Command, args []string) error {
	name := args[0]
	s, err := session.Load(name)
	if err != nil {
		return err
	}
	if s.LogFile == "" {
		return fmt.Errorf("no log file for session %q", name)
	}

	ui.StatusLine("Tailing session", name)
	ui.StatusLine("Log file", s.LogFile)
	fmt.Println()

	f, err := os.Open(s.LogFile)
	if err != nil {
		return fmt.Errorf("failed to open log: %w", err)
	}
	defer f.Close()

	// Stream existing content.
	if _, err := io.Copy(os.Stdout, f); err != nil {
		return err
	}

	// If session is still running, keep tailing.
	if s.Status != "running" {
		ui.Dim("Session is not running. Showing full log.")
		return nil
	}

	ui.Dim("Streaming live output... (Ctrl+C to stop)")
	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf)
		if n > 0 {
			os.Stdout.Write(buf[:n])
		}
		if err != nil {
			time.Sleep(200 * time.Millisecond)
			// Recheck if session is still running.
			s, _ = session.Load(name)
			if s == nil || s.Status != "running" {
				// Drain remaining.
				io.Copy(os.Stdout, f)
				ui.Dim("\nSession ended.")
				return nil
			}
		}
	}
}
