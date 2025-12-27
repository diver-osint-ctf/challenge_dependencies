package git

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFilterChallengeFiles(t *testing.T) {
	tests := []struct {
		name  string
		files []string
		want  int
	}{
		{
			name: "mixed files",
			files: []string{
				"web/challenge-1/challenge.yml",
				"web/challenge-1/solution.json",
				"crypto/challenge-2/challenge.yml",
				"README.md",
			},
			want: 2,
		},
		{
			name: "no challenge files",
			files: []string{
				"README.md",
				"go.mod",
			},
			want: 0,
		},
		{
			name: "all challenge files",
			files: []string{
				"web/challenge-1/challenge.yml",
				"web/challenge-2/challenge.yml",
				"crypto/challenge-3/challenge.yml",
			},
			want: 3,
		},
		{
			name:  "empty list",
			files: []string{},
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FilterChallengeFiles(tt.files)
			if len(got) != tt.want {
				t.Errorf("FilterChallengeFiles() returned %d files, want %d", len(got), tt.want)
			}

			// Verify all returned files are challenge.yml
			for _, f := range got {
				if filepath.Base(f) != "challenge.yml" {
					t.Errorf("FilterChallengeFiles() returned non-challenge file: %s", f)
				}
			}
		})
	}
}

func TestGetEnvOrDefault(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		want         string
	}{
		{
			name:         "env var exists",
			key:          "TEST_VAR_EXISTS",
			defaultValue: "default",
			envValue:     "custom",
			want:         "custom",
		},
		{
			name:         "env var does not exist",
			key:          "TEST_VAR_NOT_EXISTS",
			defaultValue: "default",
			envValue:     "",
			want:         "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			}

			got := GetEnvOrDefault(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("GetEnvOrDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetGitHubActionsBranches(t *testing.T) {
	tests := []struct {
		name         string
		baseRef      string
		headRef      string
		wantBase     string
		wantHead     string
		wantIsGitHub bool
	}{
		{
			name:         "both env vars set",
			baseRef:      "main",
			headRef:      "feature-branch",
			wantBase:     "main",
			wantHead:     "feature-branch",
			wantIsGitHub: true,
		},
		{
			name:         "no env vars set",
			baseRef:      "",
			headRef:      "",
			wantBase:     "",
			wantHead:     "",
			wantIsGitHub: false,
		},
		{
			name:         "only base set",
			baseRef:      "main",
			headRef:      "",
			wantBase:     "",
			wantHead:     "",
			wantIsGitHub: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.baseRef != "" {
				os.Setenv("GITHUB_BASE_REF", tt.baseRef)
				defer os.Unsetenv("GITHUB_BASE_REF")
			}
			if tt.headRef != "" {
				os.Setenv("GITHUB_HEAD_REF", tt.headRef)
				defer os.Unsetenv("GITHUB_HEAD_REF")
			}

			base, head, isGitHub := GetGitHubActionsBranches()

			if base != tt.wantBase {
				t.Errorf("GetGitHubActionsBranches() base = %v, want %v", base, tt.wantBase)
			}
			if head != tt.wantHead {
				t.Errorf("GetGitHubActionsBranches() head = %v, want %v", head, tt.wantHead)
			}
			if isGitHub != tt.wantIsGitHub {
				t.Errorf("GetGitHubActionsBranches() isGitHub = %v, want %v", isGitHub, tt.wantIsGitHub)
			}
		})
	}
}

func TestIsGitRepository(t *testing.T) {
	// Test with a temporary directory that's not a git repo
	tmpDir := t.TempDir()

	git := New(tmpDir)
	if git.IsGitRepository() {
		t.Error("IsGitRepository() should return false for non-git directory")
	}

	// Create a fake .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0755); err != nil {
		t.Fatalf("failed to create .git directory: %v", err)
	}

	if !git.IsGitRepository() {
		t.Error("IsGitRepository() should return true when .git directory exists")
	}
}

func TestNew(t *testing.T) {
	repoPath := "/test/repo"
	git := New(repoPath)

	if git.repoPath != repoPath {
		t.Errorf("New() repoPath = %v, want %v", git.repoPath, repoPath)
	}
}
