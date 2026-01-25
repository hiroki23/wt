package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCdDash(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"cd", "-"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	if got := out.String(); got != "-\n" {
		t.Fatalf("expected stdout to be \"-\\n\", got %q", got)
	}
}

func TestRunCdNoArgsUsesMainWorktree(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	root := t.TempDir()
	if err := initRepo(root, "main"); err != nil {
		t.Fatalf("init repo failed: %v", err)
	}

	worktreePath, err := expectedWorktreePath(root, "feature/one")
	if err != nil {
		t.Fatalf("expected path failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := runCmd(root, "git", "worktree", "add", "-b", "feature/one", worktreePath); err != nil {
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
	if err := os.Chdir(worktreePath); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}

	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"cd"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	got := strings.TrimSpace(out.String())
	gotNorm, err := normalizePath(got)
	if err != nil {
		t.Fatalf("normalize output failed: %v", err)
	}
	wantNorm, err := normalizePath(root)
	if err != nil {
		t.Fatalf("normalize root failed: %v", err)
	}
	if gotNorm != wantNorm {
		t.Fatalf("expected stdout to be %q, got %q", wantNorm+"\n", gotNorm+"\n")
	}
}

func TestRunCdBranchOutputsWorktreePath(t *testing.T) {
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

	code := Run([]string{"cd", "feature/one"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	got := strings.TrimSpace(out.String())
	gotNorm, err := normalizePath(got)
	if err != nil {
		t.Fatalf("normalize output failed: %v", err)
	}
	wantNorm, err := normalizePath(path)
	if err != nil {
		t.Fatalf("normalize path failed: %v", err)
	}
	if gotNorm != wantNorm {
		t.Fatalf("expected stdout to be %q, got %q", wantNorm+"\n", gotNorm+"\n")
	}
}

func TestRunCdAliasOutputsWorktreePath(t *testing.T) {
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

	code := Run([]string{"feature/one"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	got := strings.TrimSpace(out.String())
	gotNorm, err := normalizePath(got)
	if err != nil {
		t.Fatalf("normalize output failed: %v", err)
	}
	wantNorm, err := normalizePath(path)
	if err != nil {
		t.Fatalf("normalize path failed: %v", err)
	}
	if gotNorm != wantNorm {
		t.Fatalf("expected stdout to be %q, got %q", wantNorm+"\n", gotNorm+"\n")
	}
}
