package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Worktree represents a git worktree
type Worktree struct {
	Name      string // Short name (derived from path)
	Path      string // Absolute path to worktree
	Branch    string // Branch name or "detached"
	Commit    string // Commit hash
	IsDirty   bool   // Whether there are uncommitted changes
	IsMain    bool   // Whether this is the main worktree
	ShortName string // Short name without project prefix
}

// Manager handles git worktree operations
type Manager struct {
	repoRoot string // Root of the git repository
}

// NewManager creates a new worktree manager
// If repoRoot is empty, it will detect from current directory
func NewManager(repoRoot string) (*Manager, error) {
	if repoRoot == "" {
		// Try to detect repo root from current directory
		cmd := exec.Command("git", "rev-parse", "--show-toplevel")
		output, err := cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("not in a git repository: %w", err)
		}
		repoRoot = strings.TrimSpace(string(output))
	}

	return &Manager{
		repoRoot: repoRoot,
	}, nil
}

// Create creates a new worktree
func (m *Manager) Create(name, branch string) error {
	if name == "" {
		return fmt.Errorf("worktree name cannot be empty")
	}

	// Worktree path is relative to repo root's parent
	wtPath := filepath.Join(filepath.Dir(m.repoRoot), name)

	// Check if worktree already exists
	if _, err := os.Stat(wtPath); err == nil {
		return fmt.Errorf("worktree already exists at %s", wtPath)
	}

	// Create worktree with new branch
	args := []string{"worktree", "add", "-b", branch, wtPath}
	cmd := exec.Command("git", args...)
	cmd.Dir = m.repoRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree: %s: %w", string(output), err)
	}

	return nil
}

// CreateFromExisting creates a worktree from an existing branch
func (m *Manager) CreateFromExisting(name, branch string) error {
	if name == "" {
		return fmt.Errorf("worktree name cannot be empty")
	}

	wtPath := filepath.Join(filepath.Dir(m.repoRoot), name)

	// Check if worktree already exists
	if _, err := os.Stat(wtPath); err == nil {
		return fmt.Errorf("worktree already exists at %s", wtPath)
	}

	// Create worktree from existing branch (no -b flag)
	args := []string{"worktree", "add", wtPath, branch}
	cmd := exec.Command("git", args...)
	cmd.Dir = m.repoRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create worktree: %s: %w", string(output), err)
	}

	return nil
}

// List returns all worktrees in the repository
func (m *Manager) List() ([]*Worktree, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = m.repoRoot

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list worktrees: %w", err)
	}

	projectName := m.GetProjectName()
	trees := parseWorktreeList(string(output), m.repoRoot, projectName)

	// Check dirty status for each worktree
	for _, tree := range trees {
		dirty, err := m.isDirty(tree.Path)
		if err == nil {
			tree.IsDirty = dirty
		}
	}

	return trees, nil
}

// Remove removes a worktree
func (m *Manager) Remove(name string) error {
	if name == "" {
		return fmt.Errorf("worktree name cannot be empty")
	}

	// Find the worktree by name
	trees, err := m.List()
	if err != nil {
		return fmt.Errorf("failed to list worktrees: %w", err)
	}

	var targetPath string
	for _, tree := range trees {
		if tree.Name == name {
			targetPath = tree.Path
			break
		}
	}

	if targetPath == "" {
		return fmt.Errorf("worktree '%s' not found", name)
	}

	// Remove the worktree
	cmd := exec.Command("git", "worktree", "remove", targetPath)
	cmd.Dir = m.repoRoot

	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try force remove if regular remove fails
		cmd = exec.Command("git", "worktree", "remove", "--force", targetPath)
		cmd.Dir = m.repoRoot
		output, err = cmd.CombinedOutput()
		if err != nil {
			return fmt.Errorf("failed to remove worktree: %s: %w", string(output), err)
		}
	}

	return nil
}

// GetCurrent returns the current worktree
func (m *Manager) GetCurrent() (*Worktree, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get current worktree: %w", err)
	}

	currentPath := strings.TrimSpace(string(output))

	trees, err := m.List()
	if err != nil {
		return nil, err
	}

	for _, tree := range trees {
		if tree.Path == currentPath {
			return tree, nil
		}
	}

	return nil, fmt.Errorf("current worktree not found")
}

// isDirty checks if a worktree has uncommitted changes
func (m *Manager) isDirty(path string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = path

	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return len(strings.TrimSpace(string(output))) > 0, nil
}

// GetProjectName extracts the project name from the repository
func (m *Manager) GetProjectName() string {
	// Try to get from git remote URL first
	cmd := exec.Command("git", "remote", "get-url", "origin")
	cmd.Dir = m.repoRoot
	if output, err := cmd.Output(); err == nil {
		url := strings.TrimSpace(string(output))
		// Extract repo name from URL
		// Examples:
		// https://github.com/user/repo.git -> repo
		// git@github.com:user/repo.git -> repo
		// /path/to/repo -> repo
		parts := strings.Split(url, "/")
		if len(parts) > 0 {
			name := parts[len(parts)-1]
			name = strings.TrimSuffix(name, ".git")
			if name != "" {
				return name
			}
		}
	}
	
	// Fall back to directory name
	return filepath.Base(m.repoRoot)
}

// DisplayName returns the display name for a worktree
// Main worktree returns "main", others return short name without project prefix
func (w *Worktree) DisplayName() string {
	if w.IsMain {
		return "main"
	}
	return w.ShortName
}

// ExtractShortName removes the project prefix from a worktree name
// e.g., "grove-cli-testing" with project "grove-cli" returns "testing"
func ExtractShortName(fullName, projectName string) string {
	prefix := projectName + "-"
	if strings.HasPrefix(fullName, prefix) {
		return strings.TrimPrefix(fullName, prefix)
	}
	return fullName
}

// parseWorktreeList parses the output of 'git worktree list --porcelain'
func parseWorktreeList(output, repoRoot, projectName string) []*Worktree {
	var trees []*Worktree
	var current *Worktree

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			if current != nil {
				trees = append(trees, current)
				current = nil
			}
			continue
		}

		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			name := filepath.Base(path)
			isMain := path == repoRoot
			shortName := ExtractShortName(name, projectName)
			
			current = &Worktree{
				Path:      path,
				Name:      name,
				IsMain:    isMain,
				ShortName: shortName,
			}
		} else if strings.HasPrefix(line, "HEAD ") {
			if current != nil {
				current.Commit = strings.TrimPrefix(line, "HEAD ")
			}
		} else if strings.HasPrefix(line, "branch ") {
			if current != nil {
				branch := strings.TrimPrefix(line, "branch ")
				// Remove refs/heads/ prefix
				branch = strings.TrimPrefix(branch, "refs/heads/")
				current.Branch = branch
			}
		} else if strings.HasPrefix(line, "detached") {
			if current != nil {
				current.Branch = "detached"
			}
		}
	}

	// Don't forget the last worktree
	if current != nil {
		trees = append(trees, current)
	}

	return trees
}
