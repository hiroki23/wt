package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func runAdd(args []string, out io.Writer) int {
	if len(args) != 1 {
		return printCommandUsage(out, "add")
	}
	branch := args[0]

	root, err := repoRoot()
	if err != nil {
		fmt.Fprintln(out, "wt add: failed to determine git root")
		return 1
	}

	status, err := branchStatus(root, branch)
	if err != nil {
		fmt.Fprintln(out, "wt add: failed to check branch")
		return 1
	}
	if len(status.remoteMatches) > 1 {
		fmt.Fprintf(out, "wt add: branch %q exists in multiple remotes: %s\n", branch, strings.Join(status.remoteMatches, ", "))
		return 1
	}

	baseDir, err := loadBaseDir(root)
	if err != nil {
		fmt.Fprintln(out, "wt add: failed to load config")
		return 1
	}
	worktreePath, err := worktreePathForBranch(baseDir, root, branch)
	if err != nil {
		fmt.Fprintln(out, "wt add: failed to resolve worktree path")
		return 1
	}

	if err := os.MkdirAll(filepath.Dir(worktreePath), 0o755); err != nil {
		fmt.Fprintln(out, "wt add: failed to prepare worktree directory")
		return 1
	}

	if err := gitWorktreeAdd(root, worktreePath, branch, status); err != nil {
		fmt.Fprintln(out, "wt add: failed to create worktree")
		return 1
	}

	if err := runPostCreateHooks(root, worktreePath); err != nil {
		fmt.Fprintln(out, "wt add: failed to run hooks")
		return 1
	}

	if !status.local {
		fmt.Fprintf(out, "wt add: created new branch %s\n", branch)
	}
	fmt.Fprintf(out, "wt add: created worktree %s\n", worktreePath)
	return 0
}
