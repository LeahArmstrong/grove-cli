package docker

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadEnvVar(t *testing.T) {
	dir := t.TempDir()
	envContent := `ADMIN_DIR=/some/path
OTHER_VAR=value
EMPTY_VAR=
`
	if err := os.WriteFile(filepath.Join(dir, ".env"), []byte(envContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"existing key", "ADMIN_DIR", "/some/path"},
		{"another key", "OTHER_VAR", "value"},
		{"empty value", "EMPTY_VAR", ""},
		{"missing key", "NONEXISTENT", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readEnvVar(dir, tt.key)
			if got != tt.want {
				t.Errorf("readEnvVar(%q, %q) = %q, want %q", dir, tt.key, got, tt.want)
			}
		})
	}
}

func TestReadEnvVar_NoFile(t *testing.T) {
	got := readEnvVar(t.TempDir(), "KEY")
	if got != "" {
		t.Errorf("expected empty string for missing .env, got %q", got)
	}
}

func TestPathMatchesEnv(t *testing.T) {
	tests := []struct {
		name         string
		worktreePath string
		envValue     string
		composePath  string
		want         bool
	}{
		{
			name:         "empty env value",
			worktreePath: "/work/project-feature",
			envValue:     "",
			composePath:  "/docker",
			want:         false,
		},
		{
			name:         "absolute match",
			worktreePath: "/work/project-feature",
			envValue:     "/work/project-feature",
			composePath:  "/docker",
			want:         true,
		},
		{
			name:         "absolute no match",
			worktreePath: "/work/project-feature",
			envValue:     "/work/project-other",
			composePath:  "/docker",
			want:         false,
		},
		{
			name:         "relative match",
			worktreePath: "/docker/project-feature",
			envValue:     "./project-feature",
			composePath:  "/docker",
			want:         true,
		},
		{
			name:         "relative no match",
			worktreePath: "/work/project-feature",
			envValue:     "./project-other",
			composePath:  "/docker",
			want:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pathMatchesEnv(tt.worktreePath, tt.envValue, tt.composePath)
			if got != tt.want {
				t.Errorf("pathMatchesEnv(%q, %q, %q) = %v, want %v",
					tt.worktreePath, tt.envValue, tt.composePath, got, tt.want)
			}
		})
	}
}
