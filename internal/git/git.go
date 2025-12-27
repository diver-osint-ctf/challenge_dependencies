package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Git provides git operations for comparing branches
type Git struct {
	repoPath string
}

// New creates a new Git instance for the given repository path
func New(repoPath string) *Git {
	return &Git{
		repoPath: repoPath,
	}
}

// GetChangedFiles returns the list of files that differ between base and head branches
func (g *Git) GetChangedFiles(base, head string) ([]string, error) {
	// First, try to use git diff to get changed files
	cmd := exec.Command("git", "diff", "--name-only", base+"..."+head)
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		// If git diff fails, fallback to ls-tree comparison
		return g.compareTreesViaLsTree(base, head)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}

	return files, nil
}

// GetFilesInBranch returns all files in a specific branch
func (g *Git) GetFilesInBranch(branch string) ([]string, error) {
	cmd := exec.Command("git", "ls-tree", "-r", "--name-only", branch)
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list files in branch %s: %w", branch, err)
	}

	files := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(files) == 1 && files[0] == "" {
		return []string{}, nil
	}

	return files, nil
}

// compareTreesViaLsTree compares files between branches using ls-tree
func (g *Git) compareTreesViaLsTree(base, head string) ([]string, error) {
	baseFiles, err := g.GetFilesInBranch(base)
	if err != nil {
		return nil, err
	}

	headFiles, err := g.GetFilesInBranch(head)
	if err != nil {
		return nil, err
	}

	// Create a map of base files for quick lookup
	baseFileSet := make(map[string]bool)
	for _, f := range baseFiles {
		baseFileSet[f] = true
	}

	// Find files that are in head but not in base (new or modified)
	var changedFiles []string
	for _, f := range headFiles {
		if !baseFileSet[f] {
			changedFiles = append(changedFiles, f)
		}
	}

	return changedFiles, nil
}

// GetPostMergeFiles returns files that will exist after merging head into base
// This includes all files from base plus new/modified files from head
func (g *Git) GetPostMergeFiles(base, head string) ([]string, error) {
	baseFiles, err := g.GetFilesInBranch(base)
	if err != nil {
		return nil, err
	}

	headFiles, err := g.GetFilesInBranch(head)
	if err != nil {
		return nil, err
	}

	// Use a map to deduplicate files
	fileSet := make(map[string]bool)

	// Add all files from base
	for _, f := range baseFiles {
		fileSet[f] = true
	}

	// Add all files from head (will override base files if they exist)
	for _, f := range headFiles {
		fileSet[f] = true
	}

	// Convert map back to slice
	var allFiles []string
	for f := range fileSet {
		allFiles = append(allFiles, f)
	}

	return allFiles, nil
}

// FilterChallengeFiles filters a list of files to only include challenge.yml files
func FilterChallengeFiles(files []string) []string {
	var challengeFiles []string
	for _, f := range files {
		if filepath.Base(f) == "challenge.yml" {
			challengeFiles = append(challengeFiles, f)
		}
	}
	return challengeFiles
}

// GetChangedChallenges returns challenge directories that have changed between branches
func (g *Git) GetChangedChallenges(base, head string) ([]string, error) {
	changedFiles, err := g.GetChangedFiles(base, head)
	if err != nil {
		return nil, err
	}

	// Filter to only challenge.yml files
	challengeFiles := FilterChallengeFiles(changedFiles)

	// Extract directory paths
	var challengeDirs []string
	for _, f := range challengeFiles {
		dir := filepath.Dir(f)
		challengeDirs = append(challengeDirs, dir)
	}

	return challengeDirs, nil
}

// IsGitRepository checks if the given path is a git repository
func (g *Git) IsGitRepository() bool {
	gitDir := filepath.Join(g.repoPath, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// GetCurrentBranch returns the name of the current branch
func (g *Git) GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = g.repoPath

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// BranchExists checks if a branch exists
func (g *Git) BranchExists(branch string) bool {
	cmd := exec.Command("git", "rev-parse", "--verify", branch)
	cmd.Dir = g.repoPath

	err := cmd.Run()
	return err == nil
}

// GetEnvOrDefault returns an environment variable value or a default if not set
func GetEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetGitHubActionsBranches returns the base and head branches from GitHub Actions environment
func GetGitHubActionsBranches() (base, head string, isGitHubActions bool) {
	baseRef := os.Getenv("GITHUB_BASE_REF")
	headRef := os.Getenv("GITHUB_HEAD_REF")

	if baseRef != "" && headRef != "" {
		return baseRef, headRef, true
	}

	return "", "", false
}
