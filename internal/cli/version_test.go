package cli

import (
	"bytes"
	"testing"
)

func TestRunVersionFlag(t *testing.T) {
	original := Version
	Version = "test"
	defer func() {
		Version = original
	}()

	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"-v"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	if got := out.String(); got != "test\n" {
		t.Fatalf("unexpected stdout: %q", got)
	}
}
