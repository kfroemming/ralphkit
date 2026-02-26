package prd

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Answers holds the user's responses from the PRD wizard.
type Answers struct {
	ProjectName  string
	Description  string
	TechStack    string
	Features     string
	OutOfScope   string
	SuccessCrit  string
	Constraints  string
}

// Generate calls Claude to expand rough notes into a structured PRD.
func Generate(a Answers) (string, error) {
	notes := formatNotes(a)
	prompt := fmt.Sprintf(
		"You are a product manager. Take these rough notes and turn them into a clear, structured PRD with sections: Overview, Goals, Non-Goals, Features (with acceptance criteria), Technical Approach, Success Metrics. Output ONLY the PRD in markdown.\n\nNotes:\n%s",
		notes,
	)

	cmd := exec.Command("claude", "-p", prompt)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("claude failed: %w\n%s", err, stderr.String())
	}
	return stdout.String(), nil
}

func formatNotes(a Answers) string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Project: %s\n", a.ProjectName))
	b.WriteString(fmt.Sprintf("Description: %s\n", a.Description))
	b.WriteString(fmt.Sprintf("Tech Stack: %s\n", a.TechStack))
	b.WriteString(fmt.Sprintf("Core Features:\n%s\n", a.Features))
	if a.OutOfScope != "" {
		b.WriteString(fmt.Sprintf("Out of Scope:\n%s\n", a.OutOfScope))
	}
	if a.SuccessCrit != "" {
		b.WriteString(fmt.Sprintf("Success Criteria:\n%s\n", a.SuccessCrit))
	}
	if a.Constraints != "" {
		b.WriteString(fmt.Sprintf("Constraints:\n%s\n", a.Constraints))
	}
	return b.String()
}
