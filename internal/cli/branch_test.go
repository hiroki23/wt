package cli

import (
	"os/exec"
	"testing"
)

func TestBranchStatusLocalBranch(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not found in PATH")
	}

	root := t.TempDir()
	if err := initRepo(root, "main"); err != nil {
		t.Fatalf("init repo failed: %v", err)
	}
	if err := runCmd(root, "git", "branch", "feature/one"); err != nil {
		t.Fatalf("git branch failed: %v", err)
	}

	status, err := branchStatus(root, "feature/one")
	if err != nil {
		t.Fatalf("branchStatus failed: %v", err)
	}
	if !status.local {
		t.Fatalf("expected local branch to be detected")
	}
	if len(status.remoteMatches) != 0 {
		t.Fatalf("expected no remote matches, got %v", status.remoteMatches)
	}
}
