package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/NesoHQ/code-echo/codeecho-cli/scanner"
	"gopkg.in/yaml.v3"
)

type ConfigFile struct {
	// Scanning options
	Format          string   `yaml:"format" json:"format"`
	ExcludeDirs     []string `yaml:"exclude_dirs" json:"exclude_dirs"`
	IncludeExts     []string `yaml:"include_exts" json:"include_exts"`
	IncludeContent  bool     `yaml:"include_content" json:"include_content"`
	IncludeSummary  bool     `yaml:"include_summary" json:"include_summary"`
	IncludeTree     bool     `yaml:"include_tree" json:"include_tree"`
	ShowLineNumbers bool     `yaml:"show_line_numbers" json:"show_line_numbers"`

	// Processing options
	CompressCode     bool `yaml:"compress_code" json:"compress_code"`
	RemoveComments   bool `yaml:"remove_comments" json:"remove_comments"`
	RemoveEmptyLines bool `yaml:"remove_empty_lines" json:"remove_empty_lines"`

	// Output options
	Output        string `yaml:"output" json:"output"`
	OutputQuiet   bool   `yaml:"quiet" json:"quiet"`
	OutputVerbose bool   `yaml:"verbose" json:"verbose"`

	// Presets (for future use)
	Preset string `yaml:"preset" json:"preset"`
}

// FindConfigFile looks for .codeecho.yaml or .codeecho.json in the current directory
// and up to the root of the repo
// Why: Many projects store config at repo root, but user may run from subdirectory

func FindConfigFile(startPath string) (string, error) {
	// Normalize path
	if startPath == "" {
		startPath = "."
	}

	absPath, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	// Check if startPath is a file or directory
	info, err := os.Stat(absPath)
	if err != nil {
		return "", err
	}

	// if it's a file, start from its directory
	if !info.IsDir() {
		absPath = filepath.Dir(absPath)
	}

	// Walk up directory tree looking for config
	// Why limit? Prevent infinite loops and excessive searching
	maxLevels := 10
	currentPath := absPath

	for i := 0; i < maxLevels; i++ {
		// Check for YAML first (preferred)
		yamlPath := filepath.Join(currentPath, ".codeecho.yaml")
		if _, err := os.Stat(yamlPath); err == nil {
			return yamlPath, nil
		}

		// Check for JSON as fallback
		jsonPath := filepath.Join(currentPath, ".codeecho.json")
		if _, err := os.Stat(jsonPath); err == nil {
			return jsonPath, nil
		}

		// Move to parent directory
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			// Reached root directory
			break
		}
		currentPath = parentPath
	}
	return "", nil // No config found (not an error)
}

// LoadConfigFile reads and parses a .codeecho.yaml or .codeecho.json file
// Why separate function? Makes it testable and reusable
func LoadConfigFile(filePath string) (*ConfigFile, error) {
	if filePath == "" {
		return nil, fmt.Errorf("config file path is empty")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := &ConfigFile{}

	// Determine file type by extension
	ext := filepath.Ext(filePath)
	switch ext {
	case ".yaml", ".yml":
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse YAML config: %w", err)
		}
	case ".json":
		// JSON is handled by Go's encoding/json with YAML tags
		// This works because YAML is a superset of JSON
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, fmt.Errorf("failed to parse JSON config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	return config, nil
}

// ApplyConfigToOptions merges ConfigFile values into ScanOptions
// Implements precedence: CLI flags > config file > defaults
// Why: Users can override config with flags
func ApplyConfigToOptions(configFile *ConfigFile, opts *scanner.ScanOptions,
	cliOverrides map[string]bool) {

	if configFile == nil {
		return
	}

	// Only apply config values if CLI flag wasn't explicitly set
	// We use a map to track which flags were provided on CLI

	if !cliOverrides["format"] && configFile.Format != "" {
		// Note: Format is handled separately in cmd/scan.go
		// This is just for reference
	}

	if !cliOverrides["exclude-dirs"] && len(configFile.ExcludeDirs) > 0 {
		opts.ExcludeDirs = configFile.ExcludeDirs
	}

	if !cliOverrides["include-exts"] && len(configFile.IncludeExts) > 0 {
		opts.IncludeExts = configFile.IncludeExts
	}

	if !cliOverrides["include-content"] && !configFile.IncludeContent {
		// Config explicitly says don't include content
		opts.IncludeContent = configFile.IncludeContent
	}

	if !cliOverrides["include-summary"] {
		opts.IncludeSummary = configFile.IncludeSummary
	}

	if !cliOverrides["include-tree"] {
		opts.IncludeDirectoryTree = configFile.IncludeTree
	}

	if !cliOverrides["show-line-numbers"] && configFile.ShowLineNumbers {
		opts.ShowLineNumbers = configFile.ShowLineNumbers
	}

	if !cliOverrides["compress-code"] && configFile.CompressCode {
		opts.CompressCode = configFile.CompressCode
	}

	if !cliOverrides["remove-comments"] && configFile.RemoveComments {
		opts.RemoveComments = configFile.RemoveComments
	}

	if !cliOverrides["remove-empty-lines"] && configFile.RemoveEmptyLines {
		opts.RemoveEmptyLines = configFile.RemoveEmptyLines
	}
}

// CreateDefaultConfigFile generates a template config file
// Use: `codeecho init` command (future feature)
func CreateDefaultConfigFile() string {
	return `# CodeEcho Configuration File
# Save as .codeecho.yaml in your project root

# Output format: xml, json, or markdown
format: xml

# File filtering
exclude_dirs:
  - .git
  - node_modules
  - vendor
  - dist
  - build
  - .vscode
  - .idea

include_exts:
  - .go
  - .js
  - .ts
  - .jsx
  - .tsx
  - .json
  - .md
  - .html
  - .css
  - .py

# Content options
include_content: true
include_summary: true
include_tree: true
show_line_numbers: false

# Processing options
compress_code: false
remove_comments: false
remove_empty_lines: false

# Output options
output: ""      # Leave empty for auto-generated filenames
quiet: false
verbose: false

# Preset profiles (future expansion)
# preset: "ai-optimized"  # minimal, comprehensive, ai-optimized, documentation
`
}
