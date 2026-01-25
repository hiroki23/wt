package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRunHookZshOutputsScript(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"hook", "zsh"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}

	script := out.String()
	if !strings.Contains(script, "_wt_run") {
		t.Fatalf("expected script to include _wt_run, got %q", script)
	}
	if !strings.Contains(script, "wt()") {
		t.Fatalf("expected script to include wt() function, got %q", script)
	}
}

func TestRunHookZshPromptOutputsScript(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"hook", "zsh", "--prompt"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}

	script := out.String()
	if !strings.Contains(script, "_wt_update_prompt") {
		t.Fatalf("expected script to include _wt_update_prompt, got %q", script)
	}
	if !strings.Contains(script, "_WT_PROMPT_DEFAULT") {
		t.Fatalf("expected script to include _WT_PROMPT_DEFAULT, got %q", script)
	}
}

func TestRunHookBashOutputsScript(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"hook", "bash"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}

	script := out.String()
	if !strings.Contains(script, "_wt_run") {
		t.Fatalf("expected script to include _wt_run, got %q", script)
	}
	if !strings.Contains(script, "wt()") {
		t.Fatalf("expected script to include wt() function, got %q", script)
	}
}

func TestRunHookFishOutputsScript(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"hook", "fish"}, &out, &errOut)

	if code != 0 {
		t.Fatalf("expected exit code 0, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}

	script := out.String()
	if !strings.Contains(script, "function wt") {
		t.Fatalf("expected script to include fish wt function, got %q", script)
	}
}

func TestRunHookBashPromptErrors(t *testing.T) {
	var out bytes.Buffer
	var errOut bytes.Buffer

	code := Run([]string{"hook", "bash", "--prompt"}, &out, &errOut)

	if code == 0 {
		t.Fatalf("expected non-zero exit code, got %d", code)
	}
	if got := errOut.String(); got != "" {
		t.Fatalf("expected no stderr output, got %q", got)
	}
	if got := out.String(); got != "wt hook: prompt is only supported for zsh\n" {
		t.Fatalf("unexpected stdout: %q", got)
	}
}
