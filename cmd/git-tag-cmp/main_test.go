package main

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func TestMain(t *testing.T) {
	binaryName := "git-tag-cmp"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	buildCmd := exec.Command("go", "build", "-o", binaryName)
	if err := buildCmd.Run(); err != nil {
		t.Fatalf("Failed to build binary: %v", err)
	}
	defer os.Remove(binaryName)

	tests := []struct {
		name     string
		arg      string
		expected int
		out      string
	}{
		{"less_than_true", "v1.0.0<v2.0.0", 0, "true\n"},
		{"less_than_false", "v2.0.0<v1.0.0", 1, "false\n"},
		{"less_than_equal_true1", "v1.0.0<=v2.0.0", 0, "true\n"},
		{"less_than_equal_true2", "v1.0.0<=v1.0.0", 0, "true\n"},
		{"less_than_equal_false", "v2.0.0<=v1.0.0", 1, "false\n"},
		{"greater_than_true", "v2.0.0>v1.0.0", 0, "true\n"},
		{"greater_than_false", "v1.0.0>v2.0.0", 1, "false\n"},
		{"greater_than_equal_true1", "v2.0.0>=v1.0.0", 0, "true\n"},
		{"greater_than_equal_true2", "v1.0.0>=v1.0.0", 0, "true\n"},
		{"greater_than_equal_false", "v1.0.0>=v2.0.0", 1, "false\n"},
		{"equal_true", "v1.0.0==v1.0.0", 0, "true\n"},
		{"equal_false", "v1.0.0==v2.0.0", 1, "false\n"},
		{"not_equal_true", "v1.0.0!=v2.0.0", 0, "true\n"},
		{"not_equal_false", "v1.0.0!=v1.0.0", 1, "false\n"},
		{"invalid_format", "v1.0.0v2.0.0", 2, "Invalid format. Expected <tag1><op><tag2>\n"},
		{"invalid_tag1", "invalid<v1.0.0", 2, "Invalid tag: invalid\n"},
		{"invalid_tag2", "v1.0.0<invalid", 2, "Invalid tag: invalid\n"},
		{"spaced_args", "v1.0.0 <= v2.0.0", 0, "true\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("./"+binaryName, tc.arg)
			out, err := cmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					t.Fatalf("Failed to run binary: %v", err)
				}
			}
			if exitCode != tc.expected {
				t.Errorf("Expected exit code %d, got %d. Output: %s", tc.expected, exitCode, out)
			}
			if tc.out != "" {
				// Normalize output
				actualOut := strings.ReplaceAll(string(out), "\r\n", "\n")
				if !strings.HasSuffix(actualOut, tc.out) {
					t.Errorf("Expected output %q, got %q", tc.out, actualOut)
				}
			}
		})
	}
}
