package cmd

import (
	"fmt"
	"os"

	"github.com/kfroemming/ralphkit/internal/ui"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "ralphkit",
	Short: "Orchestrate Ralph-style autonomous AI coding loops with Claude Code",
	Long: `ralphkit runs Claude Code in an autonomous loop until a PRD/spec is fully complete.

Each iteration: agent reads spec -> works on tasks -> evaluates completion -> loops until done.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of ralphkit",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("ralphkit %s (commit %s, built %s)\n", Version, Commit, Date)
	},
}

func init() {
	rootCmd.Version = Version
	rootCmd.AddCommand(versionCmd)
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "Suppress UI chrome, show raw output only")
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		q, _ := cmd.Flags().GetBool("quiet")
		ui.Quiet = q
	}
}
