package mermaid

import (
	"fmt"
	"sort"
	"strings"

	"github.com/ryuse/challenge-deps/internal/graph"
	"github.com/ryuse/challenge-deps/pkg/models"
)

// Generator generates Mermaid diagrams from dependency graphs
type Generator struct {
	graph        *graph.Graph
	newChallenges map[string]bool
}

// New creates a new Mermaid generator
func New(g *graph.Graph) *Generator {
	return &Generator{
		graph:        g,
		newChallenges: make(map[string]bool),
	}
}

// SetNewChallenges marks certain challenges as new (added in current branch)
func (m *Generator) SetNewChallenges(challenges []models.ChallengeMetadata) {
	for _, ch := range challenges {
		if ch.IsNew {
			m.newChallenges[ch.Challenge.Name] = true
		}
	}
}

// getDisplayName returns the display name for a challenge in the format "category/name"
func (m *Generator) getDisplayName(challengeName string) string {
	nodes := m.graph.GetNodes()
	if challenge, ok := nodes[challengeName]; ok {
		if challenge.Category != "" {
			return fmt.Sprintf("%s/%s", challenge.Category, challenge.Name)
		}
	}
	return challengeName
}

// Generate creates a Mermaid diagram from the dependency graph
func (m *Generator) Generate(direction string) string {
	var builder strings.Builder

	// Validate direction
	if direction != "LR" && direction != "TB" && direction != "RL" && direction != "BT" {
		direction = "LR" // Default to left-right
	}

	// Start the diagram
	builder.WriteString(fmt.Sprintf("graph %s\n", direction))

	// Get all edges and sort them for deterministic output
	edges := m.graph.GetEdges()
	nodes := m.graph.GetNodes()

	// Collect all unique edges
	type edge struct {
		from string
		to   string
	}
	var allEdges []edge

	for challenge, deps := range edges {
		for _, dep := range deps {
			allEdges = append(allEdges, edge{from: challenge, to: dep})
		}
	}

	// Sort edges for deterministic output
	sort.Slice(allEdges, func(i, j int) bool {
		if allEdges[i].from != allEdges[j].from {
			return allEdges[i].from < allEdges[j].from
		}
		return allEdges[i].to < allEdges[j].to
	})

	// Add edges (dependencies)
	if len(allEdges) > 0 {
		builder.WriteString("    %% Dependencies\n")
		for _, e := range allEdges {
			fromDisplay := m.getDisplayName(e.from)
			toDisplay := m.getDisplayName(e.to)
			builder.WriteString(fmt.Sprintf("    %s --> %s\n", fromDisplay, toDisplay))
		}
	}

	// Add standalone nodes (challenges with no dependencies)
	var standaloneNodes []string
	for nodeName := range nodes {
		if len(edges[nodeName]) == 0 && len(m.graph.GetDependents(nodeName)) == 0 {
			standaloneNodes = append(standaloneNodes, nodeName)
		}
	}

	if len(standaloneNodes) > 0 {
		sort.Strings(standaloneNodes)
		builder.WriteString("\n    %% Standalone challenges\n")
		for _, node := range standaloneNodes {
			builder.WriteString(fmt.Sprintf("    %s\n", m.getDisplayName(node)))
		}
	}

	// Add styling for new challenges
	if len(m.newChallenges) > 0 {
		builder.WriteString("\n    %% Styling\n")
		builder.WriteString("    classDef newChallenge fill:#9f6,stroke:#333,stroke-width:2px\n")

		var newNodes []string
		for node := range m.newChallenges {
			newNodes = append(newNodes, node)
		}
		sort.Strings(newNodes)

		// Convert node names to display names
		var newDisplayNodes []string
		for _, node := range newNodes {
			newDisplayNodes = append(newDisplayNodes, m.getDisplayName(node))
		}

		builder.WriteString(fmt.Sprintf("    class %s newChallenge\n", strings.Join(newDisplayNodes, ",")))
	}

	return builder.String()
}

// GenerateMarkdown wraps the Mermaid diagram in markdown code fences
func (m *Generator) GenerateMarkdown(direction string) string {
	var builder strings.Builder

	builder.WriteString("```mermaid\n")
	builder.WriteString(m.Generate(direction))
	builder.WriteString("```\n")

	return builder.String()
}

// GenerateSummary creates a text summary of the dependencies
func (m *Generator) GenerateSummary() string {
	var builder strings.Builder

	nodes := m.graph.GetNodes()
	edges := m.graph.GetEdges()

	// Get sorted list of challenges
	var challenges []string
	for name := range nodes {
		challenges = append(challenges, name)
	}
	sort.Strings(challenges)

	builder.WriteString("# Challenge Dependencies Summary\n\n")
	builder.WriteString(fmt.Sprintf("Total challenges: %d\n\n", len(challenges)))

	// Group by dependency count
	noDeps := []string{}
	withDeps := []string{}

	for _, name := range challenges {
		if len(edges[name]) == 0 {
			noDeps = append(noDeps, name)
		} else {
			withDeps = append(withDeps, name)
		}
	}

	if len(noDeps) > 0 {
		builder.WriteString("## Challenges with no dependencies:\n")
		for _, name := range noDeps {
			isNew := ""
			if m.newChallenges[name] {
				isNew = " (NEW)"
			}
			builder.WriteString(fmt.Sprintf("- %s%s\n", m.getDisplayName(name), isNew))
		}
		builder.WriteString("\n")
	}

	if len(withDeps) > 0 {
		builder.WriteString("## Challenges with dependencies:\n")
		for _, name := range withDeps {
			isNew := ""
			if m.newChallenges[name] {
				isNew = " (NEW)"
			}
			deps := edges[name]
			sort.Strings(deps)
			// Convert dependency names to display names
			var displayDeps []string
			for _, dep := range deps {
				displayDeps = append(displayDeps, m.getDisplayName(dep))
			}
			builder.WriteString(fmt.Sprintf("- %s%s requires: %s\n", m.getDisplayName(name), isNew, strings.Join(displayDeps, ", ")))
		}
	}

	return builder.String()
}
