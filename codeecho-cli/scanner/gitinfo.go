package scanner

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	ignore "github.com/sabhiram/go-gitignore"
)

// GitMetadata contains Git repository information
type GitMetadata struct {
	Branch      string `json:"branch,omitempty"`
	CommitHash  string `json:"commit_hash,omitempty"`
	Author      string `json:"author,omitempty"`
	CommitDate  string `json:"commit_date,omitempty"`
	CommitCount int    `json:"commit_count,omitempty"`
}

// LoadGitMetadata extracts Git repository metadata
func LoadGitMetadata(repoPath string) *GitMetadata {
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		return nil
	}

	if _, err := exec.LookPath("git"); err != nil {
		return nil
	}

	metadata := &GitMetadata{}
	metadata.Branch = execGitCommand(repoPath, "rev-parse", "--abbrev-ref", "HEAD")
	metadata.CommitHash = execGitCommand(repoPath, "log", "-1", "--format=%h")
	metadata.Author = execGitCommand(repoPath, "log", "-1", "--format=%an")
	metadata.CommitDate = execGitCommand(repoPath, "log", "-1", "--format=%ad", "--date=iso")

	if countStr := execGitCommand(repoPath, "rev-list", "--count", "HEAD"); countStr != "" {
		if count, err := strconv.Atoi(countStr); err == nil {
			metadata.CommitCount = count
		}
	}

	if metadata.Branch == "" && metadata.CommitHash == "" {
		return nil
	}

	return metadata
}

func execGitCommand(repoPath string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// LoadGitignorePatterns loads .gitignore patterns
func LoadGitignorePatterns(repoPath string) *ignore.GitIgnore {
	gitignorePath := filepath.Join(repoPath, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		return nil
	}
	gitignore, err := ignore.CompileIgnoreFile(gitignorePath)
	if err != nil {
		return nil
	}
	return gitignore
}

// IsIgnoredByGitignore checks if path matches .gitignore
func IsIgnoredByGitignore(path string, gitignore *ignore.GitIgnore) bool {
	if gitignore == nil {
		return false
	}
	return gitignore.MatchesPath(path)
}
