package cli

import (
	"fmt"
	"os/exec"
	"sort"
	"strings"
)

func gitRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	root := strings.TrimSpace(string(output))
	if root == "" {
		return "", fmt.Errorf("empty git root")
	}
	return root, nil
}

func repoRoot() (string, error) {
	root, err := gitRoot()
	if err != nil {
		return "", err
	}
	mainRoot, err := mainWorktreePath(root)
	if err == nil && mainRoot != "" {
		return mainRoot, nil
	}
	return root, nil
}

func refExists(root string, ref string) (bool, error) {
	cmd := exec.Command("git", "show-ref", "--verify", "--quiet", ref)
	cmd.Dir = root
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 1 {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func gitRemotes(root string) ([]string, error) {
	cmd := exec.Command("git", "remote")
	cmd.Dir = root
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 1 && lines[0] == "" {
		return nil, nil
	}
	sort.Strings(lines)
	return lines, nil
}
