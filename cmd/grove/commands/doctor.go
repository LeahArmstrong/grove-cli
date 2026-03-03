package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/LeahArmstrong/grove-cli/internal/cli"
	"github.com/LeahArmstrong/grove-cli/plugins/docker"
)

func init() {
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check system health for grove",
	Long: `Run diagnostic checks to verify that grove's dependencies and configuration
are set up correctly.

Checks include Docker availability, external compose configuration,
agent stack readiness, and network connectivity.

Examples:
  grove doctor`,
	RunE: RequireGroveContext(func(cmd *cobra.Command, args []string, ctx *GroveContext) error {
		w := cli.NewStdout()
		cli.Header(w, "grove doctor")

		allPassed := true

		// Check: Docker available
		allPassed = runCheck(w, "Docker available", func() (string, error) {
			if _, err := exec.LookPath("docker"); err != nil {
				return "", fmt.Errorf("docker not found in PATH")
			}
			return "found in PATH", nil
		}) && allPassed

		// Check: Docker daemon running
		allPassed = runCheck(w, "Docker running", func() (string, error) {
			cmd := exec.Command("docker", "info", "--format", "{{.ServerVersion}}")
			out, err := cmd.Output()
			if err != nil {
				return "", fmt.Errorf("docker daemon not responding (is Docker running?)")
			}
			return "v" + strings.TrimSpace(string(out)), nil
		}) && allPassed

		// Check: External compose mode
		cfg := ctx.Config
		if cfg == nil || !cfg.IsExternalDockerMode() {
			runInfo(w, "External mode", "not configured (using local compose)")
		} else {
			ext := cfg.Plugins.Docker.External
			allPassed = runCheck(w, "External compose path", func() (string, error) {
				if ext.Path == "" {
					return "", fmt.Errorf("plugins.docker.external.path not set")
				}
				return ext.Path, nil
			}) && allPassed

			// Check: env_file configuration
			envFileName := ext.EnvFileName()
			composePath := docker.ResolveComposePath(ext.Path)

			if envFileName != ".env" {
				// Using a non-default env file — check direnv prerequisites
				allPassed = runCheck(w, "Env file target", func() (string, error) {
					return envFileName, nil
				}) && allPassed

				// Check direnv is installed
				allPassed = runCheck(w, "direnv available", func() (string, error) {
					if _, err := exec.LookPath("direnv"); err != nil {
						return "", fmt.Errorf("direnv not found — %s won't be loaded by docker compose without it", envFileName)
					}
					return "found in PATH", nil
				}) && allPassed

				// Check .envrc exists with dotenv_if_exists
				allPassed = runCheck(w, "direnv loads "+envFileName, func() (string, error) {
					envrcPath := filepath.Join(composePath, ".envrc")
					data, err := os.ReadFile(envrcPath)
					if err != nil {
						return "", fmt.Errorf(".envrc not found in %s — docker compose won't read %s without direnv", composePath, envFileName)
					}
					if !strings.Contains(string(data), envFileName) {
						return "", fmt.Errorf(".envrc does not reference %s — add: dotenv_if_exists %s", envFileName, envFileName)
					}
					return ".envrc configured", nil
				}) && allPassed
			} else {
				// Using default .env — check if .env.local setup is available
				envrcPath := filepath.Join(composePath, ".envrc")
				if data, err := os.ReadFile(envrcPath); err == nil {
					if strings.Contains(string(data), ".env.local") {
						runInfo(w, "Env file hint", "direnv is configured for .env.local — consider setting env_file = \".env.local\" to avoid dirtying tracked .env")
					}
				}
			}

			// Check: Agent stack config
			if ext.Agent == nil || ext.Agent.Enabled == nil || !*ext.Agent.Enabled {
				runInfo(w, "Agent stacks", "not enabled")
			} else {
				allPassed = runCheck(w, "Agent config", func() (string, error) {
					if len(ext.Agent.Services) == 0 {
						return "", fmt.Errorf("agent.services is empty")
					}
					if ext.Agent.TemplatePath == "" {
						return "", fmt.Errorf("agent.template_path not set")
					}
					return fmt.Sprintf("%d services, max %d slots", len(ext.Agent.Services), ext.Agent.MaxSlots), nil
				}) && allPassed

				// Check: Network exists (if configured)
				if ext.Agent.Network != "" {
					allPassed = runCheck(w, "Docker network '"+ext.Agent.Network+"'", func() (string, error) {
						cmd := exec.Command("docker", "network", "ls", "--format", "{{.Name}}")
						out, err := cmd.Output()
						if err != nil {
							return "", fmt.Errorf("failed to list networks: %w", err)
						}
						for _, line := range strings.Split(string(out), "\n") {
							if strings.TrimSpace(line) == ext.Agent.Network {
								return "exists", nil
							}
						}
						return "", fmt.Errorf("network not found (is the main stack running?)")
					}) && allPassed
				}

				// Check: Active slots
				slots, err := docker.ListActiveSlots(cfg)
				if err == nil {
					maxSlots := ext.Agent.MaxSlots
					if maxSlots <= 0 {
						maxSlots = 5
					}
					runInfo(w, "Active stacks", fmt.Sprintf("%d/%d slots in use", len(slots), maxSlots))
				}
			}
		}

		_, _ = fmt.Fprintln(w)
		if allPassed {
			cli.Success(w, "All checks passed")
		} else {
			cli.Warning(w, "Some checks failed — see above for details")
		}

		return nil
	}),
}

func runCheck(w *cli.Writer, name string, check func() (string, error)) bool {
	detail, err := check()
	if err != nil {
		cli.Error(w, "%s: %v", name, err)
		return false
	}
	if detail != "" {
		cli.Success(w, "%s (%s)", name, detail)
	} else {
		cli.Success(w, "%s", name)
	}
	return true
}

func runInfo(w *cli.Writer, name string, detail string) {
	cli.Info(w, "%s: %s", name, detail)
}
