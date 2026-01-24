package cli

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func runPostCreateHooks(root string, worktreePath string) error {
	cfg, err := loadHooks(root)
	if err != nil {
		return err
	}
	if !cfg.hasAny() {
		return nil
	}
	for _, item := range cfg.copy {
		if err := copyHookPath(root, worktreePath, item); err != nil {
			return err
		}
	}
	for _, item := range cfg.symlink {
		if err := symlinkHookPath(root, worktreePath, item); err != nil {
			return err
		}
	}
	for _, cmd := range cfg.run {
		if err := runHookCommand(worktreePath, cmd); err != nil {
			return err
		}
	}
	return nil
}

func copyHookPath(root string, worktreePath string, item string) error {
	if filepath.IsAbs(item) {
		return fmt.Errorf("absolute path not supported: %s", item)
	}
	src := filepath.Join(root, item)
	dst := filepath.Join(worktreePath, item)
	return copyPath(src, dst)
}

func symlinkHookPath(root string, worktreePath string, item string) error {
	if filepath.IsAbs(item) {
		return fmt.Errorf("absolute path not supported: %s", item)
	}
	src := filepath.Join(root, item)
	dst := filepath.Join(worktreePath, item)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.Symlink(src, dst)
}

func runHookCommand(worktreePath string, command string) error {
	cmd := exec.Command("sh", "-c", command)
	cmd.Dir = worktreePath
	return cmd.Run()
}

func copyPath(src string, dst string) error {
	info, err := os.Lstat(src)
	if err != nil {
		return err
	}
	switch {
	case info.Mode()&os.ModeSymlink != 0:
		return copySymlink(src, dst)
	case info.IsDir():
		return copyDir(src, dst, info.Mode().Perm())
	default:
		return copyFile(src, dst, info.Mode().Perm())
	}
}

func copySymlink(src string, dst string) error {
	target, err := os.Readlink(src)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return os.Symlink(target, dst)
}

func copyDir(src string, dst string, perm os.FileMode) error {
	if err := os.MkdirAll(dst, perm); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		entrySrc := filepath.Join(src, entry.Name())
		entryDst := filepath.Join(dst, entry.Name())
		if err := copyPath(entrySrc, entryDst); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(src string, dst string, perm os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, perm)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
