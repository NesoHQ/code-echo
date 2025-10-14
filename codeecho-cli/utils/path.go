package utils

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/NesoHQ/code-echo/codeecho-cli/types"
)

func GetRelativePath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}

func GenerateAutoFilename(repoPath, format string, opts types.OutputOptions) string {
	// Get project name
	projectName := filepath.Base(repoPath)
	if projectName == "." || projectName == "/" {
		projectName = "codeecho-scan"
	}

	// Add timestamp for uniqueness
	timestamp := time.Now().Format("20060102-150405")

	// Determine file extension
	var ext string
	switch format {
	case "json":
		ext = ".json"
	case "markdown", "md":
		ext = ".md"
	default:
		ext = ".xml"
	}

	// Build filename with processing indicators
	var suffix []string
	if opts.RemoveComments {
		suffix = append(suffix, "no-comments")
	}
	if opts.RemoveEmptyLines {
		suffix = append(suffix, "no-empty-lines")
	}
	if opts.CompressCode {
		suffix = append(suffix, "compressed")
	}
	if !opts.IncludeContent {
		suffix = append(suffix, "structure-only")
	}

	filename := projectName
	if len(suffix) > 0 {
		filename += "-" + strings.Join(suffix, "-")
	}
	filename += "-" + timestamp + ext

	return filename
}
