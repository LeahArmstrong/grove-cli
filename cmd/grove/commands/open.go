package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/LeahArmstrong/grove-cli/internal/cli"
	"github.com/LeahArmstrong/grove-cli/internal/grove"
	"github.com/LeahArmstrong/grove-cli/internal/hooks"
	"github.com/LeahArmstrong/grove-cli/internal/output"
	"github.com/LeahArmstrong/grove-cli/internal/state"
	"github.com/LeahArmstrong/grove-cli/internal/tmux"
	"github.com/LeahArmstrong/grove-cli/internal/worktree"
	"github.com/LeahArmstrong/grove-cli/plugins/docker"
)

var (
	openNoCreate bool
	openCommand  string
	openNoPopup  bool
	openJSON     bool
	openNoDocker bool
)

func init() {
	openCmd.Flags().BoolVar(&openNoCreate, "no-create", false, "Only attach to existing worktree, error if not found")
	openCmd.Flags().StringVar(&openCommand, "command", "", "Override session command")
	openCmd.Flags().BoolVar(&openNoPopup, "no-popup", false, "Skip popup, use tmux switch/attach instead")
	openCmd.Flags().BoolVarP(&openJSON, "json", "j", false, "Output as JSON")
	openCmd.Flags().BoolVar(&openNoDocker, "no-docker", false, "Skip Docker auto-start")
	rootCmd.AddCommand(openCmd)
}

var openCmd = &cobra.Command{
	Use:   "open <name>",
	Short: "Open a worktree session (create if needed)",
	Long: `Open a worktree by creating it if needed, ensuring a tmux session exists,
launching the configured session command, and attaching.

This is idempotent — safe to run repeatedly. If the worktree and session
already exist, it reattaches without recreating.

The session command is configured in .grove/config.toml:

  [session]
  command = "claude"   # What to run (default: shell)
  popup = true         # Use tmux display-popup
  popup_width = "80%"
  popup_height = "80%"

Examples:
  grove open feature-x              # Create if needed + attach + launch command
  grove open feature-x --no-create  # Attach only, error if doesn't exist
  grove open feature-x --command sh # Override the session command`,
	Args: cobra.ExactArgs(1),
	RunE: RequireGroveContext(func(cmd *cobra.Command, args []string, ctx *GroveContext) error {
		name := args[0]
		w := cli.NewStdout()
		stderr := cli.NewStderr()

		if name == "" {
			return fmt.Errorf("worktree name cannot be empty")
		}

		mgr, err := worktree.NewManager(ctx.ProjectRoot)
		if err != nil {
			return fmt.Errorf("failed to initialize worktree manager: %w", err)
		}

		projectName := mgr.GetProjectName()

		// Step 1: Ensure worktree exists
		wt, _ := mgr.Find(name)
		created := false

		if wt == nil {
			if openNoCreate {
				return fmt.Errorf("worktree '%s' not found (use 'grove open %s' without --no-create to create it)", name, name)
			}

			// Create the worktree (reuse logic from grove new)
			branchName := name
			if err := mgr.Create(name, branchName); err != nil {
				return fmt.Errorf("failed to create worktree: %w", err)
			}

			wt, err = mgr.Find(name)
			if err != nil || wt == nil {
				return fmt.Errorf("failed to find created worktree: %w", err)
			}

			// Symlink config
			_ = grove.EnsureConfigSymlink(ctx.ProjectRoot, wt.Path)

			// Register in state
			now := time.Now()
			wsState := &state.WorktreeState{
				Path:           wt.Path,
				Branch:         branchName,
				Root:           false,
				CreatedAt:      now,
				LastAccessedAt: now,
			}
			if err := ctx.State.AddWorktree(name, wsState); err != nil {
				if !openJSON {
					cli.Warning(stderr, "state tracking failed: %v", err)
				}
			}

			// Fire post-create hooks
			globalHookCtx := &hooks.Context{
				Worktree:     name,
				Config:       ctx.Config,
				WorktreePath: wt.Path,
				MainPath:     ctx.ProjectRoot,
			}
			if err := hooks.Fire(hooks.EventPostCreate, globalHookCtx); err != nil {
				if !openJSON {
					cli.Warning(stderr, "post-create hooks failed: %v", err)
				}
			}

			// Execute per-project post-create hooks
			hookExecutor, err := hooks.NewExecutor()
			if err != nil {
				if !openJSON {
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

				if !openJSON {
					cli.Step(w, "Running post-create hooks...")
				}

				if err := hookExecutor.Execute(hooks.EventPostCreate, hookCtx); err != nil {
					if !openJSON {
						cli.Warning(w, "Hook execution had errors: %v", err)
					}
				}
			}

			// Auto-start Docker when configured (unless --no-docker)
			if !openNoDocker && shouldAutoDocker(ctx.Config) {
				if !openJSON {
					cli.Step(w, "Starting Docker stack...")
				}
				dockerPlugin := docker.New()
				if ctx.Config.AgentMode {
					dockerPlugin.SetIsolated(true)
				}
				if err := dockerPlugin.Init(ctx.Config); err != nil {
					if !openJSON {
						cli.Warning(w, "Docker init failed: %v", err)
					}
				} else if dockerPlugin.Enabled() {
					if err := dockerPlugin.Up(wt.Path, true); err != nil {
						if !openJSON {
							cli.Warning(w, "Docker auto-start failed: %v", err)
						}
					} else if !openJSON {
						cli.Success(w, "Docker stack started")
					}
				}
			}

			created = true
			if !openJSON {
				cli.Success(w, "Created worktree '%s'", name)
			}
		}

		// Step 2: Ensure tmux session exists
		if !tmux.IsTmuxAvailable() {
			if openJSON {
				return printOpenJSON(wt, name, created)
			}
			cli.Faint(stderr, "tmux not available, skipping session management")
			cli.Directive("cd", wt.Path)
			return nil
		}

		sessionName := worktree.TmuxSessionName(projectName, name)
		sessionExists, err := tmux.SessionExists(sessionName)
		if err != nil {
			return fmt.Errorf("failed to check session: %w", err)
		}

		// Determine the session command
		sessionCmd := ctx.Config.Session.Command
		if openCommand != "" {
			sessionCmd = openCommand
		}

		if !sessionExists {
			// Create session with command if configured
			if err := tmux.CreateSessionWithCommand(sessionName, wt.Path, sessionCmd); err != nil {
				return fmt.Errorf("failed to create session: %w", err)
			}
			if !openJSON {
				if sessionCmd != "" {
					cli.Success(w, "Created session '%s' running '%s'", sessionName, sessionCmd)
				} else {
					cli.Success(w, "Created session '%s'", sessionName)
				}
			}
		} else if sessionCmd != "" && !tmux.IsCommandRunning(sessionName, sessionCmd) {
			// Session exists but command not running — send it
			pane, pErr := tmux.GetPaneInfo(sessionName)
			if pErr == nil && pane.IsShell() {
				_ = tmux.SendKeys(sessionName, sessionCmd)
				if !openJSON {
					cli.Success(w, "Launched '%s' in existing session", sessionCmd)
				}
			}
		}

		// Update state
		_ = ctx.State.TouchWorktree(name)

		// JSON output mode
		if openJSON {
			return printOpenJSON(wt, name, created)
		}

		// Step 3: Attach — popup or switch/attach
		usePopup := ctx.Config.Session.Popup != nil && *ctx.Config.Session.Popup && !openNoPopup

		if usePopup && tmux.IsInsideTmux() {
			width := ctx.Config.Session.PopupWidth
			height := ctx.Config.Session.PopupHeight
			return tmux.DisplayPopup(sessionName, width, height)
		}

		// Standard tmux attach/switch
		if tmux.IsInsideTmux() {
			return tmux.SwitchSession(sessionName)
		}

		// Outside tmux
		hasShellIntegration := os.Getenv("GROVE_SHELL") == "1"
		if hasShellIntegration {
			cli.Directive("cd", wt.Path)
			cli.Directive("tmux-attach", sessionName)
		} else {
			return tmux.AttachSession(sessionName)
		}

		return nil
	}),
}

func printOpenJSON(wt *worktree.Worktree, name string, created bool) error {
	result := output.NewWorktreeResult{
		SwitchTo: wt.Path,
		Name:     name,
		Branch:   wt.Branch,
		Path:     wt.Path,
		Created:  created,
	}
	data, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(data))
	return nil
}
