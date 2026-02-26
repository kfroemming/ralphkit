package cmd

import (
	"fmt"

	"github.com/kfroemming/ralphkit/internal/session"
	"github.com/kfroemming/ralphkit/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	sessionCmd.AddCommand(sessionListCmd)
	sessionCmd.AddCommand(sessionStopCmd)
	sessionCmd.AddCommand(sessionCleanCmd)
	rootCmd.AddCommand(sessionCmd)
}

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Manage Ralph sessions",
}

var sessionListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sessions",
	RunE: func(cmd *cobra.Command, args []string) error {
		sessions, err := session.List()
		if err != nil {
			return err
		}
		if len(sessions) == 0 {
			ui.Dim("No sessions found.")
			return nil
		}
		for _, s := range sessions {
			status := s.Status
			switch status {
			case "running":
				status = ui.FormatStatus("running")
			case "complete":
				status = ui.FormatStatus("complete")
			case "stopped":
				status = ui.FormatStatus("stopped")
			}
			fmt.Printf("%-20s  %s  iter %d/%d  %s  %s\n",
				s.Name,
				status,
				s.Iterations,
				s.MaxIterations,
				s.StartTime.Format("2006-01-02 15:04"),
				s.WorkDir,
			)
		}
		return nil
	},
}

var sessionStopCmd = &cobra.Command{
	Use:   "stop [name]",
	Short: "Stop a running session",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := session.Stop(args[0]); err != nil {
			return err
		}
		ui.Success(fmt.Sprintf("Session %q stopped.", args[0]))
		return nil
	},
}

var sessionCleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove completed/stopped session files",
	RunE: func(cmd *cobra.Command, args []string) error {
		count, err := session.Clean()
		if err != nil {
			return err
		}
		ui.Success(fmt.Sprintf("Cleaned %d session(s).", count))
		return nil
	},
}
