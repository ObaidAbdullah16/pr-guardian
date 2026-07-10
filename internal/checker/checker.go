package checker

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/google/go-github/v58/github"
	"golang.org/x/oauth2"
)

// CheckResult holds the outcome of a single PR check.
type CheckResult struct {
	Name    string `json:"name"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
	Icon    string `json:"icon"`
}

// PRInfo holds data fetched from the GitHub API about a PR.
type PRInfo struct {
	Owner       string
	Repo        string
	Number      int
	Body        string
	HeadBranch  string
	ChangedFiles []string
}

var (
	// Branch name must start with one of these prefixes.
	validBranchPrefixes = []string{"feat/", "fix/", "docs/", "chore/", "refactor/", "test/", "hotfix/"}

	// PR body must mention a closing keyword + issue number.
	issuePattern = regexp.MustCompile(`(?i)(fixes|closes|resolves|fix|close|resolve)\s*#\d+|#\d+`)

	// Markdown link pattern: [text](url)
	mdLinkPattern = regexp.MustCompile(`\[.*?\]\((https?://[^\s)]+)\)`)
)

// RunChecks fetches PR data from GitHub and runs all 7 checks.
// prURL should be in the format: https://github.com/owner/repo/pull/123
func RunChecks(ctx context.Context, prURL, token string) ([]CheckResult, error) {
	info, client, err := fetchPRInfo(ctx, prURL, token)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch PR info: %w", err)
	}

	results := []CheckResult{
		checkREADME(info),
		checkIssueLinked(info),
		checkChangelog(info),
		checkContributorGuide(ctx, client, info),
		checkPRDescription(info),
		checkBranchName(info),
		checkBrokenLinks(info),
	}

	return results, nil
}

// fetchPRInfo parses the PR URL and fetches data from GitHub API.
func fetchPRInfo(ctx context.Context, prURL, token string) (*PRInfo, *github.Client, error) {
	// Parse: https://github.com/owner/repo/pull/123
	prURL = strings.TrimRight(prURL, "/")
	parts := strings.Split(prURL, "/")
	if len(parts) < 7 || parts[5] != "pull" {
		return nil, nil, fmt.Errorf("invalid PR URL format. Expected: https://github.com/owner/repo/pull/123")
	}

	owner := parts[3]
	repo := parts[4]
	var number int
	if _, err := fmt.Sscanf(parts[6], "%d", &number); err != nil {
		return nil, nil, fmt.Errorf("invalid PR number in URL")
	}

	// Set up GitHub client
	var client *github.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(ctx, ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(nil)
	}

	// Fetch PR metadata
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, nil, fmt.Errorf("could not fetch PR (check your token and PR URL): %w", err)
	}

	// Fetch list of files changed in the PR
	files, _, err := client.PullRequests.ListFiles(ctx, owner, repo, number, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, nil, fmt.Errorf("could not fetch PR files: %w", err)
	}

	changedFiles := make([]string, 0, len(files))
	for _, f := range files {
		changedFiles = append(changedFiles, f.GetFilename())
	}

	body := ""
	if pr.Body != nil {
		body = *pr.Body
	}

	info := &PRInfo{
		Owner:        owner,
		Repo:         repo,
		Number:       number,
		Body:         body,
		HeadBranch:   pr.GetHead().GetRef(),
		ChangedFiles: changedFiles,
	}

	return info, client, nil
}

// --- Individual Checks ---

func checkREADME(info *PRInfo) CheckResult {
	for _, f := range info.ChangedFiles {
		if strings.EqualFold(f, "readme.md") || strings.EqualFold(f, "readme") {
			return CheckResult{
				Name:    "README Updated",
				Passed:  true,
				Message: "README.md was updated in this PR.",
				Icon:    "✓",
			}
		}
	}
	return CheckResult{
		Name:    "README Updated",
		Passed:  false,
		Message: "README.md was not touched. Consider updating it if your change affects usage or setup.",
		Icon:    "✗",
	}
}

func checkIssueLinked(info *PRInfo) CheckResult {
	if issuePattern.MatchString(info.Body) {
		return CheckResult{
			Name:    "Issue Linked",
			Passed:  true,
			Message: "PR description references an issue (e.g. Fixes #123).",
			Icon:    "✓",
		}
	}
	return CheckResult{
		Name:    "Issue Linked",
		Passed:  false,
		Message: "No issue reference found. Add 'Fixes #<issue_number>' to your PR description.",
		Icon:    "✗",
	}
}

func checkChangelog(info *PRInfo) CheckResult {
	for _, f := range info.ChangedFiles {
		if strings.EqualFold(f, "changelog.md") || strings.EqualFold(f, "changelog") {
			return CheckResult{
				Name:    "Changelog Updated",
				Passed:  true,
				Message: "CHANGELOG.md was updated.",
				Icon:    "✓",
			}
		}
	}
	return CheckResult{
		Name:    "Changelog Updated",
		Passed:  false,
		Message: "CHANGELOG.md was not updated. Add an entry describing your change.",
		Icon:    "✗",
	}
}

func checkContributorGuide(ctx context.Context, client *github.Client, info *PRInfo) CheckResult {
	// Check if CONTRIBUTING.md exists in the repo root
	_, _, resp, err := client.Repositories.GetContents(ctx, info.Owner, info.Repo, "CONTRIBUTING.md", nil)
	if err == nil || (resp != nil && resp.StatusCode == http.StatusOK) {
		return CheckResult{
			Name:    "Contributor Guide Exists",
			Passed:  true,
			Message: "CONTRIBUTING.md exists in this repository.",
			Icon:    "✓",
		}
	}
	return CheckResult{
		Name:    "Contributor Guide Exists",
		Passed:  false,
		Message: "No CONTRIBUTING.md found. Add one so contributors know the rules.",
		Icon:    "✗",
	}
}

func checkPRDescription(info *PRInfo) CheckResult {
	trimmed := strings.TrimSpace(info.Body)
	if len(trimmed) >= 30 {
		return CheckResult{
			Name:    "PR Description",
			Passed:  true,
			Message: fmt.Sprintf("PR has a meaningful description (%d characters).", len(trimmed)),
			Icon:    "✓",
		}
	}
	return CheckResult{
		Name:    "PR Description",
		Passed:  false,
		Message: fmt.Sprintf("PR description is too short (%d chars). Write at least 30 characters explaining what and why.", len(trimmed)),
		Icon:    "✗",
	}
}

func checkBranchName(info *PRInfo) CheckResult {
	branch := info.HeadBranch
	for _, prefix := range validBranchPrefixes {
		if strings.HasPrefix(branch, prefix) {
			return CheckResult{
				Name:    "Branch Name Convention",
				Passed:  true,
				Message: fmt.Sprintf("Branch '%s' follows naming convention.", branch),
				Icon:    "✓",
			}
		}
	}
	return CheckResult{
		Name:    "Branch Name Convention",
		Passed:  false,
		Message: fmt.Sprintf("Branch '%s' doesn't follow convention. Use prefixes: feat/, fix/, docs/, chore/, refactor/", branch),
		Icon:    "✗",
	}
}

func checkBrokenLinks(info *PRInfo) CheckResult {
	// Collect all HTTP links from PR body
	matches := mdLinkPattern.FindAllStringSubmatch(info.Body, -1)
	if len(matches) == 0 {
		return CheckResult{
			Name:    "No Broken Links",
			Passed:  true,
			Message: "No markdown links found in PR description to check.",
			Icon:    "✓",
		}
	}

	httpClient := &http.Client{Timeout: 5 * time.Second}
	broken := []string{}

	for _, match := range matches {
		url := match[1]
		resp, err := httpClient.Head(url)
		if err != nil || resp.StatusCode >= 400 {
			broken = append(broken, url)
		}
	}

	if len(broken) == 0 {
		return CheckResult{
			Name:    "No Broken Links",
			Passed:  true,
			Message: fmt.Sprintf("All %d link(s) in PR description are reachable.", len(matches)),
			Icon:    "✓",
		}
	}

	return CheckResult{
		Name:    "No Broken Links",
		Passed:  false,
		Message: fmt.Sprintf("%d broken link(s) found: %s", len(broken), strings.Join(broken, ", ")),
		Icon:    "✗",
	}
}

// FormatComment formats results as a GitHub Markdown comment for posting on a PR.
func FormatComment(results []CheckResult) string {
	passed := 0
	for _, r := range results {
		if r.Passed {
			passed++
		}
	}

	var sb strings.Builder
	sb.WriteString("## 🛡️ PR Guardian Report\n\n")
	sb.WriteString(fmt.Sprintf("**%d / %d checks passed**\n\n", passed, len(results)))
	sb.WriteString("| Check | Status | Message |\n")
	sb.WriteString("|-------|--------|---------|\n")

	for _, r := range results {
		status := "✅ Pass"
		if !r.Passed {
			status = "❌ Fail"
		}
		sb.WriteString(fmt.Sprintf("| %s | %s | %s |\n", r.Name, status, r.Message))
	}

	sb.WriteString("\n---\n")
	sb.WriteString("*Generated by [PR Guardian](https://github.com/obaid/pr-guardian)*")
	return sb.String()
}
