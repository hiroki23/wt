package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type worktreeEntry struct {
	path   string
	branch string
}

func worktreePathByBranch(root string, branch string) (string, bool, error) {
	worktrees, err := gitWorktrees(root)
	if err != nil {
		return "", false, err
	}
	for _, wt := range worktrees {
		if wt.branch == branch {
			return wt.path, true, nil
		}
	}
	return "", false, nil
}

func gitWorktrees(root string) ([]worktreeEntry, error) {
	cmd := exec.Command("git", "worktree", "list", "--porcelain")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}

	var entries []worktreeEntry
	var current *worktreeEntry
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "worktree ") {
			if current != nil {
				entries = append(entries, *current)
			}
			path := strings.TrimSpace(strings.TrimPrefix(line, "worktree "))
			current = &worktreeEntry{path: path}
			continue
		}
		if strings.HasPrefix(line, "branch ") && current != nil {
			ref := strings.TrimSpace(strings.TrimPrefix(line, "branch "))
			current.branch = strings.TrimPrefix(ref, "refs/heads/")
		}
	}
	if current != nil {
		entries = append(entries, *current)
	}
	return entries, nil
}

func gitWorktreeAdd(root string, path string, branch string, status branchStatusInfo) error {
	var args []string
	switch {
	case status.local:
		args = []string{"worktree", "add", path, branch}
	case len(status.remoteMatches) == 1:
		remoteRef := status.remoteMatches[0] + "/" + branch
		args = []string{"worktree", "add", "-b", branch, path, remoteRef}
	default:
		args = []string{"worktree", "add", "-b", branch, path}
	}
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	return cmd.Run()
}

func gitWorktreeRemove(root string, path string) error {
	cmd := exec.Command("git", "worktree", "remove", path)
	cmd.Dir = root
	return cmd.Run()
}

func gitWorktreePrune(root string) error {
	cmd := exec.Command("git", "worktree", "prune")
	cmd.Dir = root
	return cmd.Run()
}

func mainWorktreePath(root string) (string, error) {
	worktrees, err := gitWorktrees(root)
	if err != nil {
		return "", err
	}
	for _, wt := range worktrees {
		info, err := os.Stat(filepath.Join(wt.path, ".git"))
		if err != nil {
			continue
		}
		if info.IsDir() {
			return wt.path, nil
		}
	}
	return root, nil
}

func mainWorktreeBranch(root string) (string, error) {
	worktrees, err := gitWorktrees(root)
	if err != nil {
		return "", err
	}
	for _, wt := range worktrees {
		info, err := os.Stat(filepath.Join(wt.path, ".git"))
		if err != nil {
			continue
		}
		if info.IsDir() {
			if wt.branch == "" {
				return "", fmt.Errorf("main worktree branch not found")
			}
			return wt.branch, nil
		}
	}
	return "", fmt.Errorf("main worktree not found")
}
