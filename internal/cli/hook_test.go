package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
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

func TestHookZshCdHandlesMultiLineOutput(t *testing.T) {
	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh not found in PATH")
	}

	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}

	hookPath := filepath.Join(dir, "hook.zsh")
	if err := os.WriteFile(hookPath, []byte(hookZshScript(false)), 0o644); err != nil {
		t.Fatalf("write hook failed: %v", err)
	}

	wtPath := filepath.Join(dir, "wt")
	wtScript := fmt.Sprintf(`#!/bin/sh
if [ "$1" = "co" ]; then
  printf '%%s\n' "wt co: created worktree %s (branch created)"
  printf '%%s\n' "%s"
  exit 0
fi
exit 1
`, target, target)
	if err := os.WriteFile(wtPath, []byte(wtScript), 0o755); err != nil {
		t.Fatalf("write wt failed: %v", err)
	}

	script := fmt.Sprintf("PATH=%s:$PATH; source %s; wt co feature/one; pwd", shellQuote(dir), shellQuote(hookPath))
	cmd := exec.Command("zsh", "-c", script)
	output, err := cmd.Output()
	if err != nil {
		t.Fatalf("zsh failed: %v: %s", err, string(output))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) < 2 {
		t.Fatalf("expected output lines, got %q", string(output))
	}
	wantMsg := "wt co: created worktree " + target + " (branch created)"
	if lines[0] != wantMsg {
		t.Fatalf("expected message %q, got %q", wantMsg, lines[0])
	}
	if lines[len(lines)-1] != target {
		t.Fatalf("expected pwd %q, got %q", target, lines[len(lines)-1])
	}
}

func shellQuote(value string) string {
	return strconv.Quote(value)
}
