package cli

import (
	"fmt"
	"io"
)

func runCd(args []string, out io.Writer) int {
	if len(args) == 0 {
		root, err := gitRoot()
		if err != nil {
			fmt.Fprintln(out, "wt cd: failed to determine git root")
			return 1
		}
		mainPath, err := mainWorktreePath(root)
		if err != nil {
			fmt.Fprintln(out, "wt cd: failed to determine main worktree")
			return 1
		}
		fmt.Fprintln(out, mainPath)
		return 0
	}
	if len(args) == 1 && args[0] == "-" {
		fmt.Fprintln(out, "-")
		return 0
	}
	if len(args) == 1 {
		root, err := gitRoot()
		if err != nil {
			fmt.Fprintln(out, "wt cd: failed to determine git root")
			return 1
		}
		path, ok, err := worktreePathByBranch(root, args[0])
		if err != nil {
			fmt.Fprintln(out, "wt cd: failed to list worktrees")
			return 1
		}
		if !ok {
			fmt.Fprintln(out, "wt cd: worktree not found")
			return 1
		}
		fmt.Fprintln(out, path)
		return 0
	}
	return printCommandUsage(out, "cd")
}
