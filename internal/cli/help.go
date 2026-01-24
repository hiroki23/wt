package cli

import "io"

func runHelp(args []string, out io.Writer) int {
	return printHelp(out)
}
