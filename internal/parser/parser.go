package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/ryuse/challenge-deps/pkg/models"
	"gopkg.in/yaml.v3"
)

const challengeFileName = "challenge.yml"

// Parser handles parsing of challenge YAML files
type Parser struct {
	repoPath string
}

// New creates a new Parser for the given repository path
func New(repoPath string) *Parser {
	return &Parser{
		repoPath: repoPath,
	}
}

// ParseAllChallenges walks the repository and parses all challenge.yml files
func (p *Parser) ParseAllChallenges() ([]models.ChallengeMetadata, error) {
	var challenges []models.ChallengeMetadata

	err := filepath.WalkDir(p.repoPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip hidden directories and .git
		if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}

		// Only process challenge.yml files
		if !d.IsDir() && d.Name() == challengeFileName {
			challenge, err := p.parseChallenge(path)
			if err != nil {
				return fmt.Errorf("failed to parse %s: %w", path, err)
			}

			// Extract category from path (e.g., "web", "crypto", "pwn")
			category := p.extractCategory(path)

			challenges = append(challenges, models.ChallengeMetadata{
				Challenge: *challenge,
				Category:  category,
				IsNew:     false, // Will be set by git operations
			})
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return challenges, nil
}

// parseChallenge reads and parses a single challenge.yml file
func (p *Parser) parseChallenge(path string) (*models.Challenge, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	var challenge models.Challenge
	if err := yaml.Unmarshal(data, &challenge); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Validate that the challenge has a name
	if challenge.Name == "" {
		return nil, fmt.Errorf("challenge name is required")
	}

	// Store the path for reference
	challenge.Path = filepath.Dir(path)

	// If category is not specified in YAML, extract from path
	if challenge.Category == "" {
		challenge.Category = p.extractCategory(path)
	}

	return &challenge, nil
}

// extractCategory extracts the category from the challenge path
// e.g., "/repo/web/challenge-1/challenge.yml" -> "web"
func (p *Parser) extractCategory(path string) string {
	relPath, err := filepath.Rel(p.repoPath, path)
	if err != nil {
		return "unknown"
	}

	parts := strings.Split(relPath, string(os.PathSeparator))
	if len(parts) >= 2 {
		return parts[0]
	}

	return "unknown"
}

// ParseChallengesByPaths parses only the challenges at the specified paths
func (p *Parser) ParseChallengesByPaths(paths []string) ([]models.ChallengeMetadata, error) {
	var challenges []models.ChallengeMetadata

	for _, path := range paths {
		fullPath := filepath.Join(p.repoPath, path)

		// Check if this is a challenge.yml file or a directory containing one
		info, err := os.Stat(fullPath)
		if err != nil {
			continue // Skip if file doesn't exist
		}

		challengePath := fullPath
		if info.IsDir() {
			challengePath = filepath.Join(fullPath, challengeFileName)
			if _, err := os.Stat(challengePath); err != nil {
				continue // Skip if challenge.yml doesn't exist
			}
		}

		challenge, err := p.parseChallenge(challengePath)
		if err != nil {
			return nil, err
		}

		category := p.extractCategory(challengePath)

		challenges = append(challenges, models.ChallengeMetadata{
			Challenge: *challenge,
			Category:  category,
			IsNew:     false,
		})
	}

	return challenges, nil
}
