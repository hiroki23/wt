package cli

import (
	"fmt"
	"io"
)

func runPrune(args []string, out io.Writer) int {
	if len(args) != 0 {
		return printCommandUsage(out, "prune")
	}
	root, err := gitRoot()
	if err != nil {
		fmt.Fprintln(out, "wt prune: failed to determine git root")
		return 1
	}
	if err := gitWorktreePrune(root); err != nil {
		fmt.Fprintln(out, "wt prune: failed to prune worktrees")
		return 1
	}
	fmt.Fprintln(out, "wt prune: done")
	return 0
}
