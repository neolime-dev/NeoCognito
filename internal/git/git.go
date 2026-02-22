// Package git provides lightweight git integration for auto-committing blocks.
package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// IsRepo returns true if the given directory is inside a git repository.
func IsRepo(dir string) bool {
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// Init initialises a git repository in dir if one doesn't already exist.
// It also writes a .gitignore that excludes the SQLite WAL files.
func Init(dir string) error {
	if IsRepo(dir) {
		return nil
	}
	if out, err := run(dir, "git", "init"); err != nil {
		return fmt.Errorf("git init: %s: %w", out, err)
	}

	// Write a sensible .gitignore
	gitignore := filepath.Join(dir, ".gitignore")
	if _, err := os.Stat(gitignore); os.IsNotExist(err) {
		content := "# SQLite WAL/SHM files (index is rebuildable)\nindex.db-wal\nindex.db-shm\n"
		if err := os.WriteFile(gitignore, []byte(content), 0644); err != nil {
			return fmt.Errorf("writing .gitignore: %w", err)
		}
	}

	// Initial commit so the repo is valid
	_, _ = run(dir, "git", "add", ".gitignore")
	_, _ = run(dir, "git", "commit", "-m", "chore: init neocognito vault")
	return nil
}

// Commit stages all changes in dir and creates a commit with the given message.
// It is a no-op if there is nothing to commit.
func Commit(dir, message string) error {
	// Stage everything (blocks, assets; DB WAL is gitignored)
	if out, err := run(dir, "git", "add", "-A"); err != nil {
		return fmt.Errorf("git add: %s: %w", out, err)
	}

	// Check if there's anything staged
	out, _ := run(dir, "git", "status", "--porcelain")
	if out == "" {
		return nil // nothing to commit
	}

	if out, err := run(dir, "git", "commit", "-m", message); err != nil {
		return fmt.Errorf("git commit: %s: %w", out, err)
	}
	return nil
}

// Log returns the last n commit messages for a specific file.
func Log(dir, filePath string, n int) ([]string, error) {
	args := []string{
		"-C", dir,
		"log",
		fmt.Sprintf("--max-count=%d", n),
		"--pretty=format:%h %ai %s",
		"--",
		filePath,
	}
	out, err := run(dir, "git", args...)
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}
	var lines []string
	start, i := 0, 0
	for i < len(out) {
		if out[i] == '\n' {
			if i > start {
				lines = append(lines, out[start:i])
			}
			start = i + 1
		}
		i++
	}
	if start < len(out) {
		lines = append(lines, out[start:])
	}
	return lines, nil
}

func run(dir string, name string, args ...string) (string, error) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	return string(out), err
}
