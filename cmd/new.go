package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/huh"
	"github.com/kfroemming/ralphkit/internal/prd"
	"github.com/kfroemming/ralphkit/internal/ui"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newCmd)
}

var newCmd = &cobra.Command{
	Use:   "new [name]",
	Short: "Interactive PRD crafting wizard",
	Long:  "Launch an interactive wizard to build a spec, then optionally run a Ralph loop on it.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runNew,
}

func runNew(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) > 0 {
		name = args[0]
	}

	var (
		description string
		techStack   string
		features    string
		outOfScope  string
		successCrit string
		constraints string
	)

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project name").
				Value(&name).
				Placeholder("my-project"),
			huh.NewInput().
				Title("What are you building? (1-2 sentences)").
				Value(&description).
				Placeholder("A CLI tool that..."),
			huh.NewInput().
				Title("Tech stack").
				Value(&techStack).
				Placeholder("Go, Python, Node, etc."),
		),
		huh.NewGroup(
			huh.NewText().
				Title("Core features (bullet points)").
				Value(&features).
				Placeholder("- Feature one\n- Feature two"),
			huh.NewText().
				Title("Out of scope / exclusions").
				Value(&outOfScope).
				Placeholder("- Not building X\n- Skipping Y"),
		),
		huh.NewGroup(
			huh.NewText().
				Title("Success criteria — how will you know it's done?").
				Value(&successCrit).
				Placeholder("- All tests pass\n- CLI runs end-to-end"),
			huh.NewText().
				Title("Constraints (performance, compatibility, style, etc.)").
				Value(&constraints).
				Placeholder("- Must work on macOS and Linux"),
		),
	)

	if err := form.Run(); err != nil {
		return err
	}

	if name == "" {
		name = "project"
	}

	answers := prd.Answers{
		ProjectName: name,
		Description: description,
		TechStack:   techStack,
		Features:    features,
		OutOfScope:  outOfScope,
		SuccessCrit: successCrit,
		Constraints: constraints,
	}

	ui.Header("Generating PRD with Claude...")

	generated, err := prd.Generate(answers)
	if err != nil {
		return fmt.Errorf("failed to generate PRD: %w", err)
	}

	for {
		fmt.Println()
		fmt.Println(generated)
		fmt.Println()

		reader := bufio.NewReader(os.Stdin)
		fmt.Print("[S]ave / [E]dit / [R]egenerate / [C]ancel: ")
		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(strings.ToLower(choice))

		switch choice {
		case "s", "save":
			filename := name + ".prd.md"
			if err := os.WriteFile(filename, []byte(generated), 0o644); err != nil {
				return fmt.Errorf("failed to save PRD: %w", err)
			}
			ui.Success(fmt.Sprintf("Saved to %s", filename))
			fmt.Print("Run Ralph loop now? [y/N]: ")
			yn, _ := reader.ReadString('\n')
			if strings.TrimSpace(strings.ToLower(yn)) == "y" {
				ui.Header(fmt.Sprintf("Run: ralphkit run %s", filename))
			}
			return nil

		case "e", "edit":
			filename := name + ".prd.md"
			if err := os.WriteFile(filename, []byte(generated), 0o644); err != nil {
				return fmt.Errorf("failed to write temp file: %w", err)
			}
			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "nano"
			}
			c := exec.Command(editor, filename)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			if err := c.Run(); err != nil {
				return fmt.Errorf("editor failed: %w", err)
			}
			data, err := os.ReadFile(filename)
			if err != nil {
				return fmt.Errorf("failed to read edited file: %w", err)
			}
			generated = string(data)
			ui.Success("PRD updated from editor.")

		case "r", "regenerate":
			ui.Header("Regenerating PRD...")
			generated, err = prd.Generate(answers)
			if err != nil {
				return fmt.Errorf("failed to regenerate PRD: %w", err)
			}

		case "c", "cancel":
			ui.Warn("Cancelled.")
			return nil

		default:
			ui.Warn("Invalid choice. Use S, E, R, or C.")
		}
	}
}
