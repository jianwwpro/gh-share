package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	repoFlag string
	dirFlag  string
)

var rootCmd = &cobra.Command{
	Use:   "gh-share",
	Short: "Share static pages instantly via GitHub Pages (<user>/share)",
	Long: `gh-share is a GitHub CLI extension for zero-friction sharing of AI/agent
conversations, ideas, code snippets, and media via static GitHub Pages.

Pages are deployed instantly to https://<owner>.github.io/share/s/<slug>/`,
	SilenceUsage: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&repoFlag, "repo", "R", "", "Target repository [owner/]repo (defaults to <current-user>/share)")
	rootCmd.PersistentFlags().StringVarP(&dirFlag, "dir", "d", "", "Path to local share git checkout (defaults to ~/Projects/share)")
}
