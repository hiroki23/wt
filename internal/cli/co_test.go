package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCoCreatesWorktreeAndOutputsPath(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	t.Setenv("HOME", t.TempDir())

	root := t.TempDir()
	if err := initRepo(root, "main"); err != nil {
		t.Fatalf("init repo failed: %v", err)
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

	code := Run([]string{"co", "feature/one"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}

	expectedPath, err := expectedWorktreePath(root, "feature/one")
	if err != nil {
		t.Fatalf("expected path failed: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	got := strings.TrimSpace(lines[len(lines)-1])
	gotNorm, err := normalizePath(got)
	if err != nil {
		t.Fatalf("normalize output failed: %v", err)
	}
	wantNorm, err := normalizePath(expectedPath)
	if err != nil {
		t.Fatalf("normalize path failed: %v", err)
	}
	if gotNorm != wantNorm {
		t.Fatalf("expected stdout to be %q, got %q", wantNorm+"\n", gotNorm+"\n")
	}
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected worktree to exist: %v", err)
	}
	if !strings.Contains(out.String(), "wt co: created worktree "+expectedPath) {
		t.Fatalf("expected create message, got %q", out.String())
	}
}

func TestRunCoFromWorktreeUsesMainBaseDir(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	t.Setenv("HOME", t.TempDir())

	root := t.TempDir()
	if err := initRepo(root, "main"); err != nil {
		t.Fatalf("init repo failed: %v", err)
	}

	pathOne, err := expectedWorktreePath(root, "test1")
	if err != nil {
		t.Fatalf("expected path failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(pathOne), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := runCmd(root, "git", "worktree", "add", "-b", "test1", pathOne); err != nil {
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
	if err := os.Chdir(pathOne); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"co", "test2"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}

	expectedPath, err := expectedWorktreePath(root, "test2")
	if err != nil {
		t.Fatalf("expected path failed: %v", err)
	}
	lines := strings.Split(strings.TrimRight(out.String(), "\n"), "\n")
	got := strings.TrimSpace(lines[len(lines)-1])
	gotNorm, err := normalizePath(got)
	if err != nil {
		t.Fatalf("normalize output failed: %v", err)
	}
	wantNorm, err := normalizePath(expectedPath)
	if err != nil {
		t.Fatalf("normalize path failed: %v", err)
	}
	if gotNorm != wantNorm {
		t.Fatalf("expected stdout to be %q, got %q", wantNorm+"\n", gotNorm+"\n")
	}
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected worktree to exist: %v", err)
	}
	if !strings.Contains(out.String(), "wt co: created worktree "+expectedPath) {
		t.Fatalf("expected create message, got %q", out.String())
	}
}
