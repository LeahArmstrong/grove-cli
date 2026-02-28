package commands

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/LeahArmstrong/grove-cli/internal/cli"
	"github.com/LeahArmstrong/grove-cli/internal/config"
	"github.com/LeahArmstrong/grove-cli/internal/exitcode"
	"github.com/LeahArmstrong/grove-cli/internal/grove"
	"github.com/LeahArmstrong/grove-cli/internal/hooks"
	"github.com/LeahArmstrong/grove-cli/internal/output"
	"github.com/LeahArmstrong/grove-cli/internal/state"
	"github.com/LeahArmstrong/grove-cli/internal/tmux"
	"github.com/LeahArmstrong/grove-cli/internal/worktree"
	"github.com/LeahArmstrong/grove-cli/plugins/docker"
)

var (
	newJSON     bool
	newMirror   string // Remote branch to mirror (e.g., "origin/main")
	newNoDocker bool   // Skip auto-starting Docker
)

var newCmd = &cobra.Command{
	Use:     "new <name>",
	Aliases: []string{"spawn"},
	Short:   "Create a new worktree and tmux session",
	Long: `Create a new git worktree with the specified name and create a tmux session for it.

The worktree will be created in the parent directory of the current repository.
A new branch with the same name will be created automatically.

When Docker agent stacks are configured, containers start automatically.
Use --no-docker to skip Docker auto-start.

Use --mirror to create an environment worktree that tracks a remote branch.
Environment worktrees are read-only and can be synced with 'grove sync'.

Examples:
  grove new feature-auth                  # Create worktree + tmux + Docker
  grove new feature-auth --no-docker      # Skip Docker auto-start
  grove spawn feature-x                   # Alias (implies --json output)
  grove new staging --mirror origin/main  # Environment worktree tracking origin/main`,
	Args: cobra.ExactArgs(1),
	RunE: RequireGroveContext(func(cmd *cobra.Command, args []string, ctx *GroveContext) error {
		// spawn alias implies JSON output
		if cmd.CalledAs() == "spawn" {
			newJSON = true
		}

		w := cli.NewStdout()
		stderr := cli.NewStderr()

		name := args[0]
		if name == "" {
			return fmt.Errorf("worktree name cannot be empty")
		}

		mgr, err := worktree.NewManager(ctx.ProjectRoot)
		if err != nil {
			return fmt.Errorf("failed to initialize worktree manager: %w", err)
		}

		// Check if worktree already exists
		if existingWt, _ := mgr.Find(name); existingWt != nil {
			return fmt.Errorf("worktree '%s' already exists\n\nOptions:\n  • Switch to it: grove to %s\n  • Remove it first: grove rm %s\n  • Use different name: grove new %s-v2",
				name, name, name, name)
		}

		var branchName string
		isEnvironment := newMirror != ""

		if isEnvironment {
			// Environment worktree - verify remote branch exists
			if !strings.Contains(newMirror, "/") {
				// Assume origin if no remote specified
				newMirror = "origin/" + newMirror
			}

			// Fetch to ensure we have latest refs
			fetchCmd := exec.Command("git", "-C", ctx.ProjectRoot, "fetch", "--prune")
			_ = fetchCmd.Run()

			// Verify the remote branch exists
			verifyCmd := exec.Command("git", "-C", ctx.ProjectRoot, "rev-parse", "--verify", newMirror)
			if err := verifyCmd.Run(); err != nil {
				cli.Error(stderr, "remote branch '%s' not found", newMirror)
				cli.Faint(stderr, "Run 'git fetch' and verify the branch exists")
				os.Exit(exitcode.ResourceNotFound)
			}

			// Use env/{name} as local branch for environment worktrees
			branchName = "env/" + name

			// Create worktree from the remote branch
			if err := mgr.CreateFromBranch(name, newMirror); err != nil {
				return fmt.Errorf("failed to create environment worktree: %w", err)
			}

			if !newJSON {
				cli.Success(w, "Created environment worktree '%s' tracking %s", name, newMirror)
			}
		} else {
			// Regular worktree - use name as branch name
			branchName = name
			if err := mgr.Create(name, branchName); err != nil {
				return fmt.Errorf("failed to create worktree: %w", err)
			}

			if !newJSON {
				cli.Success(w, "Created worktree '%s'", name)
			}
		}

		// Find the newly created worktree to get its path
		wt, err := mgr.Find(name)
		if err != nil || wt == nil {
			return fmt.Errorf("failed to find created worktree: %w", err)
		}

		// Symlink config.toml from main worktree
		_ = grove.EnsureConfigSymlink(ctx.ProjectRoot, wt.Path)

		// Register worktree in state
		now := time.Now()
		wsState := &state.WorktreeState{
			Path:           wt.Path,
			Branch:         branchName,
			Root:           false,
			CreatedAt:      now,
			LastAccessedAt: now,
			Environment:    isEnvironment,
		}

		if isEnvironment {
			wsState.Mirror = newMirror
			wsState.LastSyncedAt = &now
		}

		if err := ctx.State.AddWorktree(name, wsState); err != nil {
			cli.Warning(w, "worktree created but state tracking failed: %v", err)
			cli.Faint(w, "run 'grove repair' to fix")
		}

		projectName := mgr.GetProjectName()

		// Create tmux session if tmux is available
		if tmux.IsTmuxAvailable() {
			sessionName := worktree.TmuxSessionName(projectName, name)
			if err := tmux.CreateSession(sessionName, wt.Path); err != nil {
				if !newJSON {
					cli.Warning(w, "Failed to create tmux session: %v", err)
				}
			} else if !newJSON {
				cli.Success(w, "Created tmux session '%s'", sessionName)
			}
		}

		// Execute post-create hooks
		hookExecutor, err := hooks.NewExecutor()
		if err != nil {
			if !newJSON {
				cli.Warning(w, "Failed to load hooks config: %v", err)
			}
		} else if hookExecutor.HasHooksForEvent(hooks.EventPostCreate) {
			hookCtx := &hooks.ExecutionContext{
				Event:        hooks.EventPostCreate,
				Worktree:     name,
				WorktreeFull: projectName + "-" + name,
				Branch:       branchName,
				Project:      projectName,
				MainPath:     ctx.ProjectRoot,
				NewPath:      wt.Path,
			}

			if !newJSON {
				cli.Step(w, "Running post-create hooks...")
			}

			if err := hookExecutor.Execute(hooks.EventPostCreate, hookCtx); err != nil {
				if !newJSON {
					cli.Warning(w, "Hook execution had errors: %v", err)
				}
			}
		}

		// Fire global registry post-create hook (for plugins like docker external)
		globalHookCtx := &hooks.Context{
			Worktree:     name,
			Config:       ctx.Config,
			WorktreePath: wt.Path,
			MainPath:     ctx.ProjectRoot,
		}
		if err := hooks.Fire(hooks.EventPostCreate, globalHookCtx); err != nil {
			if !newJSON {
				cli.Warning(w, "Post-create plugin hook failed: %v", err)
			}
		}

		// Auto-start Docker when configured (unless --no-docker)
		if !newNoDocker && shouldAutoDocker(ctx.Config) {
			if !newJSON {
				cli.Step(w, "Starting Docker stack...")
			}
			dockerPlugin := docker.New()
			if ctx.Config.AgentMode {
				dockerPlugin.SetIsolated(true)
			}
			if err := dockerPlugin.Init(ctx.Config); err != nil {
				if !newJSON {
					cli.Warning(w, "Docker init failed: %v", err)
				}
			} else if dockerPlugin.Enabled() {
				if err := dockerPlugin.Up(wt.Path, true); err != nil {
					if !newJSON {
						cli.Warning(w, "Docker auto-start failed: %v", err)
					}
				} else if !newJSON {
					cli.Success(w, "Docker stack started")
				}
			}
		}

		// JSON output mode
		if newJSON {
			result := output.NewWorktreeResult{
				SwitchTo: wt.Path,
				Name:     name,
				Branch:   branchName,
				Path:     wt.Path,
				Created:  true,
			}
			if err := output.PrintJSON(result); err != nil {
				return err
			}
		}

		return nil
	}),
}

// shouldAutoDocker returns true when Docker should be auto-started on grove new.
// Enabled by default when agent stacks are configured, or explicitly via auto_up.
func shouldAutoDocker(cfg *config.Config) bool {
	if cfg == nil {
		return false
	}

	// Explicit auto_up setting takes precedence
	if cfg.Plugins.Docker.AutoUp != nil {
		return *cfg.Plugins.Docker.AutoUp
	}

	// Default: auto-start when agent stacks are configured and enabled
	if cfg.IsExternalDockerMode() {
		ext := cfg.Plugins.Docker.External
		if ext != nil && ext.Agent != nil && ext.Agent.Enabled != nil && *ext.Agent.Enabled {
			return true
		}
	}

	return false
}

func init() {
	newCmd.Flags().BoolVarP(&newJSON, "json", "j", false, "Output as JSON with switch_to field")
	newCmd.Flags().StringVar(&newMirror, "mirror", "", "Create environment worktree tracking a remote branch (e.g., origin/main)")
	newCmd.Flags().BoolVar(&newNoDocker, "no-docker", false, "Skip Docker auto-start")
	rootCmd.AddCommand(newCmd)
}
