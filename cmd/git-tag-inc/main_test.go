package main

import (
	"bytes"
	"flag"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestUsage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "run once"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			oldOutput := flag.CommandLine.Output()
			flag.CommandLine.SetOutput(&buf)
			defer flag.CommandLine.SetOutput(oldOutput)

			Usage()

			output := buf.String()
			expectedPhrases := []string{
				"Usage of",
				"Flags:",
				"Combinations work:",
				"Preventing backwards moves:",
				"--mode arraneous switches to the legacy naming",
			}
			for _, phrase := range expectedPhrases {
				if !strings.Contains(output, phrase) {
					t.Errorf("Expected usage output to contain %q, but it didn't.", phrase)
				}
			}
		})
	}
}

func TestMain_NoGitRepo(t *testing.T) {
	// Build the binary
	tempDir, err := os.MkdirTemp("", "git-tag-inc-build")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	exeName := "git-tag-inc"
	if runtime.GOOS == "windows" {
		exeName += ".exe"
	}
	exePath := filepath.Join(tempDir, exeName)
	cmdBuild := exec.Command("go", "build", "-o", exePath, ".")
	if out, err := cmdBuild.CombinedOutput(); err != nil {
		t.Fatalf("Failed to build git-tag-inc: %v\nOutput: %s", err, out)
	}

	// Create a directory that is NOT a git repo
	nonGitDir, err := os.MkdirTemp("", "non-git-repo")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(nonGitDir)

	// Run the tool in the non-git dir
	cmd := exec.Command(exePath)
	cmd.Dir = nonGitDir
	output, err := cmd.CombinedOutput()

	// Expect exit code 1
	if err == nil {
		t.Errorf("Expected error (exit code 1), but got nil")
	} else {
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 1 {
				t.Errorf("Expected exit code 1, got %d", exitError.ExitCode())
			}
		} else {
			t.Errorf("Expected exec.ExitError, got %T: %v", err, err)
		}
	}

	// Expect proper error message
	outStr := string(output)
	expected := "Error: repository does not exist. Are you in a git repository?"
	if !strings.Contains(outStr, expected) {
		t.Errorf("Expected output to contain %q, got: %q", expected, outStr)
	}
}
