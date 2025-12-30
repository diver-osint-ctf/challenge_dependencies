package models

// Challenge represents a CTF challenge with its metadata and dependencies
type Challenge struct {
	// Name is the unique identifier for the challenge
	Name string `yaml:"name"`

	// Category is the challenge category (e.g., web, crypto, pwn)
	Category string `yaml:"category,omitempty"`

	// Requirements lists the names of challenges that must be completed before this one
	Requirements []string `yaml:"requirements,omitempty"`

	// Path is the filesystem path to the challenge directory (not in YAML)
	Path string `yaml:"-"`
}

// ChallengeMetadata contains additional information about a challenge
type ChallengeMetadata struct {
	Challenge Challenge
	Category  string // e.g., "web", "crypto", "pwn"
	IsNew     bool   // Whether this challenge was added in the current branch
}
