# Challenge Dependencies

Visualize CTF challenge dependencies as Mermaid graphs. Works locally and in GitHub Actions.

## Installation

```bash
make build
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

see [testdata](./testdata)

## Output Example

```mermaid
graph LR
    challenge-2 --> challenge-1
    challenge-3 --> challenge-1
    challenge-4 --> challenge-2
    challenge-4 --> challenge-3
```

## GitHub Actions

```yaml
name: Check Dependencies

on:
  pull_request:
    branches: [main]
  issue_comment:
    types:
      - created

permissions:
  contents: read
  pull-requests: write
  issues: write

jobs:
  analyze:
    if: |
      github.event_name == 'pull_request' ||
      (
        github.event_name == 'issue_comment' &&
        github.event.issue.pull_request &&
        contains(github.event.comment.body, '@github deps') &&
        (
          github.event.comment.author_association == 'MEMBER' ||
          github.event.comment.author_association == 'OWNER' ||
          github.event.comment.author_association == 'COLLABORATOR'
        )
      )
    runs-on: ubuntu-latest
    steps:
      - name: Set PR info
        run: |
          if [[ "${EVENT_NAME}" == "pull_request" ]]; then
            echo "BRANCH_NAME=${PR_HEAD_REF}" >> $GITHUB_ENV
            echo "BASE_REF=${PR_BASE_REF}" >> $GITHUB_ENV
          else
            BRANCH_NAME=$(gh pr view "${ISSUE_NUMBER}" --json headRefName --jq .headRefName --repo "${REPO}")
            BASE_REF=$(gh pr view "${ISSUE_NUMBER}" --json baseRefName --jq .baseRefName --repo "${REPO}")
            echo "BRANCH_NAME=${BRANCH_NAME}" >> $GITHUB_ENV
            echo "BASE_REF=${BASE_REF}" >> $GITHUB_ENV
          fi
        env:
          GH_TOKEN: ${{ github.token }}
          EVENT_NAME: ${{ github.event_name }}
          PR_HEAD_REF: ${{ github.event.pull_request.head.ref }}
          PR_BASE_REF: ${{ github.base_ref }}
          ISSUE_NUMBER: ${{ github.event.issue.number }}
          REPO: ${{ github.repository }}

      - uses: actions/checkout@34e114876b0b11c390a56381ad16ebd13914f8d5 # v4.3.1
        with:
          ref: ${{ env.BRANCH_NAME }}
          fetch-depth: 0 # Required for branch comparison

      - name: Fetch base branch
        run: git fetch origin "${BASE_REF}"

      - name: Analyze dependencies
        id: deps
        uses: diver-osint-ctf/challenge_dependencies@caf73a4c80b3fa4e0cad62c82083d9ffd1fb7f0f # v1.0.1
        with:
          repo: "."
          base: "origin/${{ env.BASE_REF }}"
          head: "HEAD"

      - name: Comment PR
        uses: actions/github-script@ed597411d8f924073f98dfc5c65a23a2325f34cd # v8.0.0
        env:
          GRAPH_OUTPUT: ${{ steps.deps.outputs.graph }}
        with:
          script: |
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: process.env.GRAPH_OUTPUT
            });
```
