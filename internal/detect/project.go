package detect

import (
	"os"
	"path/filepath"
)

type ProjectType string

const (
	ProjectGo      ProjectType = "go"
	ProjectNode    ProjectType = "node"
	ProjectPython  ProjectType = "python"
	ProjectUnknown ProjectType = "unknown"
)

// Detect returns the project type based on files in the given directory.
func Detect(dir string) ProjectType {
	if fileExists(filepath.Join(dir, "go.mod")) {
		return ProjectGo
	}
	if fileExists(filepath.Join(dir, "package.json")) {
		return ProjectNode
	}
	if fileExists(filepath.Join(dir, "pyproject.toml")) || fileExists(filepath.Join(dir, "setup.py")) || fileExists(filepath.Join(dir, "requirements.txt")) {
		return ProjectPython
	}
	return ProjectUnknown
}

// TestCommand returns the test command for the given project type.
func TestCommand(pt ProjectType) (string, []string) {
	switch pt {
	case ProjectGo:
		return "go", []string{"test", "./..."}
	case ProjectNode:
		return "npm", []string{"test"}
	case ProjectPython:
		return "python", []string{"-m", "pytest"}
	default:
		return "", nil
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
