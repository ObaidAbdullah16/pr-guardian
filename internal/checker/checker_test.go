package checker

import (
	"testing"
)

func TestCheckBranchName(t *testing.T) {
	tests := []struct {
		branch string
		want   bool
	}{
		{"feat/add-login", true},
		{"fix/null-pointer", true},
		{"docs/update-readme", true},
		{"chore/update-deps", true},
		{"refactor/cleanup", true},
		{"my-random-branch", false},
		{"main", false},
		{"update-stuff", false},
	}

	for _, tt := range tests {
		info := &PRInfo{HeadBranch: tt.branch}
		result := checkBranchName(info)
		if result.Passed != tt.want {
			t.Errorf("checkBranchName(%q) = %v, want %v (msg: %s)", tt.branch, result.Passed, tt.want, result.Message)
		}
	}
}

func TestCheckPRDescription(t *testing.T) {
	tests := []struct {
		body string
		want bool
	}{
		{"This is a short desc", false},                       // under 30 chars
		{"", false},                                           // empty
		{"This PR fixes the login bug by updating the auth middleware.", true}, // long enough
	}

	for _, tt := range tests {
		info := &PRInfo{Body: tt.body}
		result := checkPRDescription(info)
		if result.Passed != tt.want {
			t.Errorf("checkPRDescription(%q) = %v, want %v", tt.body, result.Passed, tt.want)
		}
	}
}

func TestCheckIssueLinked(t *testing.T) {
	tests := []struct {
		body string
		want bool
	}{
		{"Fixes #42", true},
		{"closes #100", true},
		{"Resolves #7", true},
		{"This PR does stuff", false},
		{"See issue 42", false}, // no # prefix
		{"References #55", true},
	}

	for _, tt := range tests {
		info := &PRInfo{Body: tt.body}
		result := checkIssueLinked(info)
		if result.Passed != tt.want {
			t.Errorf("checkIssueLinked(%q) = %v, want %v", tt.body, result.Passed, tt.want)
		}
	}
}

func TestCheckREADME(t *testing.T) {
	tests := []struct {
		files []string
		want  bool
	}{
		{[]string{"README.md", "main.go"}, true},
		{[]string{"readme.md"}, true}, // case-insensitive
		{[]string{"main.go", "utils.go"}, false},
		{[]string{}, false},
	}

	for _, tt := range tests {
		info := &PRInfo{ChangedFiles: tt.files}
		result := checkREADME(info)
		if result.Passed != tt.want {
			t.Errorf("checkREADME(%v) = %v, want %v", tt.files, result.Passed, tt.want)
		}
	}
}

func TestCheckChangelog(t *testing.T) {
	tests := []struct {
		files []string
		want  bool
	}{
		{[]string{"CHANGELOG.md", "main.go"}, true},
		{[]string{"changelog.md"}, true},
		{[]string{"main.go"}, false},
	}

	for _, tt := range tests {
		info := &PRInfo{ChangedFiles: tt.files}
		result := checkChangelog(info)
		if result.Passed != tt.want {
			t.Errorf("checkChangelog(%v) = %v, want %v", tt.files, result.Passed, tt.want)
		}
	}
}

func TestFormatComment(t *testing.T) {
	results := []CheckResult{
		{Name: "README Updated", Passed: true, Message: "Updated.", Icon: "✓"},
		{Name: "Issue Linked", Passed: false, Message: "No issue ref.", Icon: "✗"},
	}
	comment := FormatComment(results)
	if comment == "" {
		t.Error("FormatComment returned empty string")
	}
	if len(comment) < 50 {
		t.Error("FormatComment output too short, likely broken")
	}
}
