# 🛡️ PR Guardian

> A GitHub Action that automatically checks Pull Requests for common contributor mistakes and posts a checklist comment.

![GitHub Action](https://img.shields.io/badge/GitHub%20Action-available-brightgreen)
![Go](https://img.shields.io/badge/built%20with-Go-00ADD8)
![License](https://img.shields.io/badge/license-MIT-blue)

---

## What it does

When a contributor opens or updates a Pull Request in your repo, PR Guardian:

1. Fetches the PR metadata using the GitHub API
2. Runs 7 automated checks
3. Posts a formatted comment with the results

**The 7 checks:**

| Check | What it verifies |
|-------|-----------------|
| ✓ README Updated | Was `README.md` modified in this PR? |
| ✓ Issue Linked | Does the PR body mention `Fixes #123` or `Closes #`? |
| ✓ Changelog Updated | Was `CHANGELOG.md` updated? |
| ✓ Contributor Guide Exists | Does `CONTRIBUTING.md` exist in the repo? |
| ✓ PR Description | Is the PR description at least 30 characters? |
| ✓ Branch Name Convention | Does the branch start with `feat/`, `fix/`, `docs/`, `chore/`, or `refactor/`? |
| ✓ No Broken Links | Are all Markdown links in the PR description reachable? |

---

## Install in 60 seconds

Create `.github/workflows/pr-guardian.yml` in your repository:

```yaml
name: PR Guardian

on:
  pull_request:
    types: [opened, synchronize, reopened]

permissions:
  pull-requests: write
  contents: read

jobs:
  pr-guardian:
    runs-on: ubuntu-latest
    steps:
      - uses: obaid/pr-guardian@v1
        with:
          github_token: ${{ secrets.GITHUB_TOKEN }}
          fail_on_error: "false"   # set to "true" to block merging on failed checks
```

That's it. PR Guardian will automatically comment on every new PR.

---

## Inputs

| Input | Required | Default | Description |
|-------|----------|---------|-------------|
| `github_token` | Yes | `secrets.GITHUB_TOKEN` | Token for API access and posting comments |
| `fail_on_error` | No | `false` | Set to `true` to fail the Action (block merge) if any check fails |

---

## Example PR Comment

```
🛡️ PR Guardian Report

5 / 7 checks passed

| Check                    | Status  | Message                                         |
|--------------------------|---------|--------------------------------------------------|
| README Updated           | ✅ Pass | README.md was updated in this PR.               |
| Issue Linked             | ❌ Fail | No issue reference found. Add 'Fixes #<num>'    |
| Changelog Updated        | ❌ Fail | CHANGELOG.md was not updated.                   |
| Contributor Guide Exists | ✅ Pass | CONTRIBUTING.md exists in this repository.      |
| PR Description           | ✅ Pass | PR has a meaningful description (87 characters) |
| Branch Name Convention   | ✅ Pass | Branch 'feat/add-login' follows naming conv.    |
| No Broken Links          | ✅ Pass | All 2 links are reachable.                      |
```

---

## Live Demo

Try the live checker at **[pr-guardian.obaidinfo.xyz](https://pr-guardian.obaidinfo.xyz)** — paste any public PR URL and see the checks run instantly.

---

## Project Structure

```
pr-guardian/
├── action.yml              ← GitHub Action definition
├── Dockerfile              ← Container for the Action binary
├── go.mod / go.sum
├── cmd/
│   ├── action/main.go      ← Action binary entry point
│   └── lambda/main.go      ← AWS Lambda handler
├── internal/
│   └── checker/
│       ├── checker.go      ← Core check logic (shared)
│       └── checker_test.go ← Unit tests
└── web/                    ← Landing page + live checker
```

---

## Local Development

```bash
# Run tests
go test ./internal/checker/...

# Build the action binary
go build -o pr-guardian-action ./cmd/action

# Build the lambda binary (Linux, for AWS)
GOOS=linux GOARCH=amd64 go build -o bootstrap ./cmd/lambda
```

---

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md).

## Changelog

See [CHANGELOG.md](CHANGELOG.md).

## License

MIT
