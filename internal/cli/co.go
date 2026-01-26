package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func runCo(args []string, out io.Writer) int {
	if len(args) != 1 {
		return printCommandUsage(out, "co")
	}

	root, err := repoRoot()
	if err != nil {
		fmt.Fprintln(out, "wt co: failed to determine git root")
		return 1
	}

	if path, ok, err := worktreePathByBranch(root, args[0]); err != nil {
		fmt.Fprintln(out, "wt co: failed to list worktrees")
		return 1
	} else if ok {
		fmt.Fprintln(out, path)
		return 0
	}

	branch := args[0]
	status, err := branchStatus(root, branch)
	if err != nil {
		fmt.Fprintln(out, "wt co: failed to check branch")
		return 1
	}
	if len(status.remoteMatches) > 1 {
		fmt.Fprintf(out, "wt co: branch %q exists in multiple remotes: %s\n", branch, strings.Join(status.remoteMatches, ", "))
		return 1
	}

	baseDir, err := loadBaseDir(root)
	if err != nil {
		fmt.Fprintln(out, "wt co: failed to load config")
		return 1
	}
	worktreePath, err := worktreePathForBranch(baseDir, root, branch)
	if err != nil {
		fmt.Fprintln(out, "wt co: failed to resolve worktree path")
		return 1
	}

	if err := os.MkdirAll(filepath.Dir(worktreePath), 0o755); err != nil {
		fmt.Fprintln(out, "wt co: failed to prepare worktree directory")
		return 1
	}

	if err := gitWorktreeAdd(root, worktreePath, branch, status); err != nil {
		fmt.Fprintln(out, "wt co: failed to create worktree")
		return 1
	}

	if err := runPostCreateHooks(root, worktreePath); err != nil {
		fmt.Fprintln(out, "wt co: failed to run hooks")
		return 1
	}

	message := fmt.Sprintf("wt co: created worktree %s", worktreePath)
	if !status.local {
		message += " (branch created)"
	}
	fmt.Fprintln(out, message)
	fmt.Fprintln(out, worktreePath)
	return 0
}
