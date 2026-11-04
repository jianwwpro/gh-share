package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuler/gh-share/internal/config"
	"github.com/yuler/gh-share/internal/git"
	"github.com/yuler/gh-share/internal/pages"
)

var (
	updateTitle string
	updateLede  string
	updateType  string
	updateTheme string
	updateTags  []string
)

var updateCmd = &cobra.Command{
	Use:     "update <slug> [file]",
	Aliases: []string{"edit", "modify"},
	Short:   "Update an existing share",
	Long: `Update the content or metadata of an existing share.

Examples:
  gh share update 2026-11-04-my-post updated-content.md
  gh share update 2026-11-04-my-post --title "Updated Title"
  gh share update 2026-11-04-my-post --lede "New description"`,
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

		fmt.Printf("📝 Updating share '%s'...\n", slug)

		var content string

		// Read existing HTML to extract current metadata
		existingHTML, err := os.ReadFile(indexFile)
		if err != nil {
			return fmt.Errorf("failed to read existing share: %w", err)
		}

		// Parse existing metadata (simplified parsing)
		// In a real implementation, you'd want proper HTML parsing
		existingTitle := extractMetadata(string(existingHTML), "title")
		existingLede := extractMetadata(string(existingHTML), "lede")

		// Determine new content
		if len(args) == 2 {
			input := args[1]
			if info, err := os.Stat(input); err == nil && !info.IsDir() {
				data, err := os.ReadFile(input)
				if err != nil {
					return fmt.Errorf("failed to read file %s: %w", input, err)
				}
				content = string(data)
			} else {
				content = input
			}
		} else {
			// If no new content provided, keep existing content
			// Extract body content from existing HTML
			content = extractBodyContent(string(existingHTML))
		}

		// Use new values or fall back to existing ones
		title := updateTitle
		if title == "" {
			title = existingTitle
		}
		if title == "" {
			title = slug
		}

		lede := updateLede
		if lede == "" {
			lede = existingLede
		}

		postType := updateType
		if postType == "" {
			postType = "note" // default
		}

		theme := updateTheme
		if theme == "" {
			theme = "paper" // default
		}

		// Generate new HTML
		htmlContent, err := pages.BuildHTML(content, pages.CreateOptions{
			Title: title,
			Slug:  slug,
			Type:  postType,
			Theme: theme,
			Lede:  lede,
			Tags:  updateTags,
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
		commitMsg := fmt.Sprintf("📝 update share: %s", slug)
		if err := git.AddAndCommit(cfg.LocalDir, relPath, commitMsg); err != nil {
			return err
		}
		if err := git.Push(cfg.LocalDir); err != nil {
			return err
		}

		shareURL := fmt.Sprintf("%ss/%s/", strings.TrimRight(cfg.BasePagesURL, "/")+"/", slug)
		fmt.Printf("✅ Updated and pushed: %s\n", shareURL)

		return nil
	},
}

// extractMetadata extracts simple metadata from HTML
func extractMetadata(html, field string) string {
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

// extractBodyContent extracts the main content from HTML
func extractBodyContent(html string) string {
	// Simplified extraction - find content between header and footer
	// This is a basic implementation and may need improvement
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
	// Clean up extra whitespace
	body = strings.TrimSpace(body)
	return body
}

func init() {
	updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "New title for the share")
	updateCmd.Flags().StringVarP(&updateLede, "lede", "l", "", "New description/lede for the share")
	updateCmd.Flags().StringVarP(&updateType, "type", "y", "", "New content type (note, code, chat, media, raw)")
	updateCmd.Flags().StringVarP(&updateTheme, "theme", "", "", "New theme (paper, dark, terminal)")
	updateCmd.Flags().StringSliceVarP(&updateTags, "tags", "g", nil, "New tags (comma-separated)")
	rootCmd.AddCommand(updateCmd)
}