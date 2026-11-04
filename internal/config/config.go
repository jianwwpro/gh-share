package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cli/go-gh/v2/pkg/api"
)

type Config struct {
	Owner       string
	RepoName    string
	FullRepo    string // "owner/repo"
	BasePagesURL string // "https://owner.github.io/repo/"
	LocalDir    string // path to local checkout
}

// Resolve resolves the target repo and local directory
func Resolve(repoFlag string, dirFlag string) (*Config, error) {
	cfg := &Config{}

	fullRepo := repoFlag
	if fullRepo == "" {
		fullRepo = os.Getenv("GH_SHARE_REPO")
	}

	if fullRepo == "" {
		// Detect current GitHub user
		client, err := api.DefaultRESTClient()
		if err != nil {
			return nil, fmt.Errorf("failed to create GitHub API client: %w (are you logged in with 'gh auth login'?)", err)
		}
		var userResp struct {
			Login string `json:"login"`
		}
		if err := client.Get("user", &userResp); err != nil {
			return nil, fmt.Errorf("failed to fetch current GitHub user: %w", err)
		}
		if userResp.Login == "" {
			return nil, fmt.Errorf("could not determine GitHub username")
		}
		cfg.Owner = userResp.Login
		cfg.RepoName = "share"
		cfg.FullRepo = fmt.Sprintf("%s/%s", cfg.Owner, cfg.RepoName)
	} else {
		parts := strings.Split(fullRepo, "/")
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid repository format '%s', expected 'owner/repo'", fullRepo)
		}
		cfg.Owner = parts[0]
		cfg.RepoName = parts[1]
		cfg.FullRepo = fullRepo
	}

	// Calculate base GitHub Pages URL
	if strings.EqualFold(cfg.RepoName, cfg.Owner+".github.io") {
		cfg.BasePagesURL = fmt.Sprintf("https://%s.github.io/", cfg.Owner)
	} else {
		cfg.BasePagesURL = fmt.Sprintf("https://%s.github.io/%s/", cfg.Owner, cfg.RepoName)
	}

	// Resolve local directory
	if dirFlag != "" {
		abs, err := filepath.Abs(dirFlag)
		if err != nil {
			return nil, err
		}
		cfg.LocalDir = abs
		return cfg, nil
	}

	// Check current directory
	if isRepoMatch(cfg.FullRepo, ".") {
		cwd, _ := os.Getwd()
		cfg.LocalDir = cwd
		return cfg, nil
	}

	// Check ~/Projects/<repo>
	home, err := os.UserHomeDir()
	if err == nil {
		candidate := filepath.Join(home, "Projects", cfg.RepoName)
		if isRepoMatch(cfg.FullRepo, candidate) {
			cfg.LocalDir = candidate
			return cfg, nil
		}
	}

	// Default candidate path in ~/Projects/<repo>
	if home != "" {
		cfg.LocalDir = filepath.Join(home, "Projects", cfg.RepoName)
	} else {
		cfg.LocalDir = filepath.Join(".", cfg.RepoName)
	}

	return cfg, nil
}

func isRepoMatch(expectedFullRepo string, dir string) bool {
	cmd := exec.Command("git", "-C", dir, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	url := strings.TrimSpace(string(out))
	expected := strings.ToLower(expectedFullRepo)
	urlLower := strings.ToLower(url)

	// e.g. https://github.com/owner/repo.git or git@github.com:owner/repo.git
	return strings.Contains(urlLower, "/"+expected+".git") ||
		strings.Contains(urlLower, "/"+expected) ||
		strings.Contains(urlLower, ":"+expected+".git") ||
		strings.Contains(urlLower, ":"+expected)
}
