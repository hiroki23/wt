package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRunAddNewBranchPrintsMessage(t *testing.T) {
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

	code := Run([]string{"add", "feature/one"}, &out, &errOut)

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
	expectedOutput := "wt add: created worktree " + expectedPath + " (branch created)\n"
	if got := out.String(); got != expectedOutput {
		t.Fatalf("unexpected stdout: %q", got)
	}
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected worktree to exist: %v", err)
	}
}

func TestRunAddExistingBranchDoesNotPrintCreateMessage(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	t.Setenv("HOME", t.TempDir())

	root := t.TempDir()
	if err := initRepo(root, "main"); err != nil {
		t.Fatalf("init repo failed: %v", err)
	}
	if err := runCmd(root, "git", "branch", "feature/one"); err != nil {
		t.Fatalf("git branch failed: %v", err)
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

	code := Run([]string{"add", "feature/one"}, &out, &errOut)

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
	expectedOutput := "wt add: created worktree " + expectedPath + "\n"
	if got := out.String(); got != expectedOutput {
		t.Fatalf("unexpected stdout: %q", got)
	}
	if _, err := os.Stat(expectedPath); err != nil {
		t.Fatalf("expected worktree to exist: %v", err)
	}
}

func TestRunAddRemoteBranchConflictErrors(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	t.Setenv("HOME", t.TempDir())

	root := t.TempDir()
	if err := initRepo(root, "main"); err != nil {
		t.Fatalf("init repo failed: %v", err)
	}
	if err := runCmd(root, "git", "remote", "add", "origin", "https://example.com/origin.git"); err != nil {
		t.Fatalf("git remote add origin failed: %v", err)
	}
	if err := runCmd(root, "git", "remote", "add", "upstream", "https://example.com/upstream.git"); err != nil {
		t.Fatalf("git remote add upstream failed: %v", err)
	}
	if err := runCmd(root, "git", "update-ref", "refs/remotes/origin/feature/one", "HEAD"); err != nil {
		t.Fatalf("git update-ref origin failed: %v", err)
	}
	if err := runCmd(root, "git", "update-ref", "refs/remotes/upstream/feature/one", "HEAD"); err != nil {
		t.Fatalf("git update-ref upstream failed: %v", err)
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

	code := Run([]string{"add", "feature/one"}, &out, &errOut)

	if code == 0 {
		t.Fatalf("expected non-zero exit code, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	if got := out.String(); got != "wt add: branch \"feature/one\" exists in multiple remotes: origin, upstream\n" {
		t.Fatalf("unexpected stdout: %q", got)
	}
}

func TestRunAddExistingWorktreeErrors(t *testing.T) {
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

	code := Run([]string{"add", "feature/one"}, &out, &errOut)

	if code == 0 {
		t.Fatalf("expected non-zero exit code, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	expectedOutput := "wt add: worktree already exists " + path + "\n"
	if got := out.String(); got != expectedOutput {
		t.Fatalf("unexpected stdout: %q", got)
	}
}

func TestRunAddRunsPostCreateHooks(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	t.Setenv("HOME", t.TempDir())

	root := t.TempDir()
	if err := initRepo(root, "main"); err != nil {
		t.Fatalf("init repo failed: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, ".env"), []byte("secret=1\n"), 0o644); err != nil {
		t.Fatalf("write env failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "storage"), 0o755); err != nil {
		t.Fatalf("mkdir storage failed: %v", err)
	}

	config := `version: 1
base_dir: "../worktrees/{gitroot}"
hooks:
  post_create:
    copy:
      - ".env"
    symlink:
      - "storage"
    run:
      - "touch .hooked"
`
	if err := os.WriteFile(filepath.Join(root, ".wt.yaml"), []byte(config), 0o644); err != nil {
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

	code := Run([]string{"add", "feature/one"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}

	worktreePath, err := expectedWorktreePath(root, "feature/one")
	if err != nil {
		t.Fatalf("expected path failed: %v", err)
	}
	envContent, err := os.ReadFile(filepath.Join(worktreePath, ".env"))
	if err != nil {
		t.Fatalf("read .env failed: %v", err)
	}
	if got := string(envContent); got != "secret=1\n" {
		t.Fatalf("unexpected env content: %q", got)
	}

	linkPath := filepath.Join(worktreePath, "storage")
	info, err := os.Lstat(linkPath)
	if err != nil {
		t.Fatalf("stat symlink failed: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("expected storage to be a symlink")
	}
	target, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("readlink failed: %v", err)
	}
	targetNorm, err := normalizePath(target)
	if err != nil {
		t.Fatalf("normalize target failed: %v", err)
	}
	rootStorageNorm, err := normalizePath(filepath.Join(root, "storage"))
	if err != nil {
		t.Fatalf("normalize storage failed: %v", err)
	}
	if targetNorm != rootStorageNorm {
		t.Fatalf("unexpected symlink target: %q", targetNorm)
	}

	if _, err := os.Stat(filepath.Join(worktreePath, ".hooked")); err != nil {
		t.Fatalf("expected .hooked to exist: %v", err)
	}
}

func TestRunAddHelpShowsUsage(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"add", "-h"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	if got := out.String(); got != "Usage: wt add <branch>\n\n  Create a worktree (and branch if needed).\n\nExamples:\n  wt add feat/one\n" {
		t.Fatalf("unexpected stdout: %q", got)
	}
}
