package cli

import (
	"bufio"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

const defaultBaseDir = "../worktrees/{gitroot}"
const defaultConfig = "version: 1\nbase_dir: \"" + defaultBaseDir + "\"\n\n# hooks:\n#   post_create:\n#     copy:\n#       - .env\n#       - .env.local\n#     symlink:\n#       - storage\n#     run:\n#       - bundle install\n"

type hooks struct {
	copy    []string
	symlink []string
	run     []string
}

func loadBaseDir(root string) (string, error) {
	if root != "" {
		baseDir, err := readBaseDirFromConfig(filepath.Join(root, ".wt.yaml"))
		if err != nil {
			return "", err
		}
		if baseDir != "" {
			return baseDir, nil
		}
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return defaultBaseDir, nil
	}
	baseDir, err := readBaseDirFromConfig(filepath.Join(home, ".wt.yaml"))
	if err != nil {
		return "", err
	}
	if baseDir != "" {
		return baseDir, nil
	}
	return defaultBaseDir, nil
}

func readBaseDirFromConfig(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	return parseBaseDir(string(content))
}

func parseBaseDir(content string) (string, error) {
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "base_dir:") {
			value := strings.TrimSpace(strings.TrimPrefix(line, "base_dir:"))
			value = strings.Trim(value, "\"'")
			return value, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", nil
}

func loadHooks(root string) (hooks, error) {
	local, err := readHooksFromConfig(filepath.Join(root, ".wt.yaml"))
	if err != nil {
		return hooks{}, err
	}
	if local.hasAny() {
		return local, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return hooks{}, nil
	}
	return readHooksFromConfig(filepath.Join(home, ".wt.yaml"))
}

func readHooksFromConfig(path string) (hooks, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return hooks{}, nil
		}
		return hooks{}, err
	}
	return parseHooks(string(content))
}

func parseHooks(content string) (hooks, error) {
	var out hooks
	scanner := bufio.NewScanner(strings.NewReader(content))
	inHooks := false
	inPostCreate := false
	currentList := ""

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		indent := len(line) - len(strings.TrimLeft(line, " "))
		if indent == 0 && trimmed != "hooks:" {
			inHooks = false
			inPostCreate = false
			currentList = ""
		}
		if trimmed == "hooks:" {
			inHooks = true
			inPostCreate = false
			currentList = ""
			continue
		}
		if !inHooks {
			continue
		}
		if trimmed == "post_create:" {
			inPostCreate = true
			currentList = ""
			continue
		}
		if !inPostCreate {
			continue
		}
		switch trimmed {
		case "copy:":
			currentList = "copy"
			continue
		case "symlink:":
			currentList = "symlink"
			continue
		case "run:":
			currentList = "run"
			continue
		}
		if strings.HasPrefix(trimmed, "- ") && currentList != "" {
			value := strings.TrimSpace(strings.TrimPrefix(trimmed, "- "))
			value = strings.Trim(value, "\"'")
			switch currentList {
			case "copy":
				out.copy = append(out.copy, value)
			case "symlink":
				out.symlink = append(out.symlink, value)
			case "run":
				out.run = append(out.run, value)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return hooks{}, err
	}
	return out, nil
}

func (h hooks) hasAny() bool {
	return len(h.copy) > 0 || len(h.symlink) > 0 || len(h.run) > 0
}

func worktreePathForBranch(baseDir string, root string, branch string) (string, error) {
	resolved, err := resolveBaseDir(baseDir, root)
	if err != nil {
		return "", err
	}
	return filepath.Join(resolved, filepath.FromSlash(branch)), nil
}

func resolveBaseDir(baseDir string, root string) (string, error) {
	value := strings.TrimSpace(baseDir)
	if value == "" {
		value = defaultBaseDir
	}
	value = strings.ReplaceAll(value, "{gitroot}", filepath.Base(filepath.Clean(root)))
	if value == "~" || strings.HasPrefix(value, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		if value == "~" {
			value = home
		} else {
			value = filepath.Join(home, value[2:])
		}
	}
	if filepath.IsAbs(value) {
		return filepath.Clean(value), nil
	}
	return filepath.Clean(filepath.Join(root, value)), nil
}
