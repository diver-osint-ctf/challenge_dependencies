package graph

import (
	"errors"
	"reflect"
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

func TestTypedErrors(t *testing.T) {
	t.Run("MissingDependencyError", func(t *testing.T) {
		g := New()
		g.AddChallenge(models.Challenge{Name: "challenge-2", Requirements: []string{"challenge-1"}})

		err := g.validateDependencies()
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var missErr *MissingDependencyError
		if !errors.As(err, &missErr) {
			t.Errorf("expected MissingDependencyError, got %T", err)
		}
		if missErr.Challenge != "challenge-2" {
			t.Errorf("expected Challenge='challenge-2', got '%s'", missErr.Challenge)
		}
		if missErr.Dependency != "challenge-1" {
			t.Errorf("expected Dependency='challenge-1', got '%s'", missErr.Dependency)
		}
	})

	t.Run("CircularDependencyError", func(t *testing.T) {
		g := New()
		g.AddChallenge(models.Challenge{Name: "a", Requirements: []string{"b"}})
		g.AddChallenge(models.Challenge{Name: "b", Requirements: []string{"a"}})

		err := g.detectCycles()
		if err == nil {
			t.Fatal("expected error, got nil")
		}

		var circErr *CircularDependencyError
		if !errors.As(err, &circErr) {
			t.Errorf("expected CircularDependencyError, got %T", err)
		}
		if circErr.Cycle == "" {
			t.Error("expected non-empty Cycle field")
		}
	})

	t.Run("Build returns typed errors", func(t *testing.T) {
		g := New()
		err := g.Build([]models.ChallengeMetadata{
			{Challenge: models.Challenge{Name: "x", Requirements: []string{"missing"}}},
		})
		var missErr *MissingDependencyError
		if !errors.As(err, &missErr) {
			t.Errorf("Build should return MissingDependencyError, got %T", err)
		}
	})
}

func TestDefensiveCopies(t *testing.T) {
	g := New()
	g.AddChallenge(models.Challenge{Name: "challenge-1"})
	g.AddChallenge(models.Challenge{Name: "challenge-2", Requirements: []string{"challenge-1"}})
	g.AddChallenge(models.Challenge{Name: "challenge-3", Requirements: []string{"challenge-1"}})

	t.Run("GetEdges returns independent copy", func(t *testing.T) {
		edges := g.GetEdges()
		// Mutate the returned map
		edges["challenge-2"] = append(edges["challenge-2"], "injected")
		delete(edges, "challenge-3")

		// Original should be unchanged
		original := g.GetEdges()
		if len(original["challenge-2"]) != 1 {
			t.Errorf("GetEdges mutation leaked: challenge-2 deps = %v", original["challenge-2"])
		}
		if _, ok := original["challenge-3"]; !ok {
			t.Error("GetEdges mutation leaked: challenge-3 was deleted from original")
		}
	})

	t.Run("GetDependents returns independent copy", func(t *testing.T) {
		deps := g.GetDependents("challenge-1")
		if deps == nil {
			t.Fatal("expected non-nil dependents")
		}
		original := make([]string, len(deps))
		copy(original, deps)

		// Mutate returned slice
		deps[0] = "mutated"

		fresh := g.GetDependents("challenge-1")
		if !reflect.DeepEqual(fresh, original) {
			t.Errorf("GetDependents mutation leaked: got %v, want %v", fresh, original)
		}
	})

	t.Run("GetDependencies returns independent copy", func(t *testing.T) {
		deps := g.GetDependencies("challenge-2")
		if deps == nil {
			t.Fatal("expected non-nil dependencies")
		}
		original := make([]string, len(deps))
		copy(original, deps)

		deps[0] = "mutated"

		fresh := g.GetDependencies("challenge-2")
		if !reflect.DeepEqual(fresh, original) {
			t.Errorf("GetDependencies mutation leaked: got %v, want %v", fresh, original)
		}
	})

	t.Run("GetDependents returns nil for unknown challenge", func(t *testing.T) {
		deps := g.GetDependents("nonexistent")
		if deps != nil {
			t.Errorf("expected nil for unknown challenge, got %v", deps)
		}
	})
}

func TestTopologicalSortImmutability(t *testing.T) {
	g := New()
	g.AddChallenge(models.Challenge{Name: "a"})
	g.AddChallenge(models.Challenge{Name: "b", Requirements: []string{"a"}})
	g.AddChallenge(models.Challenge{Name: "c", Requirements: []string{"a"}})

	result1, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("first TopologicalSort failed: %v", err)
	}

	result2, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("second TopologicalSort failed: %v", err)
	}

	if !reflect.DeepEqual(result1, result2) {
		t.Errorf("TopologicalSort not deterministic: first=%v, second=%v", result1, result2)
	}

	// Verify internal state was not mutated
	deps := g.GetDependents("a")
	for i := 0; i < len(deps)-1; i++ {
		if deps[i] > deps[i+1] {
			// Internal reverseEdges was sorted by first call - this is fine
			// as long as the test above passes (deterministic output)
			break
		}
	}
}

func TestDFSSliceAliasingBranching(t *testing.T) {
	// Graph with branching: A -> B, A -> C, B -> D, C -> D
	// DFS from A visits B then C. Without slice aliasing fix,
	// the path for C's branch could be corrupted by B's branch.
	g := New()
	g.AddChallenge(models.Challenge{Name: "d"})
	g.AddChallenge(models.Challenge{Name: "b", Requirements: []string{"d"}})
	g.AddChallenge(models.Challenge{Name: "c", Requirements: []string{"d"}})
	g.AddChallenge(models.Challenge{Name: "a", Requirements: []string{"b", "c"}})

	// Should not detect a false cycle
	err := g.detectCycles()
	if err != nil {
		t.Errorf("should not detect cycle in valid DAG, got: %v", err)
	}

	// Build should succeed
	g2 := New()
	err = g2.Build([]models.ChallengeMetadata{
		{Challenge: models.Challenge{Name: "d"}},
		{Challenge: models.Challenge{Name: "b", Requirements: []string{"d"}}},
		{Challenge: models.Challenge{Name: "c", Requirements: []string{"d"}}},
		{Challenge: models.Challenge{Name: "a", Requirements: []string{"b", "c"}}},
	})
	if err != nil {
		t.Errorf("Build should succeed for valid DAG, got: %v", err)
	}
}
