package skill

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/yuler/gh-share/embedded"
)

type LocationStatus struct {
	Name      string
	Path      string
	Installed bool
	Target    string // if symlink
	Note      string
}

// GetLocations returns candidate skill install targets
func GetLocations() ([]LocationStatus, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	targets := []struct {
		name string
		path string
	}{
		{name: "Agents System (~/.agents/skills/share)", path: filepath.Join(home, ".agents", "skills", "share")},
		{name: "Antigravity / Gemini CLI (~/.gemini/config/skills/share)", path: filepath.Join(home, ".gemini", "config", "skills", "share")},
		{name: "Claude Code (~/.claude/skills/share)", path: filepath.Join(home, ".claude", "skills", "share")},
	}

	var statuses []LocationStatus
	for _, t := range targets {
		st := LocationStatus{
			Name: t.name,
			Path: t.path,
		}

		lstat, err := os.Lstat(t.path)
		if err == nil {
			st.Installed = true
			if lstat.Mode()&os.ModeSymlink != 0 {
				linkTarget, _ := os.Readlink(t.path)
				st.Target = linkTarget
				st.Note = fmt.Sprintf("symlink -> %s", linkTarget)
			} else {
				st.Note = "directory"
			}
		} else {
			parent := filepath.Dir(t.path)
			if _, pErr := os.Stat(parent); pErr == nil {
				st.Note = "parent dir exists, ready to install"
			} else {
				st.Note = "parent dir not found"
			}
		}
		statuses = append(statuses, st)
	}

	return statuses, nil
}

// Install installs the embedded SKILL.md to ~/.agents and symlinks to agent platforms
func Install(force bool) ([]LocationStatus, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	// 1. Read embedded SKILL.md
	skillData, err := fs.ReadFile(embedded.Content, "skill/SKILL.md")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded skill: %w", err)
	}

	// Primary location: ~/.agents/skills/share/SKILL.md
	primaryDir := filepath.Join(home, ".agents", "skills", "share")
	if err := os.MkdirAll(primaryDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create %s: %w", primaryDir, err)
	}

	primaryFile := filepath.Join(primaryDir, "SKILL.md")
	if err := os.WriteFile(primaryFile, skillData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write %s: %w", primaryFile, err)
	}

	// Secondary locations (symlinks pointing to primaryDir)
	secondaryParents := []string{
		filepath.Join(home, ".gemini", "config", "skills"),
		filepath.Join(home, ".claude", "skills"),
	}

	for _, parent := range secondaryParents {
		if _, err := os.Stat(parent); err == nil {
			targetLink := filepath.Join(parent, "share")
			lstat, lerr := os.Lstat(targetLink)
			if lerr == nil {
				if force {
					_ = os.Remove(targetLink)
					_ = os.Symlink(primaryDir, targetLink)
				} else if lstat.Mode()&os.ModeSymlink == 0 {
					// directory already exists; write file inside
					_ = os.WriteFile(filepath.Join(targetLink, "SKILL.md"), skillData, 0644)
				}
			} else {
				_ = os.Symlink(primaryDir, targetLink)
			}
		}
	}

	return GetLocations()
}

// Uninstall removes the skill from all known locations
func Uninstall() ([]LocationStatus, error) {
	statuses, err := GetLocations()
	if err != nil {
		return nil, err
	}

	for _, st := range statuses {
		if st.Installed {
			_ = os.RemoveAll(st.Path)
		}
	}

	return GetLocations()
}
