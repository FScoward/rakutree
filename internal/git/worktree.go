package git

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

	// Sort branches: main/master first, then alphabetically
	return sortBranches(uniqueBranches), nil
}

// sortBranches sorts branches with main/master at the top
func sortBranches(branches []string) []string {
	var priorityBranches []string
	var otherBranches []string

	for _, branch := range branches {
		if branch == "main" || branch == "master" {
			priorityBranches = append(priorityBranches, branch)
		} else {
			otherBranches = append(otherBranches, branch)
		}
	}

	// Sort priority branches (main before master)
	for i := 0; i < len(priorityBranches)-1; i++ {
		for j := i + 1; j < len(priorityBranches); j++ {
			if priorityBranches[i] == "master" && priorityBranches[j] == "main" {
				priorityBranches[i], priorityBranches[j] = priorityBranches[j], priorityBranches[i]
			}
		}
	}

	// Sort other branches alphabetically
	for i := 0; i < len(otherBranches)-1; i++ {
		for j := i + 1; j < len(otherBranches); j++ {
			if otherBranches[i] > otherBranches[j] {
				otherBranches[i], otherBranches[j] = otherBranches[j], otherBranches[i]
			}
		}
	}

	// Combine: priority branches first, then others
	result := append(priorityBranches, otherBranches...)
	return result
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

// AddWorktreeWithNewBranch creates a new branch and adds a worktree for it
func AddWorktreeWithNewBranch(path, newBranch, baseBranch string) error {
	cmd := exec.Command("git", "worktree", "add", "-b", newBranch, path, baseBranch)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add worktree with new branch: %s", stderr.String())
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

// PathSuggestion represents a suggested path with description
type PathSuggestion struct {
	Path        string
	Description string
	IsCustom    bool
}

// SuggestPaths generates path suggestions based on existing worktrees and the new branch
func SuggestPaths(branch string) ([]PathSuggestion, error) {
	worktrees, err := ListWorktrees()
	if err != nil {
		return nil, err
	}

	var suggestions []PathSuggestion
	seen := make(map[string]bool)

	// Skip the main worktree (first one) for pattern analysis
	if len(worktrees) > 1 {
		patterns := analyzePathPatterns(worktrees[1:])

		// Generate suggestions from learned patterns
		for _, pattern := range patterns {
			path := applyPattern(pattern, branch)
			if path != "" && !seen[path] {
				seen[path] = true
				suggestions = append(suggestions, PathSuggestion{
					Path:        path,
					Description: fmt.Sprintf("Learned pattern (%d similar)", pattern.Count),
					IsCustom:    false,
				})
			}
		}
	}

	// Add default patterns if we don't have many suggestions
	if len(suggestions) < 3 {
		defaultSuggestions := getDefaultSuggestions(branch)
		for _, sug := range defaultSuggestions {
			if !seen[sug.Path] {
				seen[sug.Path] = true
				suggestions = append(suggestions, sug)
			}
		}
	}

	// Add custom input option at the end
	suggestions = append(suggestions, PathSuggestion{
		Path:        "",
		Description: "Enter custom path...",
		IsCustom:    true,
	})

	return suggestions, nil
}

// pathPattern represents a detected path pattern
type pathPattern struct {
	Template string // e.g., "../{branch}", "../worktrees/{branch}"
	Count    int    // How many times this pattern appears
}

// analyzePathPatterns analyzes existing worktree paths to detect patterns
func analyzePathPatterns(worktrees []Worktree) []pathPattern {
	patternMap := make(map[string]int)

	for _, wt := range worktrees {
		if wt.Branch == "" {
			continue
		}

		// Try to extract pattern by replacing branch name
		pattern := extractPattern(wt.Path, wt.Branch)
		if pattern != "" {
			patternMap[pattern]++
		}
	}

	// Convert map to sorted slice
	var patterns []pathPattern
	for template, count := range patternMap {
		patterns = append(patterns, pathPattern{
			Template: template,
			Count:    count,
		})
	}

	// Sort by count (most used first)
	for i := 0; i < len(patterns)-1; i++ {
		for j := i + 1; j < len(patterns); j++ {
			if patterns[j].Count > patterns[i].Count {
				patterns[i], patterns[j] = patterns[j], patterns[i]
			}
		}
	}

	return patterns
}

// extractPattern tries to extract a path pattern by replacing branch name with placeholder
func extractPattern(path, branch string) string {
	// Normalize branch name (replace slashes with dashes for comparison)
	normalizedBranch := strings.ReplaceAll(branch, "/", "-")

	// Try different variations
	variations := []string{
		branch,
		normalizedBranch,
		strings.ToLower(branch),
		strings.ToLower(normalizedBranch),
	}

	for _, variant := range variations {
		if strings.Contains(path, variant) {
			pattern := strings.ReplaceAll(path, variant, "{branch}")
			return pattern
		}
	}

	return ""
}

// applyPattern applies a pattern template to a new branch name
func applyPattern(pattern pathPattern, branch string) string {
	// Normalize branch name for path
	normalizedBranch := strings.ReplaceAll(branch, "/", "-")

	path := strings.ReplaceAll(pattern.Template, "{branch}", normalizedBranch)
	return path
}

// getDefaultSuggestions returns default path suggestions when no patterns are learned
func getDefaultSuggestions(branch string) []PathSuggestion {
	normalizedBranch := strings.ReplaceAll(branch, "/", "-")

	// Get repository name for some suggestions
	repoName := getRepoName()

	suggestions := []PathSuggestion{
		{
			Path:        fmt.Sprintf("../%s", normalizedBranch),
			Description: "Sibling directory (default)",
			IsCustom:    false,
		},
		{
			Path:        fmt.Sprintf("../worktrees/%s", normalizedBranch),
			Description: "Organized in worktrees folder",
			IsCustom:    false,
		},
	}

	if repoName != "" {
		suggestions = append(suggestions, PathSuggestion{
			Path:        fmt.Sprintf("../%s-%s", repoName, normalizedBranch),
			Description: "With repository name prefix",
			IsCustom:    false,
		})
	}

	return suggestions
}

// getRepoName tries to get the current repository name
func getRepoName() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Base(cwd)
}

// BranchNameSuggestion represents a suggested branch name
type BranchNameSuggestion struct {
	Name        string
	Description string
	IsCustom    bool
}

// SuggestBranchNames generates branch name suggestions based on existing branches
func SuggestBranchNames() ([]BranchNameSuggestion, error) {
	branches, err := ListBranches()
	if err != nil {
		return nil, err
	}

	var suggestions []BranchNameSuggestion
	seen := make(map[string]bool)

	// Analyze existing branches for patterns
	prefixCounts := analyzeBranchPrefixes(branches)

	// Add learned prefix patterns
	for prefix, count := range prefixCounts {
		if count >= 2 && !seen[prefix] { // Only suggest if used at least twice
			seen[prefix] = true
			suggestions = append(suggestions, BranchNameSuggestion{
				Name:        prefix,
				Description: fmt.Sprintf("Learned pattern (%d branches)", count),
				IsCustom:    false,
			})
		}
	}

	// Add common prefixes
	commonPrefixes := []struct {
		prefix string
		desc   string
	}{
		{"feature/", "New feature"},
		{"bugfix/", "Bug fix"},
		{"hotfix/", "Urgent fix"},
		{"release/", "Release branch"},
		{"refactor/", "Code refactoring"},
		{"chore/", "Maintenance task"},
	}

	for _, cp := range commonPrefixes {
		if !seen[cp.prefix] {
			seen[cp.prefix] = true
			suggestions = append(suggestions, BranchNameSuggestion{
				Name:        cp.prefix,
				Description: cp.desc,
				IsCustom:    false,
			})
		}
	}

	// Add custom input option
	suggestions = append(suggestions, BranchNameSuggestion{
		Name:        "",
		Description: "Enter custom branch name...",
		IsCustom:    true,
	})

	return suggestions, nil
}

// analyzeBranchPrefixes analyzes branch names to find common prefixes
func analyzeBranchPrefixes(branches []string) map[string]int {
	prefixCounts := make(map[string]int)

	for _, branch := range branches {
		// Skip main/master branches
		if branch == "main" || branch == "master" {
			continue
		}

		// Look for patterns like "feature/", "bugfix/", etc.
		if idx := strings.Index(branch, "/"); idx > 0 {
			prefix := branch[:idx+1]
			prefixCounts[prefix]++
		}
	}

	return prefixCounts
}
