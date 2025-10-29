package git

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Worktree represents a git worktree
type Worktree struct {
	Path   string
	Branch string
	Commit string
}

// ListWorktrees returns a list of all worktrees
func ListWorktrees() ([]Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	return parseWorktrees(out.String()), nil
}

// parseWorktrees parses the output of 'git worktree list --porcelain'
func parseWorktrees(output string) []Worktree {
	var worktrees []Worktree
	var current Worktree

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if current.Path != "" {
				worktrees = append(worktrees, current)
				current = Worktree{}
			}
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) < 2 {
			continue
		}

		switch parts[0] {
		case "worktree":
			current.Path = parts[1]
		case "branch":
			current.Branch = strings.TrimPrefix(parts[1], "refs/heads/")
		case "HEAD":
			current.Commit = parts[1]
		}
	}

	// Add the last worktree if exists
	if current.Path != "" {
		worktrees = append(worktrees, current)
	}

	return worktrees
}

// ListBranches returns a list of all branches
func ListBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "-a")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	var branches []string
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Remove '* ' prefix and 'remotes/origin/' prefix
		line = strings.TrimPrefix(line, "* ")
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "remotes/origin/") {
			line = strings.TrimPrefix(line, "remotes/origin/")
		}
		// Skip HEAD pointer
		if !strings.Contains(line, "HEAD") && line != "" {
			branches = append(branches, line)
		}
	}

	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueBranches []string
	for _, branch := range branches {
		if !seen[branch] {
			seen[branch] = true
			uniqueBranches = append(uniqueBranches, branch)
		}
	}

	return uniqueBranches, nil
}

// AddWorktree adds a new worktree
func AddWorktree(path, branch string) error {
	cmd := exec.Command("git", "worktree", "add", path, branch)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add worktree: %s", stderr.String())
	}
	return nil
}

// RemoveWorktree removes a worktree
func RemoveWorktree(path string) error {
	cmd := exec.Command("git", "worktree", "remove", path)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to remove worktree: %s", stderr.String())
	}
	return nil
}
