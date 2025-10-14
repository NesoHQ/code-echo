package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/NesoHQ/code-echo/codeecho-cli/output"
	"github.com/NesoHQ/code-echo/codeecho-cli/scanner"
	"github.com/NesoHQ/code-echo/codeecho-cli/types"
	"github.com/NesoHQ/code-echo/codeecho-cli/utils"
	"github.com/spf13/cobra"
)

var (
	// Existing flags remain the same
	outputFormat         string
	outputFile           string
	includeSummary       bool
	includeDirectoryTree bool
	showLineNumbers      bool
	outputParsableFormat bool

	compressCode     bool
	removeComments   bool
	removeEmptyLines bool

	excludeDirs    []string
	includeExts    []string
	includeContent bool
	excludeContent bool

	// NEW: Progress and error flags
	verbose    bool // Show detailed progress
	quiet      bool // Suppress progress output
	strictMode bool // Fail on any error
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan repository and generate AI-ready context",
	Long: `Scan a repository and generate structured output for AI consumption.
Similar to Repomix, this command creates a single file containing your entire
codebase in a format optimized for AI tools.

Output Formats:
  xml        - Structured XML format (recommended for AI)
  json       - JSON format for programmatic use
  markdown   - Human-readable markdown format

Examples:
  codeecho scan .                              # Basic XML scan
  codeecho scan . --format json               # JSON output
  codeecho scan . --remove-comments           # Strip comments
  codeecho scan . --compress-code             # Minify code
  codeecho scan . --no-summary                # Skip file summary
  codeecho scan . --output packed-repo.xml    # Save to file
  codeecho scan . --verbose                   # Show detailed progress
  codeecho scan . --strict                    # Fail on any error`,
	Args: cobra.MaximumNArgs(1),
	RunE: runScan,
}

func init() {
	rootCmd.AddCommand(scanCmd)

	// Existing flags remain...
	scanCmd.Flags().StringVarP(&outputFormat, "format", "f", "xml", "Output format: xml, json, markdown")
	scanCmd.Flags().StringVarP(&outputFile, "output", "o", "", "Output file (default: auto-generated)")
	scanCmd.Flags().BoolVar(&includeSummary, "include-summary", true, "Include file summary section")
	scanCmd.Flags().BoolVar(&includeDirectoryTree, "include-tree", true, "Include directory structure")
	scanCmd.Flags().BoolVar(&showLineNumbers, "line-numbers", false, "Show line numbers in code blocks")
	scanCmd.Flags().BoolVar(&outputParsableFormat, "parsable", true, "Use parsable format tags")

	scanCmd.Flags().BoolVar(&compressCode, "compress-code", false, "Remove unnecessary whitespace from code")
	scanCmd.Flags().BoolVar(&removeComments, "remove-comments", false, "Strip comments from source files")
	scanCmd.Flags().BoolVar(&removeEmptyLines, "remove-empty-lines", false, "Remove empty lines from files")

	scanCmd.Flags().BoolVar(&includeContent, "content", true, "Include file contents")
	scanCmd.Flags().BoolVar(&excludeContent, "no-content", false, "Exclude file contents (structure only)")
	scanCmd.Flags().StringSliceVar(&excludeDirs, "exclude-dirs",
		[]string{".git", "node_modules", "vendor", ".vscode", ".idea", "target", "build", "dist"},
		"Directories to exclude")
	scanCmd.Flags().StringSliceVar(&includeExts, "include-exts",
		[]string{".go", ".js", ".ts", ".jsx", ".tsx", ".json", ".md", ".html", ".css", ".py", ".java", ".cpp", ".c", ".h", ".rs", ".rb", ".php", ".yml", ".yaml", ".toml", ".xml"},
		"File extensions to include")

	// NEW: Progress and error handling flags
	scanCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed progress information")
	scanCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress progress output")
	scanCmd.Flags().BoolVar(&strictMode, "strict", false, "Fail immediately on any error")
}

func runScan(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// Determine target path
	targetPath := "."
	if len(args) > 0 {
		targetPath = args[0]
	}

	// Validate path exists
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return fmt.Errorf("path does not exist: %s", targetPath)
	}

	// Get absolute path for cleaner output
	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	if !quiet {
		fmt.Printf("🔍 Scanning repository at %s...\n", absPath)
	}

	if excludeContent {
		includeContent = false
	}

	if compressCode || removeComments || removeEmptyLines {
		if !quiet {
			fmt.Println("⚙️  File processing enabled:")
			if compressCode {
				fmt.Println("    • Code compression")
			}
			if removeComments {
				fmt.Println("    • Comment removal")
			}
			if removeEmptyLines {
				fmt.Println("    • Empty line removal")
			}
		}
	}

	// Determine output file
	var outputFilePath string
	if outputFile != "" {
		outputFilePath = outputFile
	} else {
		outputOpts := types.OutputOptions{
			IncludeSummary:       includeSummary,
			IncludeDirectoryTree: includeDirectoryTree,
			ShowLineNumbers:      showLineNumbers,
			IncludeContent:       includeContent,
			RemoveComments:       removeComments,
			RemoveEmptyLines:     removeEmptyLines,
			CompressCode:         compressCode,
		}
		outputFilePath = utils.GenerateAutoFilename(absPath, outputFormat, outputOpts)
	}

	// Create output file
	outFile, err := os.Create(outputFilePath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outFile.Close()

	// Create output options
	outputOpts := types.OutputOptions{
		IncludeSummary:       includeSummary,
		IncludeDirectoryTree: includeDirectoryTree,
		ShowLineNumbers:      showLineNumbers,
		IncludeContent:       includeContent,
		RemoveComments:       removeComments,
		RemoveEmptyLines:     removeEmptyLines,
		CompressCode:         compressCode,
	}

	// Create streaming writer based on format
	writer, err := output.NewStreamingWriter(outFile, outputFormat, outputOpts)
	if err != nil {
		return err
	}
	defer writer.Close()

	// Write header
	scanTime := time.Now().Format(time.RFC3339)
	if err := writer.WriteHeader(absPath, scanTime); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Create scanner with streaming handler
	scanOpts := scanner.ScanOptions{
		IncludeSummary:       includeSummary,
		IncludeDirectoryTree: includeDirectoryTree,
		ShowLineNumbers:      showLineNumbers,
		OutputParsableFormat: outputParsableFormat,
		CompressCode:         compressCode,
		RemoveComments:       removeComments,
		RemoveEmptyLines:     removeEmptyLines,
		ExcludeDirs:          excludeDirs,
		IncludeExts:          includeExts,
		IncludeContent:       includeContent,
	}

	streamingScanner := scanner.NewStreamingScanner(absPath, scanOpts, writer.WriteFile)
	streamingScanner.SetTreeWriter(writer.WriteTree)

	// NEW: Setup progress tracking
	if !quiet {
		streamingScanner.SetProgressCallback(createProgressDisplay(verbose))
	}

	// Perform the scan
	if !quiet {
		fmt.Println("📊 Streaming scan in progress...")
	}

	stats, err := streamingScanner.Scan()

	// NEW: Check for errors in strict mode
	scanErrors := streamingScanner.GetErrors()
	if strictMode && len(scanErrors) > 0 {
		return fmt.Errorf("scan failed in strict mode: %d errors encountered", len(scanErrors))
	}

	if err != nil {
		return fmt.Errorf("scan failed: %w", err)
	}

	// Write footer with final statistics
	if err := writer.WriteFooter(stats); err != nil {
		return fmt.Errorf("failed to write footer: %w", err)
	}

	duration := time.Since(startTime)

	// Clear progress line
	if !quiet && !verbose {
		fmt.Print("\r\033[K") // Clear current line
	}

	// NEW: Display comprehensive summary
	displayScanSummary(outputFilePath, stats, scanErrors, duration)

	return nil
}

// NEW: Create progress display function
// Why: Centralized progress handling with verbose/quiet modes
func createProgressDisplay(verbose bool) scanner.ProgressCallback {
	var lastUpdate time.Time
	startTime := time.Now()

	return func(progress scanner.ScanProgress) {
		// Throttle updates to avoid terminal spam
		// Why: Updating too fast causes flickering
		now := time.Now()
		if now.Sub(lastUpdate) < 100*time.Millisecond && progress.Percentage < 100 {
			return
		}
		lastUpdate = now

		if verbose {
			// Verbose mode: Show every file
			elapsed := time.Since(startTime)
			eta := utils.EstimateTimeRemaining(progress.ProcessedFiles, progress.TotalFiles, elapsed)

			fmt.Printf("  [%s] %s - %s (ETA: %s)\n",
				progress.Phase,
				progress.CurrentFile,
				utils.CreateProgressBar(progress.ProcessedFiles, progress.TotalFiles, 20),
				eta,
			)
		} else {
			// Normal mode: Single updating line
			bar := utils.CreateProgressBar(progress.ProcessedFiles, progress.TotalFiles, 30)

			// Truncate filename if too long
			displayFile := progress.CurrentFile
			if len(displayFile) > 40 {
				displayFile = "..." + displayFile[len(displayFile)-37:]
			}

			fmt.Printf("\r  %s %s", bar, displayFile)
		}
	}
}

// NEW: Display comprehensive scan summary
// Why: Users need to see what happened - success, warnings, errors
func displayScanSummary(outputPath string, stats *scanner.StreamingStats, errors []scanner.ScanError, duration time.Duration) {
	fmt.Printf("\n✅ Output written to %s\n", outputPath)

	fmt.Printf("\n📈 Scan Summary:\n")
	fmt.Printf("  ├─ Files processed: %d\n", stats.TotalFiles)
	fmt.Printf("  ├─ Total size: %s\n", utils.FormatBytes(stats.TotalSize))
	fmt.Printf("  ├─ Text files: %d\n", stats.TextFiles)
	fmt.Printf("  ├─ Binary files: %d\n", stats.BinaryFiles)
	fmt.Printf("  └─ Duration: %s\n", utils.FormatDuration(duration))

	// Show language breakdown
	if len(stats.LanguageCounts) > 0 {
		fmt.Printf("\n💻 Languages detected:\n")

		// Sort languages by count
		type langCount struct {
			lang  string
			count int
		}
		var langs []langCount
		for lang, count := range stats.LanguageCounts {
			langs = append(langs, langCount{lang, count})
		}

		// Simple bubble sort (good enough for small lists)
		for i := 0; i < len(langs); i++ {
			for j := i + 1; j < len(langs); j++ {
				if langs[j].count > langs[i].count {
					langs[i], langs[j] = langs[j], langs[i]
				}
			}
		}

		// Show top 10
		maxShow := 10
		if len(langs) < maxShow {
			maxShow = len(langs)
		}

		for i := 0; i < maxShow; i++ {
			prefix := "├─"
			if i == maxShow-1 && len(errors) == 0 {
				prefix = "└─"
			}
			percentage := float64(langs[i].count) / float64(stats.TotalFiles) * 100
			fmt.Printf("  %s %s: %d files (%.1f%%)\n", prefix, langs[i].lang, langs[i].count, percentage)
		}

		if len(langs) > maxShow {
			fmt.Printf("  └─ ... and %d more\n", len(langs)-maxShow)
		}
	}

	// NEW: Display errors if any
	if len(errors) > 0 {
		fmt.Printf("\n⚠️  Warnings/Errors: %d issues encountered\n", len(errors))

		// Categorize errors
		readErrors := 0
		permissionErrors := 0
		otherErrors := 0

		for _, err := range errors {
			if err.Phase == "read" {
				readErrors++
			} else if err.Phase == "scan" && err.Error != nil {
				// Check if it's a permission error
				if os.IsPermission(err.Error) {
					permissionErrors++
				} else {
					otherErrors++
				}
			} else {
				otherErrors++
			}
		}

		if readErrors > 0 {
			fmt.Printf("  ├─ Read errors: %d files couldn't be read\n", readErrors)
		}
		if permissionErrors > 0 {
			fmt.Printf("  ├─ Permission denied: %d files\n", permissionErrors)
		}
		if otherErrors > 0 {
			fmt.Printf("  └─ Other errors: %d\n", otherErrors)
		}

		// Show first few errors if verbose or if there are only a few
		if verbose || len(errors) <= 5 {
			fmt.Printf("\n📝 Error details:\n")
			maxErrors := 10
			if len(errors) < maxErrors {
				maxErrors = len(errors)
			}

			for i := 0; i < maxErrors; i++ {
				prefix := "├─"
				if i == maxErrors-1 {
					prefix = "└─"
				}
				fmt.Printf("  %s %s: %v\n", prefix, errors[i].Path, errors[i].Error)
			}

			if len(errors) > maxErrors {
				fmt.Printf("  └─ ... and %d more errors (use --verbose to see all)\n", len(errors)-maxErrors)
			}
		} else {
			fmt.Printf("  💡 Use --verbose to see error details\n")
		}
	}

	fmt.Println() // Empty line for spacing
}
