package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func runInit(args []string, out io.Writer) int {
	global := false
	if len(args) > 0 {
		if len(args) == 1 && args[0] == "-g" {
			global = true
		} else {
			return printCommandUsage(out, "init")
		}
	}

	path, err := initPath(global)
	if err != nil {
		fmt.Fprintln(out, "wt init: failed to determine config path")
		return 1
	}
	if err := writeConfig(path); err != nil {
		fmt.Fprintf(out, "wt init: %v\n", err)
		return 1
	}
	fmt.Fprintf(out, "wt init: created %s\n", path)
	return 0
}

func initPath(global bool) (string, error) {
	if global {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".wt.yaml"), nil
	}
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	return filepath.Join(root, ".wt.yaml"), nil
}

func writeConfig(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.WriteString(file, defaultConfig)
	return err
}
