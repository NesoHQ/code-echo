package scanner

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	ignore "github.com/sabhiram/go-gitignore"
)

var GitCommandTimeout = 5 * time.Second

// GitMetadata contains Git repository information
type GitMetadata struct {
	Branch      string `json:"branch,omitempty"`
	CommitHash  string `json:"commit_hash,omitempty"`
	Author      string `json:"author,omitempty"`
	CommitDate  string `json:"commit_date,omitempty"`
	CommitCount int    `json:"commit_count,omitempty"`
}

// LoadGitMetadata extracts Git repository metadata
func LoadGitMetadata(repoPath string) (*GitMetadata, []error) {

	startTime := time.Now()
	defer func() {
		// Log timing if it takes more than 1 second
		duration := time.Since(startTime)
		if duration > time.Second {
			// TODO: use proper logging (In production)
			// For now, tracking it silently
			_ = duration
		}
	}()

	var errors []error

	// Check if .git directory exists
	gitDir := filepath.Join(repoPath, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		// Not a git repo - this is not an error
		return nil, nil
	}

	// Check if git command is available
	if _, err := exec.LookPath("git"); err != nil {
		errors = append(errors, fmt.Errorf("git command not found: %w", err))
		return nil, errors
	}

	metadata := &GitMetadata{}

	// Get current branch (handle detached HEAD)
	if branch, err := execGitCommand(repoPath, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		branch = sanitizeGitOutput(branch)
		if branch == "HEAD" {
			// Detached HEAD state - try to get commit hash instead
			if hash, hashErr := execGitCommand(repoPath, "rev-parse", "--short", "HEAD"); hashErr == nil {
				metadata.Branch = "detached@" + sanitizeGitOutput(hash)
			} else {
				metadata.Branch = "detached HEAD"
			}
		} else {
			metadata.Branch = branch
		}
	} else {
		errors = append(errors, fmt.Errorf("failed to get branch: %w", err))
	}

	// Get latest commit hash (short)
	if hash, err := execGitCommand(repoPath, "log", "-1", "--format=%h"); err == nil {
		metadata.CommitHash = sanitizeGitOutput(hash)
	} else {
		errors = append(errors, fmt.Errorf("failed to get commit hash: %w", err))
	}

	// Get author name
	if author, err := execGitCommand(repoPath, "log", "-1", "--format=%an"); err == nil {
		metadata.Author = sanitizeGitOutput(author)
	} else {
		errors = append(errors, fmt.Errorf("failed to get author: %w", err))
	}

	// Get commit date (ISO format)
	if date, err := execGitCommand(repoPath, "log", "-1", "--format=%ad", "--date=iso"); err == nil {
		metadata.CommitDate = sanitizeGitOutput(date)
	} else {
		errors = append(errors, fmt.Errorf("failed to get commit date: %w", err))
	}

	// Get commit count (may fail in shallow clones)
	if countStr, err := execGitCommand(repoPath, "rev-list", "--count", "HEAD"); err == nil {
		if count, parseErr := strconv.Atoi(countStr); parseErr == nil {
			metadata.CommitCount = count
		} else {
			errors = append(errors, fmt.Errorf("failed to parse commit count: %w", parseErr))
		}
	} else {
		// Shallow clones can't count commits - not critical
		// Set to -1 to indicate unknown
		metadata.CommitCount = -1
		errors = append(errors, fmt.Errorf("failed to get commit count (shallow clone?): %w", err))
	}

	// Return nil if we couldn't get any core metadata
	if metadata.Branch == "" && metadata.CommitHash == "" {
		return nil, errors
	}

	return metadata, errors
}

// sanitizeGitOutput cleans Git output to prevent injection attacks
func sanitizeGitOutput(s string) string {
	// Remove control characters except newline and tab
	s = strings.Map(func(r rune) rune {
		if r < 32 && r != '\n' && r != '\t' {
			return -1
		}
		return r
	}, s)

	// Limit length to prevent DoS
	if len(s) > 1000 {
		s = s[:1000]
	}

	return strings.TrimSpace(s)
}

func execGitCommand(repoPath string, args ...string) (string, error) {
	// Create context with 5-second timeout
	ctx, cancel := context.WithTimeout(context.Background(), GitCommandTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = repoPath

	output, err := cmd.Output()
	if err != nil {
		// Check if it was a timeout
		if ctx.Err() == context.DeadlineExceeded {
			return "", fmt.Errorf("git command timed out after 5s")
		}

		// Capture stderr for better error messages
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git command failed: %w (stderr: %s)",
				err, string(exitErr.Stderr))
		}
		return "", fmt.Errorf("git command failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// LoadGitignorePatterns loads .gitignore patterns
func LoadGitignorePatterns(repoPath string) (*ignore.GitIgnore, error) {
	gitignorePath := filepath.Join(repoPath, ".gitignore")

	// Check if .gitignore exists
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		// Not having .gitignore is not an error
		return nil, nil
	}

	// Parse .gitignore file
	gitignore, err := ignore.CompileIgnoreFile(gitignorePath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse .gitignore: %w", err)
	}

	return gitignore, nil
}

// IsIgnoredByGitignore checks if path matches .gitignore
func IsIgnoredByGitignore(path string, gitignore *ignore.GitIgnore) bool {
	if gitignore == nil {
		return false
	}
	return gitignore.MatchesPath(path)
}
