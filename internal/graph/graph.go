package graph

import (
	"fmt"
	"sort"

	"github.com/ryuse/challenge-deps/pkg/models"
)

// Graph represents a dependency graph of challenges
type Graph struct {
	nodes        map[string]*models.Challenge
	edges        map[string][]string // challenge -> list of dependencies
	reverseEdges map[string][]string // dependency -> list of challenges that depend on it
}

// New creates a new dependency graph
func New() *Graph {
	return &Graph{
		nodes:        make(map[string]*models.Challenge),
		edges:        make(map[string][]string),
		reverseEdges: make(map[string][]string),
	}
}

// AddChallenge adds a challenge to the graph
func (g *Graph) AddChallenge(challenge models.Challenge) {
	g.nodes[challenge.Name] = &challenge
	if g.edges[challenge.Name] == nil {
		g.edges[challenge.Name] = []string{}
	}

	// Add edges for dependencies
	for _, dep := range challenge.Requirements {
		g.edges[challenge.Name] = append(g.edges[challenge.Name], dep)

		// Add reverse edge
		if g.reverseEdges[dep] == nil {
			g.reverseEdges[dep] = []string{}
		}
		g.reverseEdges[dep] = append(g.reverseEdges[dep], challenge.Name)
	}
}

// Build constructs the graph from a list of challenges
func (g *Graph) Build(challenges []models.ChallengeMetadata) error {
	for _, ch := range challenges {
		g.AddChallenge(ch.Challenge)
	}

	// Validate that all dependencies exist
	if err := g.validateDependencies(); err != nil {
		return err
	}

	// Detect circular dependencies
	if err := g.detectCycles(); err != nil {
		return err
	}

	return nil
}

// validateDependencies checks that all required dependencies exist in the graph
func (g *Graph) validateDependencies() error {
	for challengeName, deps := range g.edges {
		for _, dep := range deps {
			if _, exists := g.nodes[dep]; !exists {
				return fmt.Errorf("challenge '%s' requires '%s' which does not exist", challengeName, dep)
			}
		}
	}
	return nil
}

// detectCycles uses DFS to detect circular dependencies in the graph
func (g *Graph) detectCycles() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for node := range g.nodes {
		if !visited[node] {
			if err := g.dfs(node, visited, recStack, []string{node}); err != nil {
				return err
			}
		}
	}

	return nil
}

// dfs performs depth-first search to detect cycles
func (g *Graph) dfs(node string, visited, recStack map[string]bool, path []string) error {
	visited[node] = true
	recStack[node] = true

	for _, neighbor := range g.edges[node] {
		if !visited[neighbor] {
			newPath := append(path, neighbor)
			if err := g.dfs(neighbor, visited, recStack, newPath); err != nil {
				return err
			}
		} else if recStack[neighbor] {
			// Cycle detected
			cyclePath := append(path, neighbor)
			return fmt.Errorf("circular dependency detected: %s", formatCycle(cyclePath))
		}
	}

	recStack[node] = false
	return nil
}

// formatCycle formats a cycle path for error messages
func formatCycle(path []string) string {
	result := ""
	for i, node := range path {
		if i > 0 {
			result += " -> "
		}
		result += node
	}
	return result
}

// GetEdges returns all edges in the graph
func (g *Graph) GetEdges() map[string][]string {
	return g.edges
}

// GetNodes returns all nodes in the graph
func (g *Graph) GetNodes() map[string]*models.Challenge {
	return g.nodes
}

// TopologicalSort returns challenges in dependency order (dependencies first)
func (g *Graph) TopologicalSort() ([]string, error) {
	inDegree := make(map[string]int)
	for node := range g.nodes {
		inDegree[node] = 0
	}

	// Calculate in-degrees: number of dependencies pointing to each node
	// For dependency resolution order, we count how many challenges depend on each challenge
	for node := range g.nodes {
		inDegree[node] = len(g.reverseEdges[node])
	}

	// Find nodes with no incoming edges (no one depends on them, so they can be done last)
	// Actually, for dependency resolution, we want nodes with no outgoing edges (no dependencies)
	queue := []string{}
	for node := range g.nodes {
		if len(g.edges[node]) == 0 {
			queue = append(queue, node)
		}
	}

	// Sort the initial queue for deterministic output
	sort.Strings(queue)

	result := []string{}
	processed := make(map[string]bool)

	for len(queue) > 0 {
		// Pop from queue
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)
		processed[node] = true

		// Get dependents (challenges that depend on this one)
		dependents := g.reverseEdges[node]
		sort.Strings(dependents) // For deterministic order

		for _, dependent := range dependents {
			// Check if all dependencies of this dependent are processed
			allDepsProcessed := true
			for _, dep := range g.edges[dependent] {
				if !processed[dep] {
					allDepsProcessed = false
					break
				}
			}

			if allDepsProcessed && !processed[dependent] {
				queue = append(queue, dependent)
			}
		}
	}

	// If not all nodes are in result, there's a cycle
	if len(result) != len(g.nodes) {
		return nil, fmt.Errorf("graph contains a cycle")
	}

	return result, nil
}

// GetDependents returns all challenges that depend on the given challenge
func (g *Graph) GetDependents(challengeName string) []string {
	return g.reverseEdges[challengeName]
}

// GetDependencies returns all challenges that the given challenge depends on
func (g *Graph) GetDependencies(challengeName string) []string {
	return g.edges[challengeName]
}
