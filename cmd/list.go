package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/yuler/gh-share/internal/config"
	"github.com/yuler/gh-share/internal/git"
	"github.com/yuler/gh-share/internal/pages"
)

var (
	listLimit int
	listJSON  bool
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List published shares",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Resolve(repoFlag, dirFlag)
		if err != nil {
			return err
		}

		// Ensure local repo exists
		if err := git.EnsureRepo(cfg.LocalDir, cfg.FullRepo); err != nil {
			return err
		}

		shares, err := pages.ListShares(cfg.LocalDir, cfg.BasePagesURL)
		if err != nil {
			return fmt.Errorf("failed to list shares: %w", err)
		}

		if len(shares) == 0 {
			fmt.Printf("No shares found in %s/s/\n", cfg.LocalDir)
			return nil
		}

		if listLimit > 0 && len(shares) > listLimit {
			shares = shares[:listLimit]
		}

		if listJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(shares)
		}

		fmt.Printf("\n%-12s  %-32s  %-35s  %s\n", "DATE", "SLUG", "TITLE", "URL")
		fmt.Printf("%-12s  %-32s  %-35s  %s\n", "----", "----", "-----", "---")
		for _, s := range shares {
			title := s.Title
			if len(title) > 33 {
				title = title[:30] + "..."
			}
			slug := s.Slug
			if len(slug) > 30 {
				slug = slug[:28] + ".."
			}
			fmt.Printf("%-12s  %-32s  %-35s  %s\n", s.Date, slug, title, s.URL)
		}
		fmt.Println()
		return nil
	},
}

func init() {
	listCmd.Flags().IntVarP(&listLimit, "limit", "L", 30, "Maximum number of shares to display")
	listCmd.Flags().BoolVar(&listJSON, "json", false, "Output results in JSON format")

	rootCmd.AddCommand(listCmd)
}
