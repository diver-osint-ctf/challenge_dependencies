package parser

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ryuse/challenge-deps/pkg/models"
)

func TestParseChallenge(t *testing.T) {
	tests := []struct {
		name        string
		yamlContent string
		want        *models.Challenge
		wantErr     bool
	}{
		{
			name: "valid challenge with dependencies",
			yamlContent: `name: challenge-2
requirements:
  - challenge-1
  - challenge-0`,
			want: &models.Challenge{
				Name:         "challenge-2",
				Requirements: []string{"challenge-1", "challenge-0"},
			},
			wantErr: false,
		},
		{
			name: "valid challenge without dependencies",
			yamlContent: `name: challenge-1`,
			want: &models.Challenge{
				Name:         "challenge-1",
				Requirements: nil,
			},
			wantErr: false,
		},
		{
			name:        "missing name",
			yamlContent: `requirements: []`,
			want:        nil,
			wantErr:     true,
		},
		{
			name:        "invalid yaml",
			yamlContent: `name: [invalid yaml structure`,
			want:        nil,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory and file
			tmpDir := t.TempDir()
			tmpFile := filepath.Join(tmpDir, "challenge.yml")
			err := os.WriteFile(tmpFile, []byte(tt.yamlContent), 0644)
			if err != nil {
				t.Fatalf("failed to create test file: %v", err)
			}

			parser := New(tmpDir)
			got, err := parser.parseChallenge(tmpFile)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseChallenge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				return
			}

			if got.Name != tt.want.Name {
				t.Errorf("parseChallenge() Name = %v, want %v", got.Name, tt.want.Name)
			}

			if len(got.Requirements) != len(tt.want.Requirements) {
				t.Errorf("parseChallenge() Requirements length = %v, want %v",
					len(got.Requirements), len(tt.want.Requirements))
				return
			}

			for i, req := range got.Requirements {
				if req != tt.want.Requirements[i] {
					t.Errorf("parseChallenge() Requirements[%d] = %v, want %v",
						i, req, tt.want.Requirements[i])
				}
			}
		})
	}
}

func TestExtractCategory(t *testing.T) {
	tests := []struct {
		name     string
		repoPath string
		filePath string
		want     string
	}{
		{
			name:     "web category",
			repoPath: "/repo",
			filePath: "/repo/web/challenge-1/challenge.yml",
			want:     "web",
		},
		{
			name:     "crypto category",
			repoPath: "/repo",
			filePath: "/repo/crypto/challenge-3/challenge.yml",
			want:     "crypto",
		},
		{
			name:     "nested path",
			repoPath: "/home/user/ctf",
			filePath: "/home/user/ctf/pwn/easy/challenge-1/challenge.yml",
			want:     "pwn",
		},
		{
			name:     "root level challenge",
			repoPath: "/repo",
			filePath: "/repo/challenge.yml",
			want:     "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := New(tt.repoPath)
			got := parser.extractCategory(tt.filePath)
			if got != tt.want {
				t.Errorf("extractCategory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseAllChallenges(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create test challenges
	challenges := map[string]string{
		"web/challenge-1/challenge.yml": `name: challenge-1`,
		"web/challenge-2/challenge.yml": `name: challenge-2
requirements:
  - challenge-1`,
		"crypto/challenge-3/challenge.yml": `name: challenge-3
requirements:
  - challenge-1`,
	}

	for path, content := range challenges {
		fullPath := filepath.Join(tmpDir, path)
		dir := filepath.Dir(fullPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("failed to create test file: %v", err)
		}
	}

	parser := New(tmpDir)
	got, err := parser.ParseAllChallenges()
	if err != nil {
		t.Fatalf("ParseAllChallenges() error = %v", err)
	}

	if len(got) != 3 {
		t.Errorf("ParseAllChallenges() returned %d challenges, want 3", len(got))
	}

	// Verify categories
	categoryCount := make(map[string]int)
	for _, ch := range got {
		categoryCount[ch.Challenge.Category]++
	}

	if categoryCount["web"] != 2 {
		t.Errorf("expected 2 web challenges, got %d", categoryCount["web"])
	}
	if categoryCount["crypto"] != 1 {
		t.Errorf("expected 1 crypto challenge, got %d", categoryCount["crypto"])
	}
}
