package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ryuse/challenge-deps/internal/git"
	"github.com/ryuse/challenge-deps/internal/graph"
	"github.com/ryuse/challenge-deps/internal/mermaid"
	"github.com/ryuse/challenge-deps/internal/parser"
	"github.com/ryuse/challenge-deps/pkg/models"
)

const (
	ExitSuccess         = 0
	ExitCircularDep     = 1
	ExitMissingDep      = 2
	ExitInvalidYAML     = 3
	ExitGitError        = 4
	ExitInvalidArgs     = 5
	ExitParseError      = 6
	ExitGraphBuildError = 7
)

type config struct {
	repoPath  string
	base      string
	head      string
	direction string
	format    string
	showHelp  bool
}

func main() {
	cfg := parseFlags()

	if cfg.showHelp {
		flag.Usage()
		os.Exit(ExitSuccess)
	}

	// Check if we're in GitHub Actions
	if cfg.base == "" || cfg.head == "" {
		base, head, isGitHub := git.GetGitHubActionsBranches()
		if isGitHub {
			cfg.base = base
			cfg.head = head
			fmt.Fprintf(os.Stderr, "Detected GitHub Actions environment: base=%s, head=%s\n", base, head)
		}
	}

	// Validate required arguments
	if cfg.repoPath == "" {
		fmt.Fprintf(os.Stderr, "Error: --repo is required\n")
		flag.Usage()
		os.Exit(ExitInvalidArgs)
	}

	// If both base and head are empty, we'll use current branch only
	// If only one is specified, that's an error
	if (cfg.base == "" && cfg.head != "") || (cfg.base != "" && cfg.head == "") {
		fmt.Fprintf(os.Stderr, "Error: --base and --head must both be specified or both omitted\n")
		flag.Usage()
		os.Exit(ExitInvalidArgs)
	}

	if err := run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(getExitCode(err))
	}
}

func parseFlags() config {
	cfg := config{}

	flag.StringVar(&cfg.repoPath, "repo", "", "Path to the challenge repository")
	flag.StringVar(&cfg.base, "base", "", "Base branch (e.g., 'main')")
	flag.StringVar(&cfg.head, "head", "", "Head branch (e.g., 'feature-branch')")
	flag.StringVar(&cfg.direction, "direction", "LR", "Mermaid graph direction (LR, TB, RL, BT)")
	flag.StringVar(&cfg.format, "format", "markdown", "Output format (markdown, mermaid, summary)")
	flag.BoolVar(&cfg.showHelp, "help", false, "Show help message")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: challenge-deps [options]\n\n")
		fmt.Fprintf(os.Stderr, "Generate dependency graphs for CTF challenges.\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  # Analyze current branch only\n")
		fmt.Fprintf(os.Stderr, "  challenge-deps --repo ./challenges\n\n")
		fmt.Fprintf(os.Stderr, "  # Compare two branches\n")
		fmt.Fprintf(os.Stderr, "  challenge-deps --repo ./challenges --base main --head feature-branch\n\n")
		fmt.Fprintf(os.Stderr, "  # Generate text summary\n")
		fmt.Fprintf(os.Stderr, "  challenge-deps --repo . --format summary\n")
		fmt.Fprintf(os.Stderr, "\nNote: --base and --head must both be specified or both omitted.\n")
		fmt.Fprintf(os.Stderr, "In GitHub Actions, GITHUB_BASE_REF and GITHUB_HEAD_REF are used automatically.\n")
	}

	flag.Parse()
	return cfg
}

func run(cfg config) error {
	p := parser.New(cfg.repoPath)
	var challenges []models.ChallengeMetadata
	var err error

	// If base and head are not specified, parse current branch only
	if cfg.base == "" && cfg.head == "" {
		challenges, err = p.ParseAllChallenges()
		if err != nil {
			return &ParseError{msg: err.Error()}
		}
	} else {
		// Parse with Git branch comparison
		gitOps := git.New(cfg.repoPath)

		// Check if it's a git repository
		if !gitOps.IsGitRepository() {
			return &GitError{msg: "not a git repository"}
		}

		// Check if branches exist
		if !gitOps.BranchExists(cfg.base) {
			return &GitError{msg: fmt.Sprintf("base branch '%s' does not exist", cfg.base)}
		}

		if !gitOps.BranchExists(cfg.head) {
			return &GitError{msg: fmt.Sprintf("head branch '%s' does not exist", cfg.head)}
		}

		// Get post-merge files (all files that will exist after merge)
		postMergeFiles, err := gitOps.GetPostMergeFiles(cfg.base, cfg.head)
		if err != nil {
			return &GitError{msg: err.Error()}
		}

		// Filter to only challenge.yml files
		challengeFiles := git.FilterChallengeFiles(postMergeFiles)

		// Get challenges from post-merge state
		var allChallenges []string
		for _, f := range challengeFiles {
			allChallenges = append(allChallenges, f)
		}

		challenges, err = p.ParseChallengesByPaths(allChallenges)
		if err != nil {
			return &ParseError{msg: err.Error()}
		}

		// Determine which challenges are new
		changedChallenges, err := gitOps.GetChangedChallenges(cfg.base, cfg.head)
		if err != nil {
			return &GitError{msg: err.Error()}
		}

		changedSet := make(map[string]bool)
		for _, ch := range changedChallenges {
			changedSet[ch] = true
		}

		// Mark new challenges
		for i := range challenges {
			for changedDir := range changedSet {
				if challenges[i].Challenge.Path == changedDir ||
					challenges[i].Challenge.Path == cfg.repoPath+"/"+changedDir {
					challenges[i].IsNew = true
					break
				}
			}
		}
	}

	// Build dependency graph
	g := graph.New()
	if err := g.Build(challenges); err != nil {
		// Check error type
		if isMissingDependency(err) {
			return &MissingDependencyError{msg: err.Error()}
		}
		if isCircularDependency(err) {
			return &CircularDependencyError{msg: err.Error()}
		}
		return &GraphBuildError{msg: err.Error()}
	}

	// Generate output
	gen := mermaid.New(g)
	gen.SetNewChallenges(challenges)

	var output string
	switch cfg.format {
	case "markdown":
		output = gen.GenerateMarkdown(cfg.direction)
	case "mermaid":
		output = gen.Generate(cfg.direction)
	case "summary":
		output = gen.GenerateSummary()
	default:
		output = gen.GenerateMarkdown(cfg.direction)
	}

	fmt.Print(output)
	return nil
}

// Error types
type CircularDependencyError struct {
	msg string
}

func (e *CircularDependencyError) Error() string {
	return e.msg
}

type MissingDependencyError struct {
	msg string
}

func (e *MissingDependencyError) Error() string {
	return e.msg
}

type ParseError struct {
	msg string
}

func (e *ParseError) Error() string {
	return e.msg
}

type GitError struct {
	msg string
}

func (e *GitError) Error() string {
	return e.msg
}

type GraphBuildError struct {
	msg string
}

func (e *GraphBuildError) Error() string {
	return e.msg
}

func getExitCode(err error) int {
	switch err.(type) {
	case *CircularDependencyError:
		return ExitCircularDep
	case *MissingDependencyError:
		return ExitMissingDep
	case *ParseError:
		return ExitInvalidYAML
	case *GitError:
		return ExitGitError
	case *GraphBuildError:
		return ExitGraphBuildError
	default:
		return ExitInvalidArgs
	}
}

func isCircularDependency(err error) bool {
	if err == nil {
		return false
	}
	return contains(err.Error(), "circular")
}

func isMissingDependency(err error) bool {
	if err == nil {
		return false
	}
	return contains(err.Error(), "does not exist")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
