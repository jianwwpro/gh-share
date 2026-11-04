package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuler/gh-share/internal/config"
	"github.com/yuler/gh-share/internal/git"
	"github.com/yuler/gh-share/internal/pages"
)

var appendCmd = &cobra.Command{
	Use:     "append <slug> [text]",
	Aliases: []string{"add"},
	Short:   "Append content to an existing share",
	Long: `Append new content to an existing share.

Examples:
  gh share append 2026-11-04-my-post "Additional notes..."
  echo "Update content" | gh share append 2026-11-04-my-post
  gh share append 2026-11-04-my-post update.md`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Resolve(repoFlag, dirFlag)
		if err != nil {
			return err
		}

		slug := strings.Trim(args[0], "/")
		targetDir := filepath.Join(cfg.LocalDir, "s", slug)
		indexFile := filepath.Join(targetDir, "index.html")

		// Check if the share exists
		if _, err := os.Stat(indexFile); os.IsNotExist(err) {
			return fmt.Errorf("share '%s' not found at %s", slug, targetDir)
		}

		fmt.Printf("➕ Appending to share '%s'...\n", slug)

		// Read existing HTML
		existingHTML, err := os.ReadFile(indexFile)
		if err != nil {
			return fmt.Errorf("failed to read existing share: %w", err)
		}

		// Extract existing content and metadata
		existingTitle := extractMetadataAppend(string(existingHTML), "title")
		existingLede := extractMetadataAppend(string(existingHTML), "lede")
		existingBody := extractBodyContentAppend(string(existingHTML))

		// Get new content to append
		var appendContent string
		if len(args) == 2 {
			input := args[1]
			if info, err := os.Stat(input); err == nil && !info.IsDir() {
				data, err := os.ReadFile(input)
				if err != nil {
					return fmt.Errorf("failed to read file %s: %w", input, err)
				}
				appendContent = string(data)
			} else {
				appendContent = input
			}
		} else {
			// Read from stdin
			stat, _ := os.Stdin.Stat()
			if (stat.Mode() & os.ModeCharDevice) == 0 {
				bytes, err := io.ReadAll(os.Stdin)
				if err != nil {
					return fmt.Errorf("failed to read from stdin: %w", err)
				}
				appendContent = string(bytes)
			} else {
				return fmt.Errorf("no content provided. Pass text, a file path, or pipe via stdin")
			}
		}

		// Combine existing and new content
		newContent := existingBody + "\n\n" + appendContent

		// Generate updated HTML
		htmlContent, err := pages.BuildHTML(newContent, pages.CreateOptions{
			Title: existingTitle,
			Slug:  slug,
			Type:  "note", // default type
			Theme: "paper", // default theme
			Lede:  existingLede,
			Tags:  nil, // keep existing tags
		}, cfg.BasePagesURL)
		if err != nil {
			return fmt.Errorf("failed to build HTML: %w", err)
		}

		// Write updated HTML
		if err := os.WriteFile(indexFile, []byte(htmlContent), 0644); err != nil {
			return fmt.Errorf("failed to write updated share: %w", err)
		}

		// Commit and push
		relPath := filepath.Join("s", slug)
		commitMsg := fmt.Sprintf("➕ append to share: %s", slug)
		if err := git.AddAndCommit(cfg.LocalDir, relPath, commitMsg); err != nil {
			return err
		}
		if err := git.Push(cfg.LocalDir); err != nil {
			return err
		}

		shareURL := fmt.Sprintf("%ss/%s/", strings.TrimRight(cfg.BasePagesURL, "/")+"/", slug)
		fmt.Printf("✅ Appended and pushed: %s\n", shareURL)

		return nil
	},
}

// Helper functions (different names to avoid conflict)
func extractMetadataAppend(html, field string) string {
	// Simplified extraction - in production use proper HTML parsing
	switch field {
	case "title":
		if idx := strings.Index(html, "<title>"); idx != -1 {
			endIdx := strings.Index(html[idx:], "</title>")
			if endIdx != -1 {
				return strings.TrimSpace(html[idx+7 : idx+endIdx])
			}
		}
	case "lede":
		if idx := strings.Index(html, "class=\"lede\""); idx != -1 {
			startIdx := strings.Index(html[idx:], ">")
			endIdx := strings.Index(html[idx+startIdx:], "</p>")
			if startIdx != -1 && endIdx != -1 {
				return strings.TrimSpace(html[idx+startIdx+1 : idx+startIdx+endIdx])
			}
		}
	}
	return ""
}

func extractBodyContentAppend(html string) string {
	// Simplified extraction
	startMarker := "</header>"
	endMarker := "<footer"
	
	startIdx := strings.Index(html, startMarker)
	if startIdx == -1 {
		return ""
	}
	startIdx += len(startMarker)
	
	endIdx := strings.Index(html[startIdx:], endMarker)
	if endIdx == -1 {
		endIdx = len(html) - startIdx
	}
	
	body := html[startIdx : startIdx+endIdx]
	body = strings.TrimSpace(body)
	return body
}

func init() {
	rootCmd.AddCommand(appendCmd)
}