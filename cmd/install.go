package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/kfroemming/ralphkit/internal/ui"
	"github.com/spf13/cobra"
)

var checkOnly bool

func init() {
	installCmd.Flags().BoolVar(&checkOnly, "check", false, "Only check dependencies, don't install")
	rootCmd.AddCommand(installCmd)
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Check and install dependencies",
	RunE:  runInstall,
}

func runInstall(cmd *cobra.Command, args []string) error {
	if checkOnly {
		ui.Header("Checking dependencies...")
	} else {
		ui.Header("Installing dependencies...")
	}

	var errors []string

	if runtime.GOOS == "darwin" {
		if err := ensureHomebrew(checkOnly); err != nil {
			errors = append(errors, err.Error())
		}
	}

	if err := ensureNode(checkOnly); err != nil {
		errors = append(errors, err.Error())
	}

	if err := ensureClaude(checkOnly); err != nil {
		errors = append(errors, err.Error())
	}

	if err := checkGit(); err != nil {
		errors = append(errors, err.Error())
	}

	if err := ensureConfigDir(); err != nil {
		return err
	}

	// Final summary
	fmt.Fprintln(os.Stderr)
	if len(errors) > 0 {
		ui.Header("Setup completed with issues:")
		for _, e := range errors {
			ui.Error("  " + e)
		}
		return fmt.Errorf("%d dependency issue(s) remain", len(errors))
	}

	ui.Header("All dependencies ready!")
	return nil
}

func ensureHomebrew(check bool) error {
	if _, err := exec.LookPath("brew"); err == nil {
		out, _ := exec.Command("brew", "--version").Output()
		ui.Success(fmt.Sprintf("✓ %s", trimNL(string(out))))
		return nil
	}

	if check {
		ui.Warn("✗ brew is not installed")
		ui.StatusLine("Install with", `/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`)
		return fmt.Errorf("brew not installed")
	}

	ui.Warn("brew not found, installing Homebrew...")
	cmd := exec.Command("/bin/bash", "-c", `$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)`)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	if err := cmd.Run(); err != nil {
		ui.Error("✗ Failed to install Homebrew")
		return fmt.Errorf("homebrew installation failed: %w", err)
	}
	ui.Success("✓ Homebrew installed")
	return nil
}

func ensureNode(check bool) error {
	if _, err := exec.LookPath("node"); err == nil {
		out, _ := exec.Command("node", "--version").Output()
		ui.Success(fmt.Sprintf("✓ node %s", trimNL(string(out))))
		return nil
	}

	if check {
		ui.Warn("✗ node is not installed")
		ui.StatusLine("Install from", "https://nodejs.org or via your package manager")
		return fmt.Errorf("node not installed")
	}

	if runtime.GOOS == "darwin" {
		if _, err := exec.LookPath("brew"); err == nil {
			ui.Warn("node not found, installing via Homebrew...")
			cmd := exec.Command("brew", "install", "node")
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				ui.Error("✗ Failed to install node via brew")
				return fmt.Errorf("node installation failed: %w", err)
			}
			ui.Success("✓ node installed via Homebrew")
			return nil
		}
	}

	// Linux or no brew available
	ui.Warn("✗ node is not installed")
	ui.StatusLine("Install from", "https://nodejs.org or via your package manager")
	return fmt.Errorf("node not installed (manual installation required on this platform)")
}

func ensureClaude(check bool) error {
	if _, err := exec.LookPath("claude"); err == nil {
		ui.Success("✓ claude CLI installed")
		return nil
	}

	if check {
		ui.Warn("✗ claude CLI is not installed")
		ui.StatusLine("Install with", "npm install -g @anthropic-ai/claude-code")
		return fmt.Errorf("claude CLI not installed")
	}

	ui.Warn("claude CLI not found, installing...")
	cmd := exec.Command("npm", "install", "-g", "@anthropic-ai/claude-code")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		ui.Error("✗ Failed to install claude CLI")
		return fmt.Errorf("claude CLI installation failed: %w", err)
	}
	ui.Success("✓ claude CLI installed")
	return nil
}

func checkGit() error {
	if _, err := exec.LookPath("git"); err != nil {
		ui.Error("✗ git is not installed — please install it manually")
		return fmt.Errorf("git not installed")
	}
	out, _ := exec.Command("git", "--version").Output()
	ui.Success(fmt.Sprintf("✓ %s", trimNL(string(out))))
	return nil
}

func ensureConfigDir() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	configDir := filepath.Join(home, ".ralphkit")
	sessionsDir := filepath.Join(configDir, "sessions")

	if err := os.MkdirAll(sessionsDir, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	ui.Success(fmt.Sprintf("✓ Config directory: %s", configDir))

	configFile := filepath.Join(configDir, "config.yaml")
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		defaultConfig := "# ralphkit configuration\ndefault_model: claude-opus-4-6\nmax_iterations: 10\n"
		if err := os.WriteFile(configFile, []byte(defaultConfig), 0o644); err != nil {
			return fmt.Errorf("failed to create config file: %w", err)
		}
		ui.Success("✓ Created default config.yaml")
	} else {
		ui.Success("✓ Config file exists")
	}

	return nil
}

func trimNL(s string) string {
	if len(s) > 0 && s[len(s)-1] == '\n' {
		return s[:len(s)-1]
	}
	return s
}
