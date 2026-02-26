package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kfroemming/ralphkit/internal/detect"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check the local environment for ralphkit dependencies",
	RunE:  runDoctor,
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Printf("ralphkit doctor — version %s\n\n", Version)

	criticalMissing := false

	// Critical dependencies
	gitOk := commandExists("git")
	printCheck("git", gitOk, true)
	if !gitOk {
		criticalMissing = true
	}

	claudeOk := commandExists("claude")
	printCheck("claude CLI", claudeOk, true)
	if !claudeOk {
		criticalMissing = true
	}

	// Optional dependencies
	printCheck("node", commandExists("node"), false)
	printCheck("npm", commandExists("npm"), false)

	// Environment
	apiKeyOk := os.Getenv("ANTHROPIC_API_KEY") != ""
	printCheck("ANTHROPIC_API_KEY", apiKeyOk, false)

	// Config dir
	home, _ := os.UserHomeDir()
	configDir := filepath.Join(home, ".ralphkit")
	configDirOk := dirExists(configDir)
	printCheck("~/.ralphkit/ config dir", configDirOk, false)

	// Project type detection
	cwd, _ := os.Getwd()
	pt := detect.Detect(cwd)
	fmt.Printf("\nProject type:    %s\n", pt)
	fmt.Printf("ralphkit version: %s\n", Version)

	fmt.Println()
	if criticalMissing {
		fmt.Println("❌ Critical dependencies missing — ralphkit will not function correctly.")
		return fmt.Errorf("critical dependencies missing (git and/or claude CLI)")
	}
	fmt.Println("✅ All critical checks passed.")
	return nil
}

func printCheck(name string, ok bool, critical bool) {
	label := name
	if critical {
		label = name + " (critical)"
	}
	if ok {
		fmt.Printf("  ✅  %-35s ok\n", label)
	} else {
		fmt.Printf("  ❌  %-35s not found\n", label)
	}
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
