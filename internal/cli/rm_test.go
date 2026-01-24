package cli

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunRmRemovesWorktreeAndBranch(t *testing.T) {
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

	code := Run([]string{"rm", "feature/one"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected worktree to be removed, but path exists")
	}
	if err := runCmd(root, "git", "show-ref", "--verify", "--quiet", "refs/heads/feature/one"); err == nil {
		t.Fatalf("expected branch to be removed")
	}
}

func TestRunRmDefaultBranchProtected(t *testing.T) {
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

	code := Run([]string{"rm", "main"}, &out, &errOut)

	if code == 0 {
		t.Fatalf("expected non-zero exit code, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	if got := out.String(); got != "wt rm: cannot remove default branch (protected)\n" {
		t.Fatalf("expected default branch message, got %q", got)
	}
	if err := runCmd(root, "git", "show-ref", "--verify", "--quiet", "refs/heads/main"); err != nil {
		t.Fatalf("expected default branch to remain")
	}
}

func TestRunRmDeletesBranchWhenWorktreeMissing(t *testing.T) {
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

	code := Run([]string{"rm", "feature/one"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	if err := runCmd(root, "git", "show-ref", "--verify", "--quiet", "refs/heads/feature/one"); err == nil {
		t.Fatalf("expected branch to be removed")
	}
}

func TestRunRmCurrentBranchProtected(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	t.Setenv("HOME", t.TempDir())

	root := t.TempDir()
	if err := initRepo(root, "main"); err != nil {
		t.Fatalf("init repo failed: %v", err)
	}
	if err := runCmd(root, "git", "checkout", "-b", "feature/one"); err != nil {
		t.Fatalf("git checkout failed: %v", err)
	}

	originDir := filepath.Join(root, ".git", "refs", "remotes", "origin")
	if err := os.MkdirAll(originDir, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := runCmd(root, "git", "symbolic-ref", "refs/remotes/origin/HEAD", "refs/remotes/origin/main"); err != nil {
		t.Fatalf("git symbolic-ref failed: %v", err)
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

	code := Run([]string{"rm", "feature/one"}, &out, &errOut)

	if code == 0 {
		t.Fatalf("expected non-zero exit code, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	if got := out.String(); got != "wt rm: cannot remove current branch (protected)\n" {
		t.Fatalf("expected current branch message, got %q", got)
	}
	if err := runCmd(root, "git", "show-ref", "--verify", "--quiet", "refs/heads/feature/one"); err != nil {
		t.Fatalf("expected current branch to remain")
	}
}

func TestRunRmAllRemovesWorktreesWithForce(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	t.Setenv("HOME", t.TempDir())

	root := t.TempDir()
	if err := initRepo(root, "main"); err != nil {
		t.Fatalf("init repo failed: %v", err)
	}

	pathOne, err := expectedWorktreePath(root, "feature/one")
	if err != nil {
		t.Fatalf("expected path failed: %v", err)
	}
	pathTwo, err := expectedWorktreePath(root, "feature/two")
	if err != nil {
		t.Fatalf("expected path failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(pathOne), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := runCmd(root, "git", "worktree", "add", "-b", "feature/one", pathOne); err != nil {
		t.Fatalf("git worktree add failed: %v", err)
	}
	if err := runCmd(root, "git", "worktree", "add", "-b", "feature/two", pathTwo); err != nil {
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

	code := Run([]string{"rm", "--all", "-f"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	if _, err := os.Stat(pathOne); err == nil {
		t.Fatalf("expected worktree one to be removed, but path exists")
	}
	if _, err := os.Stat(pathTwo); err == nil {
		t.Fatalf("expected worktree two to be removed, but path exists")
	}
	if err := runCmd(root, "git", "show-ref", "--verify", "--quiet", "refs/heads/feature/one"); err == nil {
		t.Fatalf("expected branch feature/one to be removed")
	}
	if err := runCmd(root, "git", "show-ref", "--verify", "--quiet", "refs/heads/feature/two"); err == nil {
		t.Fatalf("expected branch feature/two to be removed")
	}
	if err := runCmd(root, "git", "show-ref", "--verify", "--quiet", "refs/heads/main"); err != nil {
		t.Fatalf("expected default branch to remain")
	}
}

func TestRunRmAllCancelKeepsWorktrees(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	t.Setenv("HOME", t.TempDir())

	root := t.TempDir()
	if err := initRepo(root, "main"); err != nil {
		t.Fatalf("init repo failed: %v", err)
	}

	pathOne, err := expectedWorktreePath(root, "feature/one")
	if err != nil {
		t.Fatalf("expected path failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(pathOne), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := runCmd(root, "git", "worktree", "add", "-b", "feature/one", pathOne); err != nil {
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

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe failed: %v", err)
	}
	if _, err := writer.WriteString("n\n"); err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close write failed: %v", err)
	}
	oldStdin := os.Stdin
	os.Stdin = reader
	defer func() {
		os.Stdin = oldStdin
	}()

	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"rm", "--all"}, &out, &errOut)

	if code == 0 {
		t.Fatalf("expected non-zero exit code, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	if _, err := os.Stat(pathOne); err != nil {
		t.Fatalf("expected worktree to remain, got %v", err)
	}
	if err := runCmd(root, "git", "show-ref", "--verify", "--quiet", "refs/heads/feature/one"); err != nil {
		t.Fatalf("expected branch feature/one to remain")
	}
}

func TestRunRmAllSkipsCurrentBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	t.Setenv("HOME", t.TempDir())

	root := t.TempDir()
	if err := initRepo(root, "main"); err != nil {
		t.Fatalf("init repo failed: %v", err)
	}

	pathOne, err := expectedWorktreePath(root, "feature/one")
	if err != nil {
		t.Fatalf("expected path failed: %v", err)
	}
	pathTwo, err := expectedWorktreePath(root, "feature/two")
	if err != nil {
		t.Fatalf("expected path failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(pathOne), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := runCmd(root, "git", "worktree", "add", "-b", "feature/one", pathOne); err != nil {
		t.Fatalf("git worktree add failed: %v", err)
	}
	if err := runCmd(root, "git", "worktree", "add", "-b", "feature/two", pathTwo); err != nil {
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

	code := Run([]string{"rm", "--all", "-f"}, &out, &errOut)

	if code == 0 {
		t.Fatalf("expected non-zero exit code, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	got := out.String()
	prefix := "wt rm: move to main worktree ("
	suffix := ") to run --all\n"
	if !strings.HasPrefix(got, prefix) || !strings.HasSuffix(got, suffix) {
		t.Fatalf("unexpected message %q", got)
	}
	gotPath := strings.TrimSuffix(strings.TrimPrefix(got, prefix), suffix)
	gotNorm, err := normalizePath(gotPath)
	if err != nil {
		t.Fatalf("normalize output failed: %v", err)
	}
	wantNorm, err := normalizePath(root)
	if err != nil {
		t.Fatalf("normalize root failed: %v", err)
	}
	if gotNorm != wantNorm {
		t.Fatalf("expected main path %q, got %q", wantNorm, gotNorm)
	}
	if _, err := os.Stat(pathTwo); err != nil {
		t.Fatalf("expected worktree two to remain, got %v", err)
	}
	if _, err := os.Stat(pathOne); err != nil {
		t.Fatalf("expected worktree one to remain, got %v", err)
	}
	if err := runCmd(root, "git", "show-ref", "--verify", "--quiet", "refs/heads/feature/two"); err != nil {
		t.Fatalf("expected branch feature/two to remain")
	}
	if err := runCmd(root, "git", "show-ref", "--verify", "--quiet", "refs/heads/feature/one"); err != nil {
		t.Fatalf("expected branch feature/one to remain")
	}
}
