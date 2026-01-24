package cli

import (
	"fmt"
	"io"
)

func runList(args []string, out io.Writer) int {
	if len(args) != 0 {
		return printCommandUsage(out, "list")
	}

	root, err := gitRoot()
	if err != nil {
		fmt.Fprintln(out, "wt list: failed to determine git root")
		return 1
	}
	worktrees, err := gitWorktrees(root)
	if err != nil {
		fmt.Fprintln(out, "wt list: failed to list worktrees")
		return 1
	}
	fmt.Fprintln(out, "PATH\tBRANCH")
	for _, wt := range worktrees {
		fmt.Fprintf(out, "%s\t%s\n", wt.path, wt.branch)
	}
	return 0
}
