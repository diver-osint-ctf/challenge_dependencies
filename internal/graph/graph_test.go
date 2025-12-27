package graph

import (
	"strings"
	"testing"

	"github.com/ryuse/challenge-deps/pkg/models"
)

func TestAddChallenge(t *testing.T) {
	g := New()

	challenge := models.Challenge{
		Name:         "challenge-1",
		Requirements: []string{"challenge-0"},
	}

	g.AddChallenge(challenge)

	if len(g.nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(g.nodes))
	}

	if _, exists := g.nodes["challenge-1"]; !exists {
		t.Error("challenge-1 not found in nodes")
	}

	deps := g.edges["challenge-1"]
	if len(deps) != 1 || deps[0] != "challenge-0" {
		t.Errorf("expected dependency on challenge-0, got %v", deps)
	}
}

func TestValidateDependencies(t *testing.T) {
	tests := []struct {
		name       string
		challenges []models.Challenge
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid dependencies",
			challenges: []models.Challenge{
				{Name: "challenge-1"},
				{Name: "challenge-2", Requirements: []string{"challenge-1"}},
			},
			wantErr: false,
		},
		{
			name: "missing dependency",
			challenges: []models.Challenge{
				{Name: "challenge-2", Requirements: []string{"challenge-1"}},
			},
			wantErr: true,
			errMsg:  "does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New()
			for _, ch := range tt.challenges {
				g.AddChallenge(ch)
			}

			err := g.validateDependencies()
			if (err != nil) != tt.wantErr {
				t.Errorf("validateDependencies() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error message should contain '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

func TestDetectCycles(t *testing.T) {
	tests := []struct {
		name       string
		challenges []models.Challenge
		wantErr    bool
		errMsg     string
	}{
		{
			name: "no cycle",
			challenges: []models.Challenge{
				{Name: "challenge-1"},
				{Name: "challenge-2", Requirements: []string{"challenge-1"}},
				{Name: "challenge-3", Requirements: []string{"challenge-2"}},
			},
			wantErr: false,
		},
		{
			name: "direct cycle",
			challenges: []models.Challenge{
				{Name: "challenge-1", Requirements: []string{"challenge-2"}},
				{Name: "challenge-2", Requirements: []string{"challenge-1"}},
			},
			wantErr: true,
			errMsg:  "circular dependency",
		},
		{
			name: "indirect cycle",
			challenges: []models.Challenge{
				{Name: "challenge-1", Requirements: []string{"challenge-2"}},
				{Name: "challenge-2", Requirements: []string{"challenge-3"}},
				{Name: "challenge-3", Requirements: []string{"challenge-1"}},
			},
			wantErr: true,
			errMsg:  "circular dependency",
		},
		{
			name: "self cycle",
			challenges: []models.Challenge{
				{Name: "challenge-1", Requirements: []string{"challenge-1"}},
			},
			wantErr: true,
			errMsg:  "circular dependency",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New()
			for _, ch := range tt.challenges {
				g.AddChallenge(ch)
			}

			err := g.detectCycles()
			if (err != nil) != tt.wantErr {
				t.Errorf("detectCycles() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error message should contain '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

func TestBuild(t *testing.T) {
	tests := []struct {
		name       string
		challenges []models.ChallengeMetadata
		wantErr    bool
		errMsg     string
	}{
		{
			name: "valid graph",
			challenges: []models.ChallengeMetadata{
				{Challenge: models.Challenge{Name: "challenge-1"}},
				{Challenge: models.Challenge{Name: "challenge-2", Requirements: []string{"challenge-1"}}},
			},
			wantErr: false,
		},
		{
			name: "missing dependency",
			challenges: []models.ChallengeMetadata{
				{Challenge: models.Challenge{Name: "challenge-2", Requirements: []string{"challenge-1"}}},
			},
			wantErr: true,
			errMsg:  "does not exist",
		},
		{
			name: "circular dependency",
			challenges: []models.ChallengeMetadata{
				{Challenge: models.Challenge{Name: "challenge-1", Requirements: []string{"challenge-2"}}},
				{Challenge: models.Challenge{Name: "challenge-2", Requirements: []string{"challenge-1"}}},
			},
			wantErr: true,
			errMsg:  "circular",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New()
			err := g.Build(tt.challenges)

			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("error message should contain '%s', got '%s'", tt.errMsg, err.Error())
			}
		})
	}
}

func TestTopologicalSort(t *testing.T) {
	tests := []struct {
		name       string
		challenges []models.Challenge
		wantErr    bool
	}{
		{
			name: "linear dependency chain",
			challenges: []models.Challenge{
				{Name: "challenge-1"},
				{Name: "challenge-2", Requirements: []string{"challenge-1"}},
				{Name: "challenge-3", Requirements: []string{"challenge-2"}},
			},
			wantErr: false,
		},
		{
			name: "diamond dependency",
			challenges: []models.Challenge{
				{Name: "challenge-1"},
				{Name: "challenge-2", Requirements: []string{"challenge-1"}},
				{Name: "challenge-3", Requirements: []string{"challenge-1"}},
				{Name: "challenge-4", Requirements: []string{"challenge-2", "challenge-3"}},
			},
			wantErr: false,
		},
		{
			name: "no dependencies",
			challenges: []models.Challenge{
				{Name: "challenge-1"},
				{Name: "challenge-2"},
				{Name: "challenge-3"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New()
			for _, ch := range tt.challenges {
				g.AddChallenge(ch)
			}

			result, err := g.TopologicalSort()
			if (err != nil) != tt.wantErr {
				t.Errorf("TopologicalSort() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			// Verify that dependencies come before dependents
			position := make(map[string]int)
			for i, name := range result {
				position[name] = i
			}

			for _, ch := range tt.challenges {
				for _, dep := range ch.Requirements {
					if position[dep] >= position[ch.Name] {
						t.Errorf("dependency '%s' should come before '%s' in topological order",
							dep, ch.Name)
					}
				}
			}
		})
	}
}

func TestGetDependentsAndDependencies(t *testing.T) {
	g := New()
	g.AddChallenge(models.Challenge{Name: "challenge-1"})
	g.AddChallenge(models.Challenge{Name: "challenge-2", Requirements: []string{"challenge-1"}})
	g.AddChallenge(models.Challenge{Name: "challenge-3", Requirements: []string{"challenge-1"}})

	// Test GetDependents
	dependents := g.GetDependents("challenge-1")
	if len(dependents) != 2 {
		t.Errorf("expected 2 dependents for challenge-1, got %d", len(dependents))
	}

	// Test GetDependencies
	deps := g.GetDependencies("challenge-2")
	if len(deps) != 1 || deps[0] != "challenge-1" {
		t.Errorf("expected challenge-2 to depend on challenge-1, got %v", deps)
	}
}
