# ✅ CodeEcho TODOS

A running list of upcoming tasks, improvements, and ideas for the CodeEcho CLI.

---

## ✅ Stage 1 – Core Scanning (DONE ✅)

- [x] Basic CLI with Cobra
- [x] `scan` command with XML/JSON/Markdown output
- [x] Exclude default directories (`.git`, `node_modules`, etc.)
- [x] Include/exclude file extensions
- [x] File processing options:
  - [x] Remove comments (language-aware)
  - [x] Strip empty lines
  - [x] Compress whitespace
- [x] Directory tree generation
- [x] `--no-content` flag (structure only)
- [x] Rich metadata in XML/JSON/Markdown
- [x] Auto-named output files with processing indicators
- [x] `version` command

---

## ✅ Stage 2 – Quality & Polish (DONE ✅)

- [x] Better language detection (extension + shebang + content-based)
- [x] Smarter binary/text detection (UTF-8 validation, null byte detection)
- [x] Resilient error handling (skip unreadable files gracefully)
- [x] Progress tracking with callbacks
- [x] Improved CLI output:
  - [x] Verbose mode (`--verbose, -v`)
  - [x] Quiet mode (`--quiet, -q`)
  - [x] Strict error handling mode (`--strict`)
  - [x] Real-time progress bars and ETA estimates
  - [x] Comprehensive scan summary with language breakdown
- [x] Streaming architecture for memory efficiency
- [x] Error collection and reporting
- [x] Code refactoring into smaller, testable components
- [x] Analysis scanner for doc command
- [x] Progress callback system

---

## 🔹 Stage 3 – Documentation Helpers (IN PROGRESS 🛠️)

- [x] `doc` command fully functional:
  - [x] Generate `README.md` with tech stack and getting started
  - [x] Generate `OVERVIEW.md` with statistics and file distribution
  - [x] Generate `API.md` for web projects
- [x] Dependency detection for common formats (`package.json`, `go.mod`, etc.)
- [x] Project statistics and language detection
- [x] Getting started instructions (auto-detection: Node.js, Go, Docker)
- [ ] Insert dependency summary into scan output
- [ ] Support project badges (language breakdown, file counts, etc.)

---

## 🔹 Stage 4 – Advanced Features (PLANNED 🚀)

### Configuration System

- [ ] `.codeecho.yaml` configuration file support
- [ ] Default settings per project
- [ ] Preset profiles (e.g., "minimal", "comprehensive", "ai-optimized")

### Output Enhancement

- [ ] Export to LLM-friendly formats:
  - [ ] OpenAI JSONL format
  - [ ] Anthropic XML format
  - [ ] Chunked output for large files
- [ ] Syntax highlighting in Markdown outputs
- [ ] ANSI color support in terminal output
- [ ] Progress bar customization

### Integration & Automation

- [ ] VSCode extension integration
- [ ] GitHub Action: auto-generate repo context file on push
- [ ] GitHub Action: Generate docs on new releases
- [ ] npm package (`@codeecho/cli`)
- [ ] Go package for programmatic usage

---

## 🔹 Stage 5 – Advanced Analysis (UPCOMING 📊)

- [ ] Code complexity metrics
- [ ] File size analysis and warnings
- [ ] Duplicate code detection
- [ ] Security/secret detection (patterns for API keys, tokens)
- [ ] Dependency graph analysis
- [ ] Architecture overview generation
- [ ] Test coverage reporting

---

## 🔹 Stage 6 – Performance & Optimization (FUTURE ⚡)

- [ ] Incremental scans (cache + only changed files)
- [ ] Parallel file processing
- [ ] File chunking for large repositories
- [ ] Compression support for output files
- [ ] Streaming large files directly to output
- [ ] Memory profiling and optimization

---

## 🔹 Stage 7 – Extended Features (FUTURE IDEAS 💡)

- [ ] Syntax-aware comment stripping (AST-based, safer than regex)
- [ ] Remote repo scanning (GitHub/GitLab API)
- [ ] Web UI dashboard
- [ ] History tracking and diff generation
- [ ] Collaboration mode (multi-user comments)
- [ ] Plugin system for custom processors
- [ ] Template system for custom doc generation
- [ ] AI-powered code summarization

---

## 📝 Current Status & Priorities

**Current Version**: v0.2.0 (in development)

**Recently Completed** (Latest Session):

- Verbose/quiet mode implementation
- Progress tracking improvements
- Error collection and reporting
- Code refactoring for maintainability
- Doc command with full functionality
- Analysis scanner

**Next Priorities** (Short-term):

1. Configuration file support (`.codeecho.yaml`)
2. Preset profiles for common use cases
3. GitHub Actions integration
4. Performance benchmarking
5. Comprehensive test suite

**Medium-term Goals**:

- VSCode extension
- NPM package release
- Secret detection
- Advanced analytics

---

## 📋 Development Notes

### Architecture

- **Streaming Scanner**: Processes files one-by-one without loading entire repo into memory
- **Analysis Scanner**: Full in-memory scan for doc generation with progress callbacks
- **Output Handlers**: Factory pattern with pluggable format writers (XML, JSON, Markdown)
- **Language Detection**: Multi-stage detection (extension → filename → shebang → content)

### Testing Recommendations

- [ ] Unit tests for language detection
- [ ] Integration tests for scan process
- [ ] Performance benchmarks for large repos
- [ ] Error handling edge cases

### Known Limitations

- Comment removal uses regex (not AST-based), so may have edge cases
- Large files (>100MB) not recommended without chunking
- Windows path handling may need additional testing

---

## 🐛 Bug Reports & Feedback

Feedback-driven development: issues/discussions welcome!

- Report bugs on GitHub Issues
- Suggest features in Discussions
- Questions? Check the README or open a Discussion
