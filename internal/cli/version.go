package cli

import (
	"fmt"
	"io"
)

func runVersion(out io.Writer) int {
	fmt.Fprintln(out, Version)
	return 0
}
