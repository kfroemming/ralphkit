package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"syscall"
	"time"
)

// State represents a saved session.
type State struct {
	Name          string     `json:"name"`
	Status        string     `json:"status"` // running, stopped, complete
	PID           int        `json:"pid"`
	Iterations    int        `json:"iterations"`
	MaxIterations int        `json:"maxIterations"`
	Model         string     `json:"model"`
	WorkDir       string     `json:"workDir"`
	PRDFile       string     `json:"prdFile"`
	StartTime     time.Time  `json:"startTime"`
	EndTime       *time.Time `json:"endTime"`
	LogFile       string     `json:"logFile"`
}

// Dir returns the sessions directory, creating it if needed.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".ralphkit", "sessions")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

// Save persists a session state to disk.
func Save(s *State) error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, s.Name+".json"), data, 0o644)
}

// Load reads a session state from disk.
func Load(name string) (*State, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(filepath.Join(dir, name+".json"))
	if err != nil {
		return nil, fmt.Errorf("session %q not found", name)
	}
	var s State
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

// List returns all saved sessions.
func List() ([]*State, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var sessions []*State
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		name := e.Name()[:len(e.Name())-5]
		s, err := Load(name)
		if err != nil {
			continue
		}
		// Check if "running" sessions are actually still alive.
		if s.Status == "running" && !processAlive(s.PID) {
			s.Status = "stopped"
			_ = Save(s)
		}
		sessions = append(sessions, s)
	}
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].StartTime.After(sessions[j].StartTime)
	})
	return sessions, nil
}

// Stop sends SIGINT to a running session's process.
func Stop(name string) error {
	s, err := Load(name)
	if err != nil {
		return err
	}
	if s.Status != "running" {
		return fmt.Errorf("session %q is not running (status: %s)", name, s.Status)
	}
	if s.PID > 0 {
		proc, err := os.FindProcess(s.PID)
		if err == nil {
			_ = proc.Signal(syscall.SIGINT)
		}
	}
	now := time.Now()
	s.Status = "stopped"
	s.EndTime = &now
	return Save(s)
}

// Clean removes all completed or stopped session files.
func Clean() (int, error) {
	dir, err := Dir()
	if err != nil {
		return 0, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		name := e.Name()[:len(e.Name())-5]
		s, err := Load(name)
		if err != nil {
			continue
		}
		if s.Status == "complete" || s.Status == "stopped" {
			os.Remove(filepath.Join(dir, e.Name()))
			if s.LogFile != "" {
				os.Remove(s.LogFile)
			}
			count++
		}
	}
	return count, nil
}

// LogPath returns the log file path for a session.
func LogPath(name string) (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, name+".log"), nil
}

func processAlive(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}
