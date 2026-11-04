package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yuler/gh-share/internal/config"
)

var openCmd = &cobra.Command{
	Use:     "open [slug]",
	Aliases: []string{"view"},
	Short:   "Open a share in your default web browser",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Resolve(repoFlag, dirFlag)
		if err != nil {
			return err
		}

		targetURL := cfg.BasePagesURL
		if len(args) > 0 {
			slug := strings.Trim(args[0], "/")
			targetURL = fmt.Sprintf("%ss/%s/", strings.TrimRight(cfg.BasePagesURL, "/")+"/", slug)
		}

		fmt.Printf("🌐 Opening: %s\n", targetURL)
		openBrowser(targetURL)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}
