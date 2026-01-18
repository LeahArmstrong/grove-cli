package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
)

// projectConfig represents the project-level configuration
type projectConfig struct {
	ProjectName string `toml:"project_name"`
}

// detectProjectName determines the project name using priority:
// 1. .grove/config.toml -> project_name
// 2. Git remote origin URL -> repo name
// 3. Parent directory name as fallback
func (m *Manager) detectProjectName() string {
	// Priority 1: Check .grove/config.toml
	configPath := filepath.Join(m.repoRoot, ".grove", "config.toml")
	if data, err := os.ReadFile(configPath); err == nil {
		var cfg projectConfig
		if err := toml.Unmarshal(data, &cfg); err == nil && cfg.ProjectName != "" {
			return cfg.ProjectName
		}
	}

	// Priority 2: Extract from git remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = m.repoRoot
	if output, err := cmd.Output(); err == nil {
		remoteURL := strings.TrimSpace(string(output))
		if projectName := extractProjectNameFromRemote(remoteURL); projectName != "" {
			return projectName
		}
	}

	// Priority 3: Use directory name
	return filepath.Base(m.repoRoot)
}

// extractProjectNameFromRemote extracts the repository name from a git remote URL
// Handles both HTTPS and SSH formats:
// - https://github.com/user/repo.git -> repo
// - git@github.com:user/repo.git -> repo
// - https://github.com/user/repo -> repo
func extractProjectNameFromRemote(remoteURL string) string {
	// Remove .git suffix if present
	remoteURL = strings.TrimSuffix(remoteURL, ".git")

	// Handle SSH format (git@github.com:user/repo)
	if strings.Contains(remoteURL, ":") && strings.Contains(remoteURL, "@") {
		parts := strings.Split(remoteURL, ":")
		if len(parts) >= 2 {
			path := parts[len(parts)-1]
			return filepath.Base(path)
		}
	}

	// Handle HTTPS format (https://github.com/user/repo)
	// Just get the last component of the path
	return filepath.Base(remoteURL)
}

// FullName returns the full worktree name with project prefix
// Format: {project}-{name}
// Example: grove-cli-testing
func (m *Manager) FullName(wt *Worktree) string {
	if m.projectName == "" {
		m.projectName = m.detectProjectName()
	}

	// If the worktree is the main repo, return just the project name
	if wt.Path == m.repoRoot {
		return m.projectName
	}

	// Return project-name format
	return m.projectName + "-" + wt.Name
}

// DisplayName returns the short name for display purposes
// Strips the project prefix if present
// Example: grove-cli-testing -> testing
func (m *Manager) DisplayName(wt *Worktree) string {
	if m.projectName == "" {
		m.projectName = m.detectProjectName()
	}

	// Get the base name from the path
	baseName := filepath.Base(wt.Path)

	// If it's the main repo, return the project name
	if wt.Path == m.repoRoot {
		return baseName
	}

	// Try to strip the project prefix
	prefix := m.projectName + "-"
	if strings.HasPrefix(baseName, prefix) {
		return strings.TrimPrefix(baseName, prefix)
	}

	// No prefix found, return as-is
	return baseName
}

// GetProjectName returns the detected project name
func (m *Manager) GetProjectName() string {
	if m.projectName == "" {
		m.projectName = m.detectProjectName()
	}
	return m.projectName
}
