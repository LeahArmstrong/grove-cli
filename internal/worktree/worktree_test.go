package worktree

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestParseWorktreeList(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		repoRoot    string
		projectName string
		wantCount   int
		wantFirst   string
		wantBranch  string
		wantIsMain  bool
		wantShort   string
	}{
		{
			name:        "single worktree",
			repoRoot:    "/home/user/project",
			projectName: "project",
			input: `worktree /home/user/project
HEAD 1234567
branch refs/heads/main
`,
			wantCount:  1,
			wantFirst:  "/home/user/project",
			wantBranch: "main",
			wantIsMain: true,
			wantShort:  "project",
		},
		{
			name:        "multiple worktrees with project prefix",
			repoRoot:    "/home/user/grove-cli",
			projectName: "grove-cli",
			input: `worktree /home/user/grove-cli
HEAD 1234567
branch refs/heads/main

worktree /home/user/grove-cli-testing
HEAD abcdef0
branch refs/heads/testing
`,
			wantCount:  2,
			wantFirst:  "/home/user/grove-cli",
			wantBranch: "main",
			wantIsMain: true,
			wantShort:  "grove-cli",
		},
		{
			name:        "detached HEAD",
			repoRoot:    "/home/user/project",
			projectName: "project",
			input: `worktree /home/user/project
HEAD 1234567
detached
`,
			wantCount:  1,
			wantFirst:  "/home/user/project",
			wantBranch: "detached",
			wantIsMain: true,
			wantShort:  "project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trees := parseWorktreeList(tt.input, tt.repoRoot, tt.projectName)
			if len(trees) != tt.wantCount {
				t.Errorf("parseWorktreeList() got %d worktrees, want %d", len(trees), tt.wantCount)
			}
			if tt.wantCount > 0 {
				if trees[0].Path != tt.wantFirst {
					t.Errorf("First worktree path = %s, want %s", trees[0].Path, tt.wantFirst)
				}
				if trees[0].Branch != tt.wantBranch {
					t.Errorf("First worktree branch = %s, want %s", trees[0].Branch, tt.wantBranch)
				}
				if trees[0].IsMain != tt.wantIsMain {
					t.Errorf("First worktree IsMain = %v, want %v", trees[0].IsMain, tt.wantIsMain)
				}
				if trees[0].ShortName != tt.wantShort {
					t.Errorf("First worktree ShortName = %s, want %s", trees[0].ShortName, tt.wantShort)
				}
			}
		})
	}
}

func TestWorktreeCreate(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir := t.TempDir()

	// Initialize a git repo
	initCmd := exec.Command("git", "init")
	initCmd.Dir = tmpDir
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user
	configNameCmd := exec.Command("git", "config", "user.name", "Test User")
	configNameCmd.Dir = tmpDir
	if err := configNameCmd.Run(); err != nil {
		t.Fatalf("Failed to config git user.name: %v", err)
	}

	configEmailCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configEmailCmd.Dir = tmpDir
	if err := configEmailCmd.Run(); err != nil {
		t.Fatalf("Failed to config git user.email: %v", err)
	}

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = tmpDir
	if err := addCmd.Run(); err != nil {
		t.Fatalf("Failed to add files: %v", err)
	}

	commitCmd := exec.Command("git", "commit", "-m", "initial commit")
	commitCmd.Dir = tmpDir
	if err := commitCmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	tests := []struct {
		name    string
		wtName  string
		branch  string
		wantErr bool
	}{
		{
			name:    "create new worktree",
			wtName:  "feature-test",
			branch:  "feature-test",
			wantErr: false,
		},
		{
			name:    "empty name",
			wtName:  "",
			branch:  "test",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Manager{
				repoRoot: tmpDir,
			}

			err := m.Create(tt.wtName, tt.branch)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				// Verify worktree was created
				wtPath := filepath.Join(tmpDir, "..", tt.wtName)
				if _, err := os.Stat(wtPath); os.IsNotExist(err) {
					t.Errorf("Worktree directory not created: %s", wtPath)
				}
			}
		})
	}
}

func TestWorktreeList(t *testing.T) {
	// Skip if git is not available
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}

	tmpDir := t.TempDir()

	// Initialize a git repo
	initCmd := exec.Command("git", "init")
	initCmd.Dir = tmpDir
	if err := initCmd.Run(); err != nil {
		t.Fatalf("Failed to init git repo: %v", err)
	}

	// Configure git user
	configNameCmd := exec.Command("git", "config", "user.name", "Test User")
	configNameCmd.Dir = tmpDir
	if err := configNameCmd.Run(); err != nil {
		t.Fatalf("Failed to config git user.name: %v", err)
	}

	configEmailCmd := exec.Command("git", "config", "user.email", "test@example.com")
	configEmailCmd.Dir = tmpDir
	if err := configEmailCmd.Run(); err != nil {
		t.Fatalf("Failed to config git user.email: %v", err)
	}

	// Create initial commit
	testFile := filepath.Join(tmpDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	addCmd := exec.Command("git", "add", ".")
	addCmd.Dir = tmpDir
	if err := addCmd.Run(); err != nil {
		t.Fatalf("Failed to add files: %v", err)
	}

	commitCmd := exec.Command("git", "commit", "-m", "initial commit")
	commitCmd.Dir = tmpDir
	if err := commitCmd.Run(); err != nil {
		t.Fatalf("Failed to commit: %v", err)
	}

	m := &Manager{
		repoRoot: tmpDir,
	}

	trees, err := m.List()
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Should have at least the main worktree
	if len(trees) < 1 {
		t.Errorf("List() returned %d worktrees, want at least 1", len(trees))
	}
}

func TestExtractShortName(t *testing.T) {
	tests := []struct {
		name        string
		fullName    string
		projectName string
		want        string
	}{
		{
			name:        "with project prefix",
			fullName:    "grove-cli-testing",
			projectName: "grove-cli",
			want:        "testing",
		},
		{
			name:        "with project prefix and multiple hyphens",
			fullName:    "grove-cli-feature-auth",
			projectName: "grove-cli",
			want:        "feature-auth",
		},
		{
			name:        "without project prefix",
			fullName:    "testing",
			projectName: "grove-cli",
			want:        "testing",
		},
		{
			name:        "main worktree",
			fullName:    "grove-cli",
			projectName: "grove-cli",
			want:        "grove-cli",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractShortName(tt.fullName, tt.projectName)
			if got != tt.want {
				t.Errorf("ExtractShortName(%q, %q) = %q, want %q", tt.fullName, tt.projectName, got, tt.want)
			}
		})
	}
}

func TestWorktreeDisplayName(t *testing.T) {
	tests := []struct {
		name      string
		worktree  *Worktree
		want      string
	}{
		{
			name: "main worktree",
			worktree: &Worktree{
				IsMain:    true,
				ShortName: "grove-cli",
			},
			want: "main",
		},
		{
			name: "non-main worktree",
			worktree: &Worktree{
				IsMain:    false,
				ShortName: "testing",
			},
			want: "testing",
		},
		{
			name: "non-main with complex name",
			worktree: &Worktree{
				IsMain:    false,
				ShortName: "feature-auth",
			},
			want: "feature-auth",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.worktree.DisplayName()
			if got != tt.want {
				t.Errorf("Worktree.DisplayName() = %q, want %q", got, tt.want)
			}
		})
	}
}
