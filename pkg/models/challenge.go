package models

import "gopkg.in/yaml.v3"

// Challenge represents a CTF challenge with its metadata and dependencies
type Challenge struct {
	// Name is the unique identifier for the challenge
	Name string `yaml:"name"`

	// Category is the challenge category (e.g., web, crypto, pwn)
	Category string `yaml:"category,omitempty"`

	// Requirements lists the names of challenges that must be completed before this one
	Requirements Requirements `yaml:"requirements,omitempty"`

	// Path is the filesystem path to the challenge directory (not in YAML)
	Path string `yaml:"-"`
}

// Requirements holds the names of prerequisite challenges.
//
// It accepts two YAML shapes so both the simple list form and the CTFd/ctfcli
// native form can be used interchangeably:
//
//	requirements:
//	  - welcome
//	  - prism-1
//
//	requirements:
//	  prerequisites:
//	    - welcome
//	    - prism-1
//	  anonymize: true
//
// In both cases it is flattened to the list of prerequisite challenge names.
// The `anonymize` flag is a CTFd display option and is ignored here.
type Requirements []string

// UnmarshalYAML implements custom decoding for the two supported shapes.
func (r *Requirements) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.SequenceNode:
		var list []string
		if err := value.Decode(&list); err != nil {
			return err
		}
		*r = list
	case yaml.MappingNode:
		var m struct {
			Prerequisites []string `yaml:"prerequisites"`
			Anonymize     bool     `yaml:"anonymize"`
		}
		if err := value.Decode(&m); err != nil {
			return err
		}
		*r = m.Prerequisites
	default:
		// Null or any other scalar (e.g. an empty `requirements:`) means no deps.
		*r = nil
	}
	return nil
}

// ChallengeMetadata contains additional information about a challenge
type ChallengeMetadata struct {
	Challenge Challenge
	IsNew     bool // Whether this challenge was added in the current branch
}
