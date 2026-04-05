package mermaid

import (
	"strings"
	"testing"

	"github.com/ryuse/challenge-deps/internal/graph"
	"github.com/ryuse/challenge-deps/pkg/models"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name       string
		challenges []models.Challenge
		direction  string
		want       []string // strings that should be in output
	}{
		{
			name: "simple dependency",
			challenges: []models.Challenge{
				{Name: "challenge-1"},
				{Name: "challenge-2", Requirements: []string{"challenge-1"}},
			},
			direction: "LR",
			want: []string{
				"graph LR",
				`challenge-2["challenge-2"] --> challenge-1["challenge-1"]`,
			},
		},
		{
			name: "multiple dependencies",
			challenges: []models.Challenge{
				{Name: "challenge-1"},
				{Name: "challenge-2", Requirements: []string{"challenge-1"}},
				{Name: "challenge-3", Requirements: []string{"challenge-1"}},
				{Name: "challenge-4", Requirements: []string{"challenge-2", "challenge-3"}},
			},
			direction: "TB",
			want: []string{
				"graph TB",
				`challenge-2["challenge-2"] --> challenge-1["challenge-1"]`,
				`challenge-3["challenge-3"] --> challenge-1["challenge-1"]`,
				`challenge-4["challenge-4"] --> challenge-2["challenge-2"]`,
				`challenge-4["challenge-4"] --> challenge-3["challenge-3"]`,
			},
		},
		{
			name: "no dependencies",
			challenges: []models.Challenge{
				{Name: "challenge-1"},
				{Name: "challenge-2"},
			},
			direction: "LR",
			want: []string{
				"graph LR",
				"Standalone",
				"challenge-1",
				"challenge-2",
			},
		},
		{
			name: "invalid direction defaults to LR",
			challenges: []models.Challenge{
				{Name: "challenge-1"},
			},
			direction: "INVALID",
			want: []string{
				"graph LR",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := graph.New()
			for _, ch := range tt.challenges {
				g.AddChallenge(ch)
			}

			gen := New(g)
			output := gen.Generate(tt.direction)

			for _, want := range tt.want {
				if !strings.Contains(output, want) {
					t.Errorf("Generate() output should contain '%s'\nGot:\n%s", want, output)
				}
			}
		})
	}
}

func TestSetNewChallenges(t *testing.T) {
	g := graph.New()
	g.AddChallenge(models.Challenge{Name: "challenge-1"})
	g.AddChallenge(models.Challenge{Name: "challenge-2", Requirements: []string{"challenge-1"}})

	gen := New(g)
	gen.SetNewChallenges([]models.ChallengeMetadata{
		{Challenge: models.Challenge{Name: "challenge-2"}, IsNew: true},
	})

	output := gen.Generate("LR")

	// Should contain styling for new challenges
	if !strings.Contains(output, "classDef newChallenge") {
		t.Error("Output should contain styling for new challenges")
	}

	if !strings.Contains(output, "class challenge-2 newChallenge") {
		t.Error("Output should mark challenge-2 as new")
	}
}

func TestGenerateMarkdown(t *testing.T) {
	g := graph.New()
	g.AddChallenge(models.Challenge{Name: "challenge-1"})
	g.AddChallenge(models.Challenge{Name: "challenge-2", Requirements: []string{"challenge-1"}})

	gen := New(g)
	output := gen.GenerateMarkdown("LR")

	// Should be wrapped in markdown code fences
	if !strings.HasPrefix(output, "```mermaid\n") {
		t.Error("Output should start with ```mermaid")
	}

	if !strings.HasSuffix(output, "```\n") {
		t.Error("Output should end with ```")
	}

	// Should contain the graph
	if !strings.Contains(output, "graph LR") {
		t.Error("Output should contain graph definition")
	}
}

func TestGenerateSummary(t *testing.T) {
	g := graph.New()
	g.AddChallenge(models.Challenge{Name: "challenge-1"})
	g.AddChallenge(models.Challenge{Name: "challenge-2", Requirements: []string{"challenge-1"}})
	g.AddChallenge(models.Challenge{Name: "challenge-3"})

	gen := New(g)
	gen.SetNewChallenges([]models.ChallengeMetadata{
		{Challenge: models.Challenge{Name: "challenge-3"}, IsNew: true},
	})

	output := gen.GenerateSummary()

	// Should contain summary header
	if !strings.Contains(output, "Challenge Dependencies Summary") {
		t.Error("Output should contain summary header")
	}

	// Should mention total count
	if !strings.Contains(output, "Total challenges: 3") {
		t.Error("Output should contain total challenge count")
	}

	// Should list challenges with no dependencies
	if !strings.Contains(output, "challenge-1") {
		t.Error("Output should list challenge-1")
	}

	// Should list challenges with dependencies
	if !strings.Contains(output, "challenge-2") {
		t.Error("Output should list challenge-2")
	}

	// Should mark new challenges
	if !strings.Contains(output, "challenge-3 (NEW)") {
		t.Error("Output should mark challenge-3 as NEW")
	}

	// Should show dependency relationships
	if !strings.Contains(output, "requires:") {
		t.Error("Output should show dependency relationships")
	}
}

func TestDirectionOptions(t *testing.T) {
	directions := []string{"LR", "TB", "RL", "BT"}

	g := graph.New()
	g.AddChallenge(models.Challenge{Name: "challenge-1"})

	gen := New(g)

	for _, dir := range directions {
		t.Run(dir, func(t *testing.T) {
			output := gen.Generate(dir)
			expected := "graph " + dir
			if !strings.Contains(output, expected) {
				t.Errorf("Output should contain '%s'", expected)
			}
		})
	}
}
