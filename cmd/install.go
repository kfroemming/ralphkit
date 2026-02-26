package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/kfroemming/ralphkit/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Check and install dependencies",
	RunE:  runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	ui.Header("Checking dependencies...")

	// Check Node/npm
	if _, err := exec.LookPath("node"); err != nil {
		ui.Error("node is not installed")
		ui.Warn("Install Node.js from https://nodejs.org or via your package manager")
	} else {
		out, _ := exec.Command("node", "--version").Output()
		ui.Success(fmt.Sprintf("node %s", trimNL(string(out))))
	}

	if _, err := exec.LookPath("npm"); err != nil {
		ui.Error("npm is not installed")
	} else {
		out, _ := exec.Command("npm", "--version").Output()
		ui.Success(fmt.Sprintf("npm %s", trimNL(string(out))))
	}

	// Check Claude CLI
	if _, err := exec.LookPath("claude"); err != nil {
		ui.Warn("claude CLI is not installed")
		ui.StatusLine("Install with", "npm install -g @anthropic-ai/claude-code")
	} else {
		ui.Success("claude CLI is installed")
	}

	// Check git
	if _, err := exec.LookPath("git"); err != nil {
		ui.Error("git is not installed")
	} else {
		out, _ := exec.Command("git", "--version").Output()
		ui.Success(trimNL(string(out)))
	}

	// Create config dir
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".ralphkit")
	sessionsDir := filepath.Join(configDir, "sessions")

	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	ui.Success(fmt.Sprintf("Config directory: %s", configDir))

	// Create default config if missing
	configFile := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := "# ralphkit configuration\ndefault_model: claude-opus-4-6\nmax_iterations: 10\n"
		if err := os.WriteFile(configFile, []byte(defaultConfig), 0o644); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
		ui.Success("Created default config.yaml")
	} else {
		ui.Success("Config file exists")
	}

	ui.Header("Setup complete!")
	return nil
}

func trimNL(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}
