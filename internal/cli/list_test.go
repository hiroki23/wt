package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunListShowsWorktrees(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	t.Setenv("HOME", t.TempDir())

	root := t.TempDir()
	if err := initRepo(root, "main"); err != nil {
		t.Fatalf("init repo failed: %v", err)
	}

	path, err := expectedWorktreePath(root, "feature/one")
	if err != nil {
		t.Fatalf("expected path failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := runCmd(root, "git", "worktree", "add", "-b", "feature/one", path); err != nil {
		t.Fatalf("git worktree add failed: %v", err)
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

	code := Run([]string{"list"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}

	lines := strings.Split(strings.TrimSpace(out.String()), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected header + rows, got %q", out.String())
	}
	if lines[0] != "PATH\tBRANCH" {
		t.Fatalf("unexpected header: %q", lines[0])
	}

	entries := make(map[string]string)
	for _, line := range lines[1:] {
		parts := strings.SplitN(line, "\t", 2)
		if len(parts) != 2 {
			t.Fatalf("unexpected row: %q", line)
		}
		norm, err := normalizePath(parts[0])
		if err != nil {
			t.Fatalf("normalize path failed: %v", err)
		}
		entries[norm] = parts[1]
	}

	rootNorm, err := normalizePath(root)
	if err != nil {
		t.Fatalf("normalize root failed: %v", err)
	}
	pathNorm, err := normalizePath(path)
	if err != nil {
		t.Fatalf("normalize path failed: %v", err)
	}

	if got := entries[rootNorm]; got != "main" {
		t.Fatalf("expected root branch main, got %q", got)
	}
	if got := entries[pathNorm]; got != "feature/one" {
		t.Fatalf("expected feature branch, got %q", got)
	}
}
