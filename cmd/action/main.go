package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/google/go-github/v58/github"
	"github.com/obaid/pr-guardian/internal/checker"
	"golang.org/x/oauth2"
)

func main() {
	// GitHub Actions injects these environment variables automatically.
	token := os.Getenv("GITHUB_TOKEN")
	repo := os.Getenv("GITHUB_REPOSITORY")   // format: "owner/repo"
	prNumberStr := os.Getenv("PR_NUMBER")     // set in action.yml via ${{ github.event.pull_request.number }}

	if token == "" || repo == "" || prNumberStr == "" {
		fmt.Println("❌ Missing required environment variables: GITHUB_TOKEN, GITHUB_REPOSITORY, PR_NUMBER")
		os.Exit(1)
	}

	prNumber, err := strconv.Atoi(prNumberStr)
	if err != nil {
		fmt.Printf("❌ Invalid PR number: %s\n", prNumberStr)
		os.Exit(1)
	}

	// Build the PR URL from env vars
	parts := splitRepo(repo)
	if len(parts) != 2 {
		fmt.Printf("❌ Invalid GITHUB_REPOSITORY format: %s\n", repo)
		os.Exit(1)
	}
	prURL := fmt.Sprintf("https://github.com/%s/%s/pull/%d", parts[0], parts[1], prNumber)

	fmt.Printf("🛡️  PR Guardian checking: %s\n", prURL)

	ctx := context.Background()
	results, err := checker.RunChecks(ctx, prURL, token)
	if err != nil {
		fmt.Printf("❌ Error running checks: %v\n", err)
		os.Exit(1)
	}

	// Print results to Action log
	passed := 0
	for _, r := range results {
		status := "PASS ✓"
		if !r.Passed {
			status = "FAIL ✗"
		} else {
			passed++
		}
		fmt.Printf("[%s] %s — %s\n", status, r.Name, r.Message)
	}
	fmt.Printf("\n%d/%d checks passed\n", passed, len(results))

	// Post comment on PR
	comment := checker.FormatComment(results)
	if err := postComment(ctx, token, parts[0], parts[1], prNumber, comment); err != nil {
		fmt.Printf("⚠️  Could not post comment: %v\n", err)
		// Don't fail the action just because commenting failed
	} else {
		fmt.Println("✅ Comment posted to PR.")
	}

	// Fail the action if any check failed (optional — configurable)
	failOnError := os.Getenv("FAIL_ON_ERROR")
	if failOnError == "true" && passed < len(results) {
		os.Exit(1)
	}
}

// postComment posts (or updates) the PR Guardian comment on a PR.
func postComment(ctx context.Context, token, owner, repo string, prNumber int, body string) error {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Check if we already posted a comment (to avoid spam on re-runs)
	comments, _, err := client.Issues.ListComments(ctx, owner, repo, prNumber, nil)
	if err == nil {
		for _, c := range comments {
			if c.Body != nil && len(*c.Body) > 20 {
				// Look for our signature
				if contains(*c.Body, "PR Guardian Report") {
					// Update existing comment
					_, _, err := client.Issues.EditComment(ctx, owner, repo, c.GetID(), &github.IssueComment{Body: &body})
					return err
				}
			}
		}
	}

	// Post new comment
	_, _, err = client.Issues.CreateComment(ctx, owner, repo, prNumber, &github.IssueComment{Body: &body})
	return err
}

func splitRepo(repo string) []string {
	for i, c := range repo {
		if c == '/' {
			return []string{repo[:i], repo[i+1:]}
		}
	}
	return []string{repo}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
