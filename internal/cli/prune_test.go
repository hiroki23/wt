package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPruneRunsGitWorktreePrune(t *testing.T) {
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
	if err := os.RemoveAll(path); err != nil {
		t.Fatalf("remove worktree dir failed: %v", err)
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

	code := Run([]string{"prune"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	if got := strings.TrimSpace(out.String()); got != "wt prune: done" {
		t.Fatalf("unexpected stdout: %q", got)
	}

	worktrees, err := gitWorktrees(root)
	if err != nil {
		t.Fatalf("list worktrees failed: %v", err)
	}
	for _, wt := range worktrees {
		if wt.branch == "feature/one" {
			t.Fatalf("expected pruned worktree to be removed from list")
		}
	}
}
