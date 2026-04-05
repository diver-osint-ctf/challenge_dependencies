package main

import (
	"errors"
	"testing"

	"github.com/ryuse/challenge-deps/internal/graph"
)

func TestGetExitCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{"circular dependency", &CircularDependencyError{msg: "cycle"}, ExitCircularDep},
		{"missing dependency", &MissingDependencyError{msg: "missing"}, ExitMissingDep},
		{"parse error", &ParseError{msg: "bad yaml"}, ExitInvalidYAML},
		{"git error", &GitError{msg: "not a repo"}, ExitGitError},
		{"graph build error", &GraphBuildError{msg: "build failed"}, ExitGraphBuildError},
		{"unknown error", errors.New("unknown"), ExitInvalidArgs},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getExitCode(tt.err)
			if got != tt.want {
				t.Errorf("getExitCode() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestRunWithTestdata(t *testing.T) {
	cfg := config{
		repoPath:  "../../testdata/sample-repo",
		direction: "LR",
		format:    "mermaid",
	}

	err := run(cfg)
	if err != nil {
		t.Errorf("run() with testdata should succeed, got: %v", err)
	}
}

func TestRunFormats(t *testing.T) {
	formats := []string{"markdown", "mermaid", "summary", "unknown-defaults-to-markdown"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			cfg := config{
				repoPath:  "../../testdata/sample-repo",
				direction: "LR",
				format:    format,
			}
			err := run(cfg)
			if err != nil {
				t.Errorf("run() with format=%s should succeed, got: %v", format, err)
			}
		})
	}
}

func TestGraphErrorTypesWithErrorsAs(t *testing.T) {
	var circErr *graph.CircularDependencyError
	testErr := &graph.CircularDependencyError{Cycle: "a -> b -> a"}
	if !errors.As(testErr, &circErr) {
		t.Error("CircularDependencyError should match with errors.As")
	}
	if circErr.Cycle != "a -> b -> a" {
		t.Errorf("expected Cycle='a -> b -> a', got '%s'", circErr.Cycle)
	}

	var missErr *graph.MissingDependencyError
	testErr2 := &graph.MissingDependencyError{Challenge: "b", Dependency: "a"}
	if !errors.As(testErr2, &missErr) {
		t.Error("MissingDependencyError should match with errors.As")
	}
	if missErr.Challenge != "b" || missErr.Dependency != "a" {
		t.Errorf("expected Challenge='b' Dependency='a', got Challenge='%s' Dependency='%s'",
			missErr.Challenge, missErr.Dependency)
	}
}
