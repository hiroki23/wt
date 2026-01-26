package cli

import "io"

func runHelp(args []string, out io.Writer) int {
	if len(args) == 1 {
		return printCommandHelp(out, args[0])
	}
	if len(args) > 1 {
		return printCommandUsage(out, "help")
	}
	return printHelp(out)
}
