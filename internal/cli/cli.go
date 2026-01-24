package cli

import (
	"fmt"
	"io"
	"sort"
)

type command struct {
	name  string
	usage string
	run   func(args []string, out io.Writer) int
}

func Run(args []string, out io.Writer, errOut io.Writer) int {
	if len(args) == 0 || isHelp(args[0]) {
		return printHelp(out)
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

	fmt.Fprintf(errOut, "wt: unknown command: %s\n", args[0])
	fmt.Fprint(errOut, "Run 'wt help' for usage.\n")
	return 2
}

func commands() []command {
	cmds := []command{
		{
			name:  "init",
			usage: "wt init [-g]",
			run:   runInit,
		},
		{
			name:  "hook",
			usage: "wt hook <shell> [--prompt]",
			run:   runHook,
		},
		{
			name:  "add",
			usage: "wt add <branch>",
			run:   runAdd,
		},
		{
			name:  "co",
			usage: "wt co [branch]",
			run:   runCo,
		},
		{
			name:  "cd",
			usage: "wt cd [branch|-]",
			run:   runCd,
		},
		{
			name:  "rm",
			usage: "wt rm <branch> | wt rm --all [-f]",
			run:   runRm,
		},
		{
			name:  "list",
			usage: "wt list",
			run:   runList,
		},
		{
			name:  "prune",
			usage: "wt prune",
			run:   runPrune,
		},
		{
			name:  "help",
			usage: "wt help",
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

func hasHelpFlag(args []string) bool {
	for _, arg := range args {
		if arg == "-h" || arg == "--help" {
			return true
		}
	}
	return false
}

func printHelp(out io.Writer) int {
	fmt.Fprintln(out, "wt - worktree helper (stub)")
	fmt.Fprintln(out, "")
	fmt.Fprintln(out, "Usage:")
	for _, cmd := range commands() {
		fmt.Fprintf(out, "  %s\n", cmd.usage)
	}
	return 0
}

func commandUsage(name string) string {
	for _, cmd := range commands() {
		if cmd.name == name {
			return fmt.Sprintf("Usage: %s\n", cmd.usage)
		}
	}
	return "Usage: wt help\n"
}

func printCommandUsage(out io.Writer, name string) int {
	fmt.Fprint(out, commandUsage(name))
	return 2
}

func printCommandHelp(out io.Writer, name string) int {
	fmt.Fprint(out, commandUsage(name))
	return 0
}
