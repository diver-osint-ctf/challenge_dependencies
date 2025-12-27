# Challenge Dependencies

Visualize CTF challenge dependencies as Mermaid graphs. Works locally and in GitHub Actions.

## Installation

```bash
go build -o bin/challenge-deps ./cmd/challenge-deps
```

## Usage

```bash
# Analyze current branch only
./bin/challenge-deps --repo ./testdata

# Compare two branches
./bin/challenge-deps --repo ./testdata --base main --head feature-branch
```

### Options

- `--repo`: Challenge repository path (required)
- `--base`: Base branch (optional, requires --head)
- `--head`: Head branch (optional, requires --base)
- `--format`: Output format: `markdown`, `mermaid`, `summary` (default: `markdown`)
- `--direction`: Graph direction: `LR`, `TB`, `RL`, `BT` (default: `LR`)

## Challenge Format

Each challenge needs a `challenge.yml`:

```yaml
name: challenge-2
requirements:
  - challenge-1
```

## Output Example

```mermaid
graph LR
    challenge-2 --> challenge-1
    challenge-3 --> challenge-1
    challenge-4 --> challenge-2
    challenge-4 --> challenge-3
```

## GitHub Actions

### Use as an Action

**Simple usage (current branch only):**

```yaml
- name: Analyze dependencies
  uses: diver-osint-ctf/challenge_dependencies@main
  with:
    repo: '.'
```

**With branch comparison:**

```yaml
name: Check Dependencies

on:
  pull_request:
    branches: [main]

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Required for branch comparison

      - name: Fetch base branch
        run: git fetch origin ${{ github.base_ref }}

      - name: Analyze dependencies
        id: deps
        uses: diver-osint-ctf/challenge_dependencies@main
        with:
          repo: '.'
          base: 'origin/${{ github.base_ref }}'
          head: 'HEAD'

      - name: Comment PR
        uses: actions/github-script@v7
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: '${{ steps.deps.outputs.graph }}'
            });
```

**Important:** When comparing branches, you must:
1. Use `fetch-depth: 0` to get full git history
2. Explicitly fetch the base branch with `git fetch origin <branch>`
3. Use `origin/<branch>` format for the base parameter

### Inputs

- `repo`: Repository path (default: `.`)
- `base`: Base branch for comparison (optional)
- `head`: Head branch for comparison (optional)
- `format`: Output format - `markdown`, `mermaid`, `summary` (default: `markdown`)
- `direction`: Graph direction - `LR`, `TB`, `RL`, `BT` (default: `LR`)

### Outputs

- `graph`: Generated dependency graph
- `summary`: Text summary of dependencies

## Development

```bash
# Run tests
go test ./...

# Build
go build -o bin/challenge-deps ./cmd/challenge-deps
```
