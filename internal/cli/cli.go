package cli

import (
	"fmt"
	"io"
	"sort"
	"strings"
)

type command struct {
	name  string
	usage string
	short string
	help  string
	run   func(args []string, out io.Writer) int
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
	for _, cmd := range cmds {
		if cmd.name == args[0] {
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
			run:   runInit,
		},
		{
			name:  "hook",
			usage: "wt hook <shell> [--prompt]",
			short: "print shell integration script",
			help:  "Supported shells: zsh, bash, fish. Prompt is zsh-only.",
			run:   runHook,
		},
		{
			name:  "add",
			usage: "wt add <branch>",
			short: "create a worktree for a branch",
			help:  "Create a worktree (and branch if needed).",
			run:   runAdd,
		},
		{
			name:  "co",
			usage: "wt co <branch>",
			short: "create a worktree and move",
			help:  "Create the worktree if needed, then move to it.",
			run:   runCo,
		},
		{
			name:  "cd",
			usage: "wt cd [branch|-]",
			short: "move to a worktree",
			help:  "No args: main worktree. '-' goes back. Shorthand: wt <branch>.",
			run:   runCd,
		},
		{
			name:  "rm",
			usage: "wt rm <branch> | wt rm --all [-f]",
			short: "remove worktree and branch",
			help:  "Remove worktrees and branches. --all requires main worktree.",
			run:   runRm,
		},
		{
			name:  "list",
			usage: "wt list",
			short: "list worktrees",
			help:  "List existing worktrees.",
			run:   runList,
		},
		{
			name:  "prune",
			usage: "wt prune",
			short: "prune stale worktrees",
			help:  "Remove stale worktree entries.",
			run:   runPrune,
		},
		{
			name:  "version",
			usage: "wt version",
			short: "show version",
			help:  "Print version number.",
			run:   func(_ []string, out io.Writer) int { return runVersion(out) },
		},
		{
			name:  "help",
			usage: "wt help [command]",
			short: "show help",
			help:  "Show help for commands.",
			run:   runHelp,
		},
	}

	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].name < cmds[j].name
	})
	return cmds
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
	for _, cmd := range commands() {
		fmt.Fprintf(out, "  %-8s %s\n", cmd.name, cmd.short)
	}
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Shortcuts:")
	fmt.Fprintln(out, "  wt <branch>  same as wt cd <branch>")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Notes:")
	fmt.Fprintln(out, "  cd/co move only when shell hook is enabled")
	fmt.Fprintln(out, "  prompt integration is zsh-only")
	return 0
}

func commandByName(name string) (command, bool) {
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
	if cmd.help != "" {
		fmt.Fprintln(out, cmd.help)
	}
	return 0
}
