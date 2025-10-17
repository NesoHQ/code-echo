# Changelog

## [Unreleased]

### Added

- **Git Awareness**: Automatic .gitignore support and Git metadata extraction
  - Respects .gitignore patterns during scanning
  - Captures branch name, commit hash, author, date, and commit count
  - CLI flags: `--git-aware`, `--no-git-aware`, `--git-timeout`
  - Config option: `gitAware: true`
  - Production-grade error handling with timeouts
  - Handles edge cases: detached HEAD, shallow clones, missing Git
  - Comprehensive test coverage

### Changed

- Git metadata now included in all output formats (XML, JSON, Markdown)
- Improved error reporting for Git-related operations
- Better handling of repositories without Git

### Fixed

- Git commands no longer hang on network filesystems (5s timeout)
- Proper error messages when Git is unavailable
- Sanitization of Git output to prevent injection attacks
