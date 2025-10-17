package scanner

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/NesoHQ/code-echo/codeecho-cli/utils"
	ignore "github.com/sabhiram/go-gitignore"
)

type AnalysisScanner struct {
	rootPath string
	opts     ScanOptions

	// NEW: Progress and error tracking
	progressCallback ProgressCallback
	errors           []ScanError
	startTime        time.Time

	gitignore *ignore.GitIgnore
	gitMeta   *GitMetadata
}

func NewAnalysisScanner(rootPath string, opts ScanOptions) *AnalysisScanner {
	scanner := &AnalysisScanner{
		rootPath: rootPath,
		opts:     opts,
		errors:   []ScanError{},
	}

	// Load Git information if git-aware mode is enabled
	if opts.GitAware {
		// Load .gitignore patterns
		gitignore, err := LoadGitignorePatterns(rootPath)
		if err != nil {
			scanner.errors = append(scanner.errors, ScanError{
				Path:    filepath.Join(rootPath, ".gitignore"),
				Phase:   "gitignore",
				Error:   err,
				Skipped: false,
			})
		}
		scanner.gitignore = gitignore

		// Load Git metadata
		gitMeta, gitErrors := LoadGitMetadata(rootPath)
		scanner.gitMeta = gitMeta

		for _, err := range gitErrors {
			scanner.errors = append(scanner.errors, ScanError{
				Path:    rootPath,
				Phase:   "git-metadata",
				Error:   err,
				Skipped: false,
			})
		}
	}

	return scanner
}

// NEW: Set progress callback
func (a *AnalysisScanner) SetProgressCallback(callback ProgressCallback) {
	a.progressCallback = callback
}

// NEW: Get collected errors
func (a *AnalysisScanner) GetErrors() []ScanError {
	return a.errors
}

// NEW: Report progress
func (a *AnalysisScanner) reportProgress(phase string, currentFile string, processed, total int) {
	if a.progressCallback == nil {
		return
	}

	progress := ScanProgress{
		Phase:          phase,
		CurrentFile:    currentFile,
		ProcessedFiles: processed,
		TotalFiles:     total,
	}

	if total > 0 {
		progress.Percentage = float64(processed) / float64(total) * 100
	}

	a.progressCallback(progress)
}

// NEW: Record error
func (a *AnalysisScanner) recordError(path string, phase string, err error) {
	a.errors = append(a.errors, ScanError{
		Path:    path,
		Phase:   phase,
		Error:   err,
		Skipped: true,
	})
}

// Scan performs a full repository scan and returns complete results
// Unlike StreamingScanner, this keeps all data in memory
func (a *AnalysisScanner) Scan() (*ScanResult, error) {
	a.startTime = time.Now()

	result := &ScanResult{
		RepoPath:       a.rootPath,
		ScanTime:       time.Now().Format(time.RFC3339),
		Files:          []FileInfo{},
		ProcessedBy:    "CodeEcho CLI",
		LanguageCounts: make(map[string]int),
		Git:            a.gitMeta,
	}

	// First pass: Count total files
	a.reportProgress("counting", "calculating total files...", 0, 0)
	totalFiles := 0
	filepath.WalkDir(a.rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		// Skip excluded directories
		if d.IsDir() && shouldExcludeDir(d.Name(), a.opts.ExcludeDirs) {
			return filepath.SkipDir
		}
		// Check .gitignore if enabled
		if a.opts.GitAware && a.gitignore != nil {
			relativePath := utils.GetRelativePath(a.rootPath, path)
			if IsIgnoredByGitignore(relativePath, a.gitignore) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if !d.IsDir() && shouldIncludeFile(path, a.opts.IncludeExts) {
			totalFiles++
		}
		return nil
	})

	// Second pass: Process files
	processedFiles := 0
	err := filepath.WalkDir(a.rootPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			a.recordError(path, "scan", err)
			return nil // Continue
		}

		// Skip excluded directories
		if d.IsDir() && shouldExcludeDir(d.Name(), a.opts.ExcludeDirs) {
			return filepath.SkipDir
		}
		// Check .gitignore if enabled
		if a.opts.GitAware && a.gitignore != nil {
			relativePath := utils.GetRelativePath(a.rootPath, path)
			if IsIgnoredByGitignore(relativePath, a.gitignore) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		// Process files only
		if !d.IsDir() && shouldIncludeFile(path, a.opts.IncludeExts) {
			relativePath := utils.GetRelativePath(a.rootPath, path)
			a.reportProgress("scanning", relativePath, processedFiles, totalFiles)

			info, err := d.Info()
			if err != nil {
				a.recordError(path, "stat", err)
				return nil // Continue
			}

			language := detectLanguage(path)
			extension := filepath.Ext(path)

			fileInfo := FileInfo{
				Path:             path,
				RelativePath:     relativePath,
				Size:             info.Size(),
				SizeFormatted:    utils.FormatBytes(info.Size()),
				ModTime:          info.ModTime().Format(time.RFC3339),
				ModTimeFormatted: info.ModTime().Format("2006-01-02 15:04:05"),
				Language:         language,
				Extension:        extension,
				IsText:           isTextFile(path, extension),
			}

			// Include content if requested and it's a text file
			if a.opts.IncludeContent && fileInfo.IsText {
				content, err := os.ReadFile(path)
				if err != nil {
					a.recordError(path, "read", err)
				} else {
					// ENHANCED: Content-based detection
					if fileInfo.Language == "" {
						fileInfo.Language = detectLanguageFromContent(path, content)
					}
					if !fileInfo.IsText && isTextContent(content) {
						fileInfo.IsText = true
					}

					processedContent := processFileContent(string(content), fileInfo.Language, a.opts)
					fileInfo.Content = processedContent
					fileInfo.LineCount = utils.CountLines(processedContent)
				}
			}

			result.Files = append(result.Files, fileInfo)
			result.TotalFiles++
			result.TotalSize += info.Size()

			if fileInfo.IsText {
				result.TextFiles++
			} else {
				result.BinaryFiles++
			}

			if fileInfo.Language != "" {
				result.LanguageCounts[fileInfo.Language]++
			}
			processedFiles++
		}

		return nil
	})

	// Sort files by path for consistent output
	a.reportProgress("sorting", "organizing results...", totalFiles, totalFiles)
	sort.Slice(result.Files, func(i, j int) bool {
		return result.Files[i].RelativePath < result.Files[j].RelativePath
	})

	return result, err
}
