package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/kfroemming/ralphkit/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	worktreeCmd.AddCommand(worktreeAddCmd)
	worktreeCmd.AddCommand(worktreeListCmd)
	worktreeCmd.AddCommand(worktreeRemoveCmd)
	rootCmd.AddCommand(worktreeCmd)
}

var worktreeCmd = &cobra.Command{
	Use:   "worktree",
	Short: "Manage git worktrees",
}

var worktreeAddCmd = &cobra.Command{
	Use:   "add [branch] [path]",
	Short: "Add a git worktree (creates branch if needed)",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		branch := args[0]
		path := branch
		if len(args) > 1 {
			path = args[1]
		}

		// Try to add with existing branch first.
		c := exec.Command("git", "worktree", "add", path, branch)
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			// Branch doesn't exist, create it.
			c = exec.Command("git", "worktree", "add", "-b", branch, path)
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("failed to add worktree: %w", err)
			}
		}
		ui.Success(fmt.Sprintf("Worktree created at %s on branch %s", path, branch))
		return nil
	},
}

var worktreeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List git worktrees",
	RunE: func(cmd *cobra.Command, args []string) error {
		out, err := exec.Command("git", "worktree", "list").Output()
		if err != nil {
			return fmt.Errorf("failed to list worktrees: %w", err)
		}
		lines := strings.Split(strings.TrimSpace(string(out)), "\n")
		for _, line := range lines {
			fmt.Println(line)
		}
		return nil
	},
}

var worktreeRemoveCmd = &cobra.Command{
	Use:   "remove [path]",
	Short: "Remove a git worktree",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := exec.Command("git", "worktree", "remove", args[0])
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("failed to remove worktree: %w", err)
		}
		ui.Success(fmt.Sprintf("Worktree %s removed.", args[0]))
		return nil
	},
}
