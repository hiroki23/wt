package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type command struct {
	name     string
	usage    string
	short    string
	help     string
	examples []string
	run      func(args []string, out io.Writer) int
}

var Version = "dev"

func Run(args []string, out io.Writer, errOut io.Writer) int {
	if len(args) == 0 || isHelp(args[0]) {
		return printHelp(out)
	}
	if len(args) == 1 && isVersion(args[0]) {
		return runVersion(out)
	}

	cmds := commands()
	lookup := args[0]
	if alias, ok := aliasCommands()[lookup]; ok {
		lookup = alias
	}

	for _, cmd := range cmds {
		if cmd.name == lookup {
			if hasHelpFlag(args[1:]) {
				return printCommandHelp(out, cmd.name)
			}
			return cmd.run(args[1:], out)
		}
	}

	if len(args) == 1 && !strings.HasPrefix(args[0], "-") {
		return runCd(args, out)
	}

	fmt.Fprintf(errOut, "wt: unknown command: %s\n", args[0])
	fmt.Fprint(errOut, "Run 'wt help' for usage.\n")
	return 2
}

func commands() []command {
	cmds := []command{
		{
			name:  "init",
			usage: "wt init [-g]",
			short: "create .wt.yaml (local or global)",
			help:  "Create .wt.yaml in the repo or HOME (-g).",
			examples: []string{
				"wt init",
				"wt init -g",
			},
			run:   runInit,
		},
		{
			name:  "hook",
			usage: "wt hook <shell> [--prompt]",
			short: "print shell integration script",
			help:  "Supported shells: zsh, bash, fish. Prompt is zsh-only.",
			examples: []string{
				`eval "$(wt hook zsh)"`,
				"wt hook zsh --prompt",
			},
			run:   runHook,
		},
		{
			name:  "add",
			usage: "wt add <branch>",
			short: "create a worktree for a branch",
			help:  "Create a worktree (and branch if needed).",
			examples: []string{
				"wt add feat/one",
			},
			run:   runAdd,
		},
		{
			name:  "co",
			usage: "wt co <branch>",
			short: "create a worktree and move",
			help:  "Create the worktree if needed, then move to it.",
			examples: []string{
				"wt co feat/one",
			},
			run:   runCo,
		},
		{
			name:  "cd",
			usage: "wt cd [branch|-]",
			short: "move to a worktree",
			help:  "No args: main worktree (git root). '-' goes back. Shorthand: wt <branch>.",
			examples: []string{
				"wt cd feat/one",
				"wt cd -",
				"wt cd",
			},
			run:   runCd,
		},
		{
			name:  "rm",
			usage: "wt rm <branch> | wt rm --all [-f]",
			short: "remove worktree and branch",
			help:  "Remove worktrees and branches. --all requires main worktree (git root).",
			examples: []string{
				"wt rm feat/one",
				"wt rm --all",
				"wt rm --all -f",
			},
			run:   runRm,
		},
		{
			name:  "list",
			usage: "wt list",
			short: "list worktrees",
			help:  "List existing worktrees.",
			examples: []string{
				"wt list",
			},
			run:   runList,
		},
		{
			name:  "prune",
			usage: "wt prune",
			short: "prune stale worktrees",
			help:  "Remove stale worktree entries.",
			examples: []string{
				"wt prune",
			},
			run:   runPrune,
		},
		{
			name:  "version",
			usage: "wt version",
			short: "show version",
			help:  "Print version number.",
			examples: []string{
				"wt version",
				"wt -v",
			},
			run:   func(_ []string, out io.Writer) int { return runVersion(out) },
		},
		{
			name:  "help",
			usage: "wt help [command]",
			short: "show help",
			help:  "Show help for commands.",
			examples: []string{
				"wt help",
				"wt help add",
			},
			run:   runHelp,
		},
	}

	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].name < cmds[j].name
	})
	return cmds
}

func aliasCommands() map[string]string {
	return map[string]string{
		"checkout": "co",
		"remove":   "rm",
	}
}

func isHelp(arg string) bool {
	switch arg {
	case "help", "-h", "--help":
		return true
	default:
		return false
	}
}

func isVersion(arg string) bool {
	switch arg {
	case "-v", "--version":
		return true
	default:
		return false
	}
}

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func printHelp(out io.Writer) int {
	fmt.Fprintln(out, "wt - worktree helper")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Usage:")
	fmt.Fprintln(out, "  wt <command> [args]")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Commands:")
	cmds := commands()
	displayNames := make([]string, len(cmds))
	maxLen := 0
	for i, cmd := range cmds {
		displayNames[i] = commandDisplayName(cmd.name)
		if len(displayNames[i]) > maxLen {
			maxLen = len(displayNames[i])
		}
	}
	for i, cmd := range cmds {
		fmt.Fprintf(out, "  %-*s %s\n", maxLen, displayNames[i], cmd.short)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Shortcuts:")
	fmt.Fprintln(out, "  wt <branch>  same as wt cd <branch>")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Notes:")
	fmt.Fprintln(out, "  cd/co move only when shell hook is enabled")
	fmt.Fprintln(out, "  prompt integration is zsh-only")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Examples:")
	fmt.Fprintln(out, "  wt init")
	fmt.Fprintln(out, "  wt add feat/one")
	fmt.Fprintln(out, "  wt co feat/one")
	fmt.Fprintln(out, "  wt cd -")
	return 0
}

func commandByName(name string) (command, bool) {
	if alias, ok := aliasCommands()[name]; ok {
		name = alias
	}
	for _, cmd := range commands() {
		if cmd.name == name {
			return cmd, true
		}
	}
	return command{}, false
}

func commandUsage(name string) string {
	cmd, ok := commandByName(name)
	if !ok {
		return "Usage: wt help\n"
	}
	return fmt.Sprintf("Usage: %s\n", cmd.usage)
}

func printCommandUsage(out io.Writer, name string) int {
	fmt.Fprint(out, commandUsage(name))
	return 2
}

func printCommandHelp(out io.Writer, name string) int {
	cmd, ok := commandByName(name)
	if !ok {
		fmt.Fprint(out, commandUsage(name))
		return 0
	}
	fmt.Fprint(out, commandUsage(name))
	if aliasLine := commandAliasLine(cmd); aliasLine != "" {
		fmt.Fprintln(out, aliasLine)
	}
	if cmd.help != "" {
		fmt.Fprintln(out, "")
		writeIndentedLines(out, cmd.help)
	}
	if len(cmd.examples) > 0 {
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Examples:")
		for _, example := range cmd.examples {
			fmt.Fprintf(out, "  %s\n", example)
		}
	}
	return 0
}

func commandDisplayName(name string) string {
	aliases := aliasesForCommand(name)
	if len(aliases) == 0 {
		return name
	}
	parts := make([]string, 0, 1+len(aliases))
	parts = append(parts, name)
	parts = append(parts, aliases...)
	return strings.Join(parts, ", ")
}

func aliasesForCommand(name string) []string {
	var aliases []string
	for alias, cmd := range aliasCommands() {
		if cmd == name {
			aliases = append(aliases, alias)
		}
	}
	sort.Strings(aliases)
	return aliases
}

func commandAliasLine(cmd command) string {
	aliases := aliasesForCommand(cmd.name)
	if len(aliases) == 0 {
		return ""
	}
	usages := make([]string, 0, len(aliases))
	for _, alias := range aliases {
		usages = append(usages, aliasUsage(cmd.usage, cmd.name, alias))
	}
	if len(usages) == 1 {
		return fmt.Sprintf("Alias: %s", usages[0])
	}
	return fmt.Sprintf("Aliases: %s", strings.Join(usages, ", "))
}

func aliasUsage(usage string, name string, alias string) string {
	return strings.ReplaceAll(usage, "wt "+name, "wt "+alias)
}

func writeIndentedLines(out io.Writer, text string) {
	lines := strings.Split(text, "\n")
	for _, line := range lines {
		fmt.Fprintf(out, "  %s\n", line)
	}
}
