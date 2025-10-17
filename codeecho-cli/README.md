# 📢 CodeEcho

_"Let your code speak back."_

CodeEcho is an open-source CLI tool that scans your repository and packages it into a single AI-friendly file. Perfect for feeding into ChatGPT, Claude, or any LLM.

Transform your entire codebase into structured formats (XML, JSON, or Markdown) that AI models can easily consume for analysis, documentation, and context generation.

---

## Features

- **Repository Scanning**: Extract file structure and content from any directory
- **Multiple Output Formats**: XML, JSON, and Markdown support
- **Streaming Architecture**: Process large repositories efficiently without loading everything into memory
- **Git Awareness**: Automatically respects `.gitignore` and captures Git metadata (branch, commits, author)
- **File Processing**: Remove comments, compress code, strip empty lines
- **Smart Filtering**: Include/exclude files and directories based on patterns
- **Progress Tracking**: Real-time feedback with verbose and quiet modes
- **Comprehensive Documentation Generation**: Auto-generate README, API docs, and project overviews
- **Cross-Platform**: Works on Linux, macOS, and Windows
- **Language Detection**: Automatic language identification with content-based analysis
- **Error Resilience**: Graceful error handling with detailed reporting

---

## Installation

### Download Binary

Download the latest release for your platform from the [releases page](https://github.com/opskraken/code-echo/releases).

### Build from Source

```bash
git clone https://github.com/opskraken/code-echo.git
cd code-echo/codeecho-cli
go build -o codeecho main.go
```

### Install to System PATH (Optional)

```bash
# Linux/macOS
sudo mv codeecho /usr/local/bin/

# Windows
# Move codeecho.exe to a directory in your PATH
```

---

## Quick Start

```bash
# Basic repository scan (generates XML file)
codeecho scan .

# Scan with comment removal and code compression
codeecho scan . --remove-comments --compress-code

# Generate JSON output
codeecho scan . --format json

# Show detailed progress for each file
codeecho scan . --verbose

# Suppress progress output
codeecho scan . --quiet

# Auto-generate project README
codeecho doc . --type readme

# Show version information
codeecho version
```

---

## Commands

### `scan` - Repository Scanning

Scan a repository and generate AI-ready context files.

```bash
codeecho scan [path] [flags]
```

#### Output Format Flags

| Flag             | Type   | Default        | Description                        |
| ---------------- | ------ | -------------- | ---------------------------------- |
| `--format, -f`   | string | `xml`          | Output format: xml, json, markdown |
| `--out, -o`      | string | auto-generated | Output file path                   |
| `--include-tree` | bool   | `true`         | Include directory structure        |
| `--line-numbers` | bool   | `false`        | Show line numbers in code blocks   |

#### File Processing Flags

| Flag                   | Type | Default | Description                      |
| ---------------------- | ---- | ------- | -------------------------------- |
| `--compress-code`      | bool | `false` | Remove unnecessary whitespace    |
| `--remove-comments`    | bool | `false` | Strip comments from source files |
| `--remove-empty-lines` | bool | `false` | Remove blank lines               |

#### Git Awareness Flags

| Flag             | Type | Default | Description                                     |
| ---------------- | ---- | ------- | ----------------------------------------------- |
| `--git-aware`    | bool | `true`  | Enable git-aware scanning (respects .gitignore) |
| `--no-git-aware` | bool | `false` | Disable all git integration                     |
| `--git-timeout`  | int  | `5`     | Timeout for git commands (seconds)              |

#### File Filtering Flags

| Flag             | Type    | Default   | Description                                                   |
| ---------------- | ------- | --------- | ------------------------------------------------------------- |
| `--content`      | bool    | `true`    | Include file contents (use `--no-content` for structure only) |
| `--exclude-dirs` | strings | See below | Directories to exclude                                        |
| `--include-exts` | strings | See below | File extensions to include                                    |

#### Progress & Output Flags

| Flag                | Type | Default | Description                                    |
| ------------------- | ---- | ------- | ---------------------------------------------- |
| `--verbose, -v`     | bool | `false` | Show detailed progress for each file processed |
| `--quiet, -q`       | bool | `false` | Suppress all progress output                   |
| `--strict`          | bool | `false` | Fail immediately on any error                  |
| `--include-summary` | bool | `true`  | Include file summary section in output         |

**Default Excluded Directories:**
`.git`, `node_modules`, `vendor`, `.vscode`, `.idea`, `target`, `build`, `dist`

**Default Included Extensions:**
`.go`, `.js`, `.ts`, `.jsx`, `.tsx`, `.json`, `.md`, `.html`, `.css`, `.py`, `.java`, `.cpp`, `.c`, `.h`, `.rs`, `.rb`, `.php`, `.yml`, `.yaml`, `.toml`, `.xml`

#### Scan Examples

```bash
# Scan current repo into default XML
codeecho scan .

# JSON structure only (no file contents)
codeecho scan . --format json --no-content

# Markdown with line numbers
codeecho scan . --format markdown --line-numbers

# Exclude dirs and compress code
codeecho scan . --exclude-dirs .git,node_modules --compress-code

# Include only Go + Python files
codeecho scan . --include-exts .go,.py

# Verbose scanning with detailed progress
codeecho scan . --verbose

# Disable git awareness
codeecho scan . --no-git-aware

# Increase git timeout for slow systems
codeecho scan . --git-timeout 10

# Silent scan with error reporting only
codeecho scan . --quiet --strict
```

---

### `doc` - Documentation Generation

Generate documentation from repository analysis.

```bash
codeecho doc [path] [flags]
```

#### Documentation Flags

| Flag            | Type   | Default        | Description                               |
| --------------- | ------ | -------------- | ----------------------------------------- |
| `--out, -o`     | string | auto-generated | Output file path                          |
| `--type, -t`    | string | `readme`       | Documentation type: readme, api, overview |
| `--verbose, -v` | bool   | `false`        | Show detailed progress information        |
| `--quiet, -q`   | bool   | `false`        | Suppress progress output                  |

The `doc` command supports three documentation types:

- **readme**: Generates a comprehensive README with project overview, tech stack, structure, and getting started guide
- **api**: Creates API documentation for projects with route handlers and endpoints
- **overview**: Produces a high-level project overview with statistics and file distribution

#### Doc Examples

```bash
# Generate README for current directory
codeecho doc .

# Generate API documentation
codeecho doc . --type api

# Generate overview with custom output file
codeecho doc . --type overview -o OVERVIEW.md

# Show progress for each analyzed file
codeecho doc . --verbose
```

---

### `version` - Version Information

Display version and build information.

```bash
codeecho version
```

---

## Output Files

### Auto-Generated Filenames

When no `--out` flag is specified, files are automatically named using the pattern:

```
{project-name}-{processing-options}-{timestamp}.{extension}
```

**Examples:**

- `my-project-20250128-143022.xml` - Basic scan
- `my-project-no-comments-compressed-20250128-143025.xml` - Processed scan
- `my-project-structure-only-20250128-143028.xml` - Structure-only scan
- `my-project-20250128-143030.json` - JSON format

### Output Formats

#### XML Format (Default)

Structured XML similar to Repomix format, optimized for AI consumption. Includes:

- File metadata (size, language, modification time)
- Directory structure
- File contents (with optional line numbers)
- Scan statistics

#### JSON Format

Machine-readable JSON with complete file metadata and content. Suitable for programmatic processing and analysis.

#### Markdown Format

Human-readable documentation with syntax highlighting and organized sections. Perfect for documentation sites and reviews.

---

## Use Cases

### AI Context Generation

```bash
# Generate comprehensive context for AI tools
codeecho scan . --remove-comments --compress-code -o project-context.xml
```

### Code Review Preparation

```bash
# Clean, compressed version for review
codeecho scan . --remove-comments --remove-empty-lines --compress-code
```

### Project Analysis

```bash
# Structure and statistics only
codeecho scan . --no-content --format json -o project-analysis.json
```

### Automated Documentation

```bash
# Auto-generate project README
codeecho doc . --type readme

# Auto-generate API documentation
codeecho doc . --type api
```

---

## Configuration

### Custom File Extensions

```bash
# Include specific file types
codeecho scan . --include-exts .go,.py,.js,.md

# Include all files (remove default filtering)
codeecho scan . --include-exts ""
```

### Custom Directory Exclusions

```bash
# Exclude additional directories
codeecho scan . --exclude-dirs .git,node_modules,build,dist,tmp
```

---

## System Requirements

- **No dependencies**: Single binary with everything included
- **Cross-platform**: Linux, macOS, Windows support
- **Permissions**: Requires read access to target directories
- **Memory**: Minimal memory usage, processes files individually with streaming

---

## Common Issues

### Permission Denied

```bash
# Make binary executable (Linux/macOS)
chmod +x codeecho
```

### Large Repositories

```bash
# For very large repos, exclude build directories
codeecho scan . --exclude-dirs .git,node_modules,target,build,dist,vendor

# Use quiet mode to reduce terminal output
codeecho scan . --quiet
```

### Binary Files

Binary files are automatically excluded from scans. Only text files matching the included extensions are processed. The tool uses intelligent detection combining file extensions, filenames, shebangs, and content analysis.

### Permission or Read Errors

By default, CodeEcho continues scanning even when it encounters read errors or permission issues. Use `--strict` mode to fail immediately on any error and see detailed error reporting.

---

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

---

## Support

For issues, questions, or contributions:

- **GitHub Issues**: [Report bugs or request features](https://github.com/opskraken/code-echo/issues)
- **Discussions**: [Community discussions](https://github.com/opskraken/code-echo/discussions)

---

> **CodeEcho CLI – Making your repositories AI-ready**
