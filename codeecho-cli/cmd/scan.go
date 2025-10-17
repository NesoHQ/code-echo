package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/NesoHQ/code-echo/codeecho-cli/config"
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

	verbose    bool
	quiet      bool
	strictMode bool

	configFile string
	gitAware   bool
	noGitAware bool
	gitTimeout int
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
	codeecho scan . --config /path/to/.codeecho.yaml
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

	// Progress and error handling flags
	scanCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show detailed progress information")
	scanCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Suppress progress output")
	scanCmd.Flags().BoolVar(&strictMode, "strict", false, "Fail immediately on any error")

	// NEW: Config file flag
	scanCmd.Flags().StringVar(&configFile, "config", "", "Path to .codeecho.yaml or .codeecho.json config file")

	scanCmd.Flags().BoolVar(&gitAware, "git-aware", true, "Enable git-aware scanning")
	scanCmd.Flags().BoolVar(&noGitAware, "no-git-aware", false, "Disable git integration")
	scanCmd.Flags().IntVar(&gitTimeout, "git-timeout", 5, "Timeout for git commands in seconds")
}

// NEW: Track which CLI flags were explicitly set
// Why: Distinguish between "user didn't set flag" vs "flag has default value"
// This allows config file to provide defaults while CLI overrides them
func getCliOverrides(cmd *cobra.Command) map[string]bool {
	overrides := make(map[string]bool)

	// Check which flags were explicitly set by user
	// We only check the important ones that might conflict with config
	if cmd.Flags().Changed("format") {
		overrides["format"] = true
	}
	if cmd.Flags().Changed("exclude-dirs") {
		overrides["exclude-dirs"] = true
	}
	if cmd.Flags().Changed("include-exts") {
		overrides["include-exts"] = true
	}
	if cmd.Flags().Changed("content") {
		overrides["include-content"] = true
	}
	if cmd.Flags().Changed("no-content") {
		overrides["no-content"] = true
	}
	if cmd.Flags().Changed("include-summary") {
		overrides["include-summary"] = true
	}
	if cmd.Flags().Changed("include-tree") {
		overrides["include-tree"] = true
	}
	if cmd.Flags().Changed("line-numbers") {
		overrides["show-line-numbers"] = true
	}
	if cmd.Flags().Changed("compress-code") {
		overrides["compress-code"] = true
	}
	if cmd.Flags().Changed("remove-comments") {
		overrides["remove-comments"] = true
	}
	if cmd.Flags().Changed("remove-empty-lines") {
		overrides["remove-empty-lines"] = true
	}

	return overrides
}

// NEW: Load and merge configuration
// Why: Centralize config logic, make it testable
func loadAndMergeConfig(targetPath string, cmd *cobra.Command) error {
	// Step 1: Determine which config file to load
	var configPath string
	var err error

	if configFile != "" {
		// User specified explicit config file
		configPath = configFile
	} else {
		// Auto-discover config file starting from targetPath
		configPath, err = config.FindConfigFile(targetPath)
		if err != nil {
			return fmt.Errorf("failed to search for config file: %w", err)
		}
	}

	// If no config found, that's OK - just use CLI flags
	if configPath == "" {
		if !quiet {
			// Mention that config could be used (informative, not an error)
			// Actually, don't spam - only show if verbose
			if verbose {
				fmt.Println("No .codeecho.yaml or .codeecho.json found, using CLI defaults")
			}
		}
		return nil
	}

	// Step 2: Load the config file
	if !quiet {
		fmt.Printf("⚙️  Loading config from %s\n", configPath)
	}

	cfg, err := config.LoadConfigFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to load config file: %w", err)
	}

	// Validate config
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	// Step 3: Determine which flags were explicitly set on CLI
	// This is crucial for proper precedence
	cliOverrides := getCliOverrides(cmd)

	// Step 4: Merge config into our current flag values
	// Why: Apply config defaults, but respect CLI overrides
	mergeConfigIntoFlags(cfg, cliOverrides)

	if !quiet && verbose {
		fmt.Println("✓ Config merged successfully (CLI flags take precedence)")
	}

	return nil
}

// NEW: Merge config file values into global flag variables
// Why: Modify the actual flags so rest of code sees merged values
func mergeConfigIntoFlags(cfg *config.ConfigFile, cliOverrides map[string]bool) {
	// Format: handled separately in runScan
	if !cliOverrides["format"] && cfg.Format != "" {
		outputFormat = cfg.Format
	}

	// Exclude dirs: merge if not overridden
	if !cliOverrides["exclude-dirs"] && len(cfg.ExcludeDirs) > 0 {
		excludeDirs = cfg.ExcludeDirs
	}

	// Include exts: merge if not overridden
	if !cliOverrides["include-exts"] && len(cfg.IncludeExts) > 0 {
		includeExts = cfg.IncludeExts
	}

	// Include content: respect config if not explicitly set
	if !cliOverrides["content"] && !cliOverrides["no-content"] {
		includeContent = cfg.IncludeContent
	}

	// Include summary
	if !cliOverrides["include-summary"] {
		includeSummary = cfg.IncludeSummary
	}

	// Include tree
	if !cliOverrides["include-tree"] {
		includeDirectoryTree = cfg.IncludeTree
	}

	// Show line numbers
	if !cliOverrides["show-line-numbers"] && cfg.ShowLineNumbers {
		showLineNumbers = cfg.ShowLineNumbers
	}

	// Processing options
	if !cliOverrides["compress-code"] && cfg.CompressCode {
		compressCode = cfg.CompressCode
	}

	if !cliOverrides["remove-comments"] && cfg.RemoveComments {
		removeComments = cfg.RemoveComments
	}

	if !cliOverrides["remove-empty-lines"] && cfg.RemoveEmptyLines {
		removeEmptyLines = cfg.RemoveEmptyLines
	}

	// Output file
	if outputFile == "" && cfg.Output != "" {
		outputFile = cfg.Output
	}

	// Progress flags
	if !cliOverrides["verbose"] && cfg.OutputVerbose {
		verbose = cfg.OutputVerbose
	}

	if !cliOverrides["quiet"] && cfg.OutputQuiet {
		quiet = cfg.OutputQuiet
	}

	// Git awareness
	if !cliOverrides["git-aware"] && !cliOverrides["no-git-aware"] {
		gitAware = cfg.GitAware
	}
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

	// NEW: Load config before proceeding with scan
	// Why: Do this early so all subsequent operations use merged config
	if err := loadAndMergeConfig(absPath, cmd); err != nil {
		// Config errors should be shown but not fatal (unless we want strict mode)
		if strictMode {
			return err
		}
		if !quiet {
			fmt.Printf("Warning: %v\n", err)
		}
	}

	if noGitAware {
		gitAware = false
	}

	// Set git timeout if specified
	if gitTimeout > 0 && gitTimeout != 5 {
		scanner.SetGitTimeout(time.Duration(gitTimeout) * time.Second)
	}

	if !quiet {
		fmt.Printf("🔍 Scanning repository at %s...\n", absPath)
		if gitAware {
			fmt.Println("⚙️  Git-aware mode enabled")
		}
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
		GitAware:             gitAware,
	}

	streamingScanner := scanner.NewStreamingScanner(absPath, scanOpts, writer.WriteFile)
	streamingScanner.SetTreeWriter(writer.WriteTree)
	// Get and display git info if available
	gitMeta := streamingScanner.GetGitMetadata()
	if gitAware && !quiet {
		if gitMeta != nil {
			commitCountStr := fmt.Sprintf("%d commits", gitMeta.CommitCount)
			if gitMeta.CommitCount == -1 {
				commitCountStr = "shallow clone"
			}
			fmt.Printf("✔ Detected Git branch: %s (%s)\n", gitMeta.Branch, commitCountStr)
		}

		// Check for .gitignore
		gitignorePath := filepath.Join(absPath, ".gitignore")
		if _, err := os.Stat(gitignorePath); err == nil {
			fmt.Println("✔ Loaded .gitignore rules")
		}

		// Show Git-related warnings if any
		gitErrors := 0
		for _, scanErr := range streamingScanner.GetErrors() {
			if scanErr.Phase == "git-metadata" || scanErr.Phase == "gitignore" {
				gitErrors++
			}
		}
		if gitErrors > 0 && verbose {
			fmt.Printf("⚠️  %d Git-related warnings (use --verbose for details)\n", gitErrors)
		}
	}

	// Write Git metadata to output
	if err := writer.WriteGitMetadata(gitMeta); err != nil {
		return fmt.Errorf("failed to write git metadata: %w", err)
	}
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
