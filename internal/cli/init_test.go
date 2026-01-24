package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunInitGlobalCreatesConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"init", "-g"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}

	path := filepath.Join(home, ".wt.yaml")
	if got := out.String(); got != fmt.Sprintf("wt init: created %s\n", path) {
		t.Fatalf("unexpected stdout: %q", got)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config failed: %v", err)
	}
	if got := string(content); got != defaultConfig {
		t.Fatalf("unexpected config content: %q", got)
	}
}

func TestRunInitLocalCreatesConfigAtGitRoot(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	root := t.TempDir()
	if err := runCmd(root, "git", "init"); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	subdir := filepath.Join(root, "sub", "dir")
	if err := os.MkdirAll(subdir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd failed: %v", chdirErr)
		}
	}()
	if err := os.Chdir(subdir); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"init"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}

	rootNorm, err := normalizePath(root)
	if err != nil {
		t.Fatalf("normalize root failed: %v", err)
	}
	path := filepath.Join(rootNorm, ".wt.yaml")
	if got := strings.TrimSpace(out.String()); got != "wt init: created "+path {
		t.Fatalf("unexpected stdout: %q", got)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config failed: %v", err)
	}
	if got := string(content); got != defaultConfig {
		t.Fatalf("unexpected config content: %q", got)
	}
}

func TestRunInitLocalDoesNotOverwriteExistingConfig(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	root := t.TempDir()
	if err := runCmd(root, "git", "init"); err != nil {
		t.Fatalf("git init failed: %v", err)
	}

	path := filepath.Join(root, ".wt.yaml")
	existing := "version: 1\nbase_dir: \"../custom\"\n"
	if err := os.WriteFile(path, []byte(existing), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	defer func() {
		if chdirErr := os.Chdir(wd); chdirErr != nil {
			t.Fatalf("restore wd failed: %v", chdirErr)
		}
	}()
	if err := os.Chdir(root); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"init"}, &out, &errOut)

	if code == 0 {
		t.Fatalf("expected non-zero exit code, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	if got := strings.TrimSpace(out.String()); !strings.HasPrefix(got, "wt init: ") {
		t.Fatalf("unexpected stdout: %q", got)
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read config failed: %v", err)
	}
	if got := string(content); got != existing {
		t.Fatalf("expected config to remain unchanged, got %q", got)
	}
}
