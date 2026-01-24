package cli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func runRm(args []string, out io.Writer) int {
	opts, ok := parseRmArgs(args)
	if !ok {
		return printCommandUsage(out, "rm")
	}
	if opts.all {
		return runRmAll(opts.force, out)
	}
	return runRmBranch(opts.branch, out)
}

type rmOptions struct {
	all    bool
	force  bool
	branch string
}

func parseRmArgs(args []string) (rmOptions, bool) {
	var opts rmOptions
	for _, arg := range args {
		switch arg {
		case "--all":
			opts.all = true
		case "-f", "--force":
			opts.force = true
		default:
			if strings.HasPrefix(arg, "-") {
				return rmOptions{}, false
			}
			if opts.branch != "" {
				return rmOptions{}, false
			}
			opts.branch = arg
		}
	}
	if opts.all && opts.branch != "" {
		return rmOptions{}, false
	}
	if !opts.all && opts.branch == "" {
		return rmOptions{}, false
	}
	if opts.force && !opts.all {
		return rmOptions{}, false
	}
	return opts, true
}

func runRmBranch(branch string, out io.Writer) int {
	root, err := gitRoot()
	if err != nil {
		fmt.Fprintln(out, "wt rm: failed to determine git root")
		return 1
	}

	defaultBranch, err := gitDefaultBranch(root)
	if err == nil && defaultBranch != "" && defaultBranch == branch {
		fmt.Fprintln(out, "wt rm: cannot remove default branch (protected)")
		return 1
	}

	currentBranch, err := gitCurrentBranch(root)
	if err == nil && currentBranch != "" && currentBranch == branch {
		fmt.Fprintln(out, "wt rm: cannot remove current branch (protected)")
		return 1
	}

	path, ok, err := worktreePathByBranch(root, branch)
	if err != nil {
		fmt.Fprintln(out, "wt rm: failed to list worktrees")
		return 1
	}
	if !ok {
		exists, err := refExists(root, "refs/heads/"+branch)
		if err != nil {
			fmt.Fprintln(out, "wt rm: failed to check branch")
			return 1
		}
		if !exists {
			fmt.Fprintln(out, "wt rm: worktree not found")
			return 1
		}
		if err := gitBranchDelete(root, branch); err != nil {
			fmt.Fprintln(out, "wt rm: failed to delete branch")
			return 1
		}
		fmt.Fprintf(out, "wt rm: removed %s\n", branch)
		return 0
	}

	if err := gitWorktreeRemove(root, path); err != nil {
		fmt.Fprintln(out, "wt rm: failed to remove worktree")
		return 1
	}
	if err := gitBranchDelete(root, branch); err != nil {
		exists, checkErr := refExists(root, "refs/heads/"+branch)
		if checkErr != nil || exists {
			fmt.Fprintln(out, "wt rm: failed to delete branch")
			return 1
		}
	}
	fmt.Fprintf(out, "wt rm: removed %s\n", branch)
	return 0
}

func runRmAll(force bool, out io.Writer) int {
	root, err := gitRoot()
	if err != nil {
		fmt.Fprintln(out, "wt rm: failed to determine git root")
		return 1
	}

	mainPath, err := mainWorktreePath(root)
	if err != nil {
		mainPath = root
	}
	if !samePath(root, mainPath) {
		fmt.Fprintf(out, "wt rm: move to main worktree (%s) to run --all\n", mainPath)
		return 1
	}

	worktrees, err := gitWorktrees(root)
	if err != nil {
		fmt.Fprintln(out, "wt rm: failed to list worktrees")
		return 1
	}

	defaultBranch, _ := gitDefaultBranch(root)
	currentBranch, _ := gitCurrentBranch(root)

	var targets []worktreeEntry
	for _, wt := range worktrees {
		if wt.path == mainPath {
			continue
		}
		if wt.branch != "" {
			if defaultBranch != "" && wt.branch == defaultBranch {
				continue
			}
			if currentBranch != "" && wt.branch == currentBranch {
				continue
			}
		}
		targets = append(targets, wt)
	}

	if len(targets) == 0 {
		fmt.Fprintln(out, "wt rm: no worktrees to remove")
		return 0
	}

	if !force {
		if !confirmRmAll(os.Stdin, out, len(targets)) {
			fmt.Fprintln(out, "wt rm: canceled")
			return 1
		}
	}

	for _, wt := range targets {
		if err := removeWorktreeAndBranch(root, wt); err != nil {
			if wt.branch != "" {
				fmt.Fprintf(out, "wt rm: failed to remove %s\n", wt.branch)
			} else {
				fmt.Fprintf(out, "wt rm: failed to remove %s\n", wt.path)
			}
			return 1
		}
		if wt.branch != "" {
			fmt.Fprintf(out, "wt rm: removed %s\n", wt.branch)
		} else {
			fmt.Fprintf(out, "wt rm: removed %s\n", wt.path)
		}
	}
	return 0
}

func samePath(a string, b string) bool {
	aClean := filepath.Clean(a)
	bClean := filepath.Clean(b)
	if aClean == bClean {
		return true
	}
	aEval, err := filepath.EvalSymlinks(aClean)
	if err != nil {
		return false
	}
	bEval, err := filepath.EvalSymlinks(bClean)
	if err != nil {
		return false
	}
	return aEval == bEval
}

func confirmRmAll(in io.Reader, out io.Writer, count int) bool {
	fmt.Fprintf(out, "wt rm: remove %d worktrees? [y/N] ", count)
	reader := bufio.NewReader(in)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes"
}

func removeWorktreeAndBranch(root string, wt worktreeEntry) error {
	if err := gitWorktreeRemove(root, wt.path); err != nil {
		return err
	}
	if wt.branch == "" {
		return nil
	}
	if err := gitBranchDelete(root, wt.branch); err != nil {
		exists, checkErr := refExists(root, "refs/heads/"+wt.branch)
		if checkErr != nil || exists {
			return err
		}
	}
	return nil
}

func gitBranchDelete(root string, branch string) error {
	cmd := exec.Command("git", "branch", "-D", branch)
	cmd.Dir = root
	return cmd.Run()
}

func gitDefaultBranch(root string) (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "--quiet", "refs/remotes/origin/HEAD")
	cmd.Dir = root
	output, err := cmd.Output()
	if err == nil {
		ref := strings.TrimSpace(string(output))
		return strings.TrimPrefix(ref, "refs/remotes/origin/"), nil
	}

	branch, err := mainWorktreeBranch(root)
	if err == nil && branch != "" {
		return branch, nil
	}

	return "", err
}

func gitCurrentBranch(root string) (string, error) {
	cmd := exec.Command("git", "branch", "--show-current")
	cmd.Dir = root
	output, err := cmd.Output()
	if err == nil {
		branch := strings.TrimSpace(string(output))
		if branch != "" {
			return branch, nil
		}
	}

	cmd = exec.Command("git", "symbolic-ref", "--quiet", "HEAD")
	cmd.Dir = root
	output, err = cmd.Output()
	if err != nil {
		return "", err
	}
	ref := strings.TrimSpace(string(output))
	return strings.TrimPrefix(ref, "refs/heads/"), nil
}
