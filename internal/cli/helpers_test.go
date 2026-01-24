package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func runCmd(dir string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s %v failed: %w: %s", name, args, err, string(output))
	}
	return nil
}

func normalizePath(path string) (string, error) {
	clean := filepath.Clean(path)
	return filepath.EvalSymlinks(clean)
}

func expectedWorktreePath(root string, branch string) (string, error) {
	rootNorm, err := normalizePath(root)
	if err != nil {
		return "", err
	}
	baseDir := filepath.Join(filepath.Dir(rootNorm), "worktrees", filepath.Base(rootNorm))
	return filepath.Join(baseDir, filepath.FromSlash(branch)), nil
}

func initRepo(tdir string, defaultBranch string) error {
	args := []string{"init"}
	if defaultBranch != "" {
		args = append(args, "-b", defaultBranch)
	}
	if err := runCmd(tdir, "git", args...); err != nil {
		return err
	}
	if err := runCmd(tdir, "git", "config", "user.email", "test@example.com"); err != nil {
		return err
	}
	if err := runCmd(tdir, "git", "config", "user.name", "Test User"); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(tdir, "README.md"), []byte("init\n"), 0o644); err != nil {
		return err
	}
	if err := runCmd(tdir, "git", "add", "README.md"); err != nil {
		return err
	}
	return runCmd(tdir, "git", "commit", "-m", "init")
}
