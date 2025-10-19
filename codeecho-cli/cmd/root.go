// cmd/root.go
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "codeecho",
	Short: "CodeEcho - Make your repository AI-ready",
	Long: `CodeEcho is a CLI tool that scans repositories and generates AI-ready context.
It converts your entire codebase into structured formats (JSON, Markdown) that can
be easily consumed by AI tools like ChatGPT, Claude, or Gemini.

Perfect for:
• Generating documentation automatically
• Creating context for AI-assisted coding
• Repository analysis and insights
• Code reviews and refactoring guidance`,
	Version: "1.0.0-beta",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Global flags can be added here
	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.codeecho.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")
}
