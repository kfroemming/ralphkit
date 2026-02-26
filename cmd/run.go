package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/kfroemming/ralphkit/internal/loop"
	"github.com/kfroemming/ralphkit/internal/ui"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func init() {
	runCmd.Flags().StringP("model", "m", "", "Claude model (default from config, shortcuts: opus, sonnet, haiku)")
	runCmd.Flags().Bool("skip-tests", false, "Skip running tests between iterations")
	runCmd.Flags().Bool("with-tests", true, "Run tests between iterations (default)")
	runCmd.Flags().IntP("max-iterations", "n", 0, "Max loop iterations (default from config)")
	runCmd.Flags().StringP("worktree", "w", "", "Git worktree path to run in")
	runCmd.Flags().StringP("session-name", "s", "", "Name this session (auto-generated if not provided)")
	runCmd.Flags().StringP("dir", "d", "", "Working directory (default: current dir)")
	runCmd.Flags().Bool("notify", false, "Send macOS notification on completion")
	runCmd.Flags().Bool("dangerously-skip-permissions", false, "Pass --dangerously-skip-permissions to claude")
	runCmd.Flags().Bool("dry-run", false, "Print resolved config and prompt without running Claude")
	rootCmd.AddCommand(runCmd)
}

var runCmd = &cobra.Command{
	Use:   "run [prd-file]",
	Short: "Start a Ralph loop from a PRD/spec file",
	Args:  cobra.ExactArgs(1),
	RunE:  runRun,
}

func runRun(cmd *cobra.Command, args []string) error {
	prdFile := args[0]
	data, err := os.ReadFile(prdFile)
	if err != nil {
		return fmt.Errorf("failed to read PRD file: %w", err)
	}

	model, _ := cmd.Flags().GetString("model")
	if model == "" {
		model = viper.GetString("default_model")
	}
	if model == "" {
		model = "claude-opus-4-6"
	}
	model = resolveModel(model)

	maxIter, _ := cmd.Flags().GetInt("max-iterations")
	if maxIter == 0 {
		maxIter = viper.GetInt("max_iterations")
	}
	if maxIter == 0 {
		maxIter = 10
	}

	skipTests, _ := cmd.Flags().GetBool("skip-tests")
	dangerouslySkip, _ := cmd.Flags().GetBool("dangerously-skip-permissions")
	notify, _ := cmd.Flags().GetBool("notify")

	workDir, _ := cmd.Flags().GetString("dir")
	if workDir == "" {
		worktree, _ := cmd.Flags().GetString("worktree")
		if worktree != "" {
			workDir = worktree
		} else {
			workDir, _ = os.Getwd()
		}
	}
	workDir, _ = filepath.Abs(workDir)

	sessionName, _ := cmd.Flags().GetString("session-name")
	if sessionName == "" {
		base := strings.TrimSuffix(filepath.Base(prdFile), filepath.Ext(prdFile))
		sessionName = fmt.Sprintf("%s-%d", base, time.Now().Unix())
	}

	quiet, _ := cmd.Flags().GetBool("quiet")

	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		fmt.Println("--- DRY RUN MODE ---")
		fmt.Println()
		fmt.Println("Resolved config:")
		ui.StatusLine("PRD", prdFile)
		ui.StatusLine("Model", model)
		ui.StatusLine("Max iterations", fmt.Sprintf("%d", maxIter))
		ui.StatusLine("Work dir", workDir)
		ui.StatusLine("Session", sessionName)
		testCmd := ""
		if !skipTests {
			testCmd = "(auto-detect)"
		} else {
			testCmd = "(skipped)"
		}
		ui.StatusLine("Test command", testCmd)
		fmt.Println()
		fmt.Println("Prompt that would be sent to Claude (iteration 1):")
		fmt.Println("---")
		fmt.Println(loop.BuildPrompt(string(data), "", 1))
		fmt.Println("---")
		fmt.Println()
		fmt.Println("(dry-run complete — no Claude invocation performed)")
		return nil
	}

	ui.Header("Ralph Loop")
	ui.StatusLine("PRD", prdFile)
	ui.StatusLine("Model", model)
	ui.StatusLine("Max iterations", fmt.Sprintf("%d", maxIter))
	ui.StatusLine("Work dir", workDir)
	ui.StatusLine("Session", sessionName)
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		cancel()
	}()

	cfg := loop.Config{
		PRDContent:                 string(data),
		Model:                     model,
		MaxIterations:             maxIter,
		SkipTests:                 skipTests,
		WorkDir:                   workDir,
		SessionName:               sessionName,
		DangerouslySkipPermissions: dangerouslySkip,
		Quiet:                     quiet,
	}

	err = loop.Run(ctx, cfg)
	if notify {
		sendNotification(err)
	}
	return err
}

func resolveModel(m string) string {
	switch strings.ToLower(m) {
	case "opus":
		return "claude-opus-4-6"
	case "sonnet":
		return "claude-sonnet-4-6"
	case "haiku":
		return "claude-haiku-4-5"
	default:
		return m
	}
}

func sendNotification(err error) {
	msg := "Ralph loop complete!"
	if err != nil {
		msg = fmt.Sprintf("Ralph loop failed: %v", err)
	}
	script := fmt.Sprintf(`display notification "%s" with title "ralphkit"`, msg)
	exec.Command("osascript", "-e", script).Run()
}
