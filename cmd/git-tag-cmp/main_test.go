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
	defer func() { _ = os.Remove(binaryName) }()

	tests := []struct {
		name     string
		args     []string
		stdin    string
		expected int
		out      string
	}{
		{"less_than_true", []string{"v1.0.0<v2.0.0"}, "", 0, "true\n"},
		{"less_than_false", []string{"v2.0.0<v1.0.0"}, "", 1, "false\n"},
		{"less_than_equal_true1", []string{"v1.0.0<=v2.0.0"}, "", 0, "true\n"},
		{"less_than_equal_true2", []string{"v1.0.0<=v1.0.0"}, "", 0, "true\n"},
		{"less_than_equal_false", []string{"v2.0.0<=v1.0.0"}, "", 1, "false\n"},
		{"greater_than_true", []string{"v2.0.0>v1.0.0"}, "", 0, "true\n"},
		{"greater_than_false", []string{"v1.0.0>v2.0.0"}, "", 1, "false\n"},
		{"greater_than_equal_true1", []string{"v2.0.0>=v1.0.0"}, "", 0, "true\n"},
		{"greater_than_equal_true2", []string{"v1.0.0>=v1.0.0"}, "", 0, "true\n"},
		{"greater_than_equal_false", []string{"v1.0.0>=v2.0.0"}, "", 1, "false\n"},
		{"equal_true", []string{"v1.0.0==v1.0.0"}, "", 0, "true\n"},
		{"equal_false", []string{"v1.0.0==v2.0.0"}, "", 1, "false\n"},
		{"not_equal_true", []string{"v1.0.0!=v2.0.0"}, "", 0, "true\n"},
		{"not_equal_false", []string{"v1.0.0!=v1.0.0"}, "", 1, "false\n"},

		{"bash_lt_true", []string{"v1.0.0", "-lt", "v2.0.0"}, "", 0, "true\n"},
		{"bash_le_true", []string{"v1.0.0", "-le", "v2.0.0"}, "", 0, "true\n"},
		{"bash_gt_true", []string{"v2.0.0", "-gt", "v1.0.0"}, "", 0, "true\n"},
		{"bash_ge_true", []string{"v2.0.0", "-ge", "v1.0.0"}, "", 0, "true\n"},
		{"bash_eq_true", []string{"v1.0.0", "-eq", "v1.0.0"}, "", 0, "true\n"},
		{"bash_ne_true", []string{"v1.0.0", "-ne", "v2.0.0"}, "", 0, "true\n"},

		{"word_lt_true", []string{"v1.0.0", "lt", "v2.0.0"}, "", 0, "true\n"},
		{"word_le_true", []string{"v1.0.0", "le", "v2.0.0"}, "", 0, "true\n"},
		{"word_gt_true", []string{"v2.0.0", "gt", "v1.0.0"}, "", 0, "true\n"},
		{"word_ge_true", []string{"v2.0.0", "ge", "v1.0.0"}, "", 0, "true\n"},
		{"word_eq_true", []string{"v1.0.0", "eq", "v1.0.0"}, "", 0, "true\n"},
		{"word_ne_true", []string{"v1.0.0", "ne", "v2.0.0"}, "", 0, "true\n"},

		{"full_word_lessthan_true", []string{"v1.0.0", "lessthan", "v2.0.0"}, "", 0, "true\n"},
		{"full_word_less-than_true", []string{"v1.0.0", "less-than", "v2.0.0"}, "", 0, "true\n"},
		{"full_word_lessthanorequal_true", []string{"v1.0.0", "lessthanorequal", "v2.0.0"}, "", 0, "true\n"},
		{"full_word_less-than-or-equal_true", []string{"v1.0.0", "less-than-or-equal", "v2.0.0"}, "", 0, "true\n"},
		{"full_word_greaterthan_true", []string{"v2.0.0", "greaterthan", "v1.0.0"}, "", 0, "true\n"},
		{"full_word_greater-than_true", []string{"v2.0.0", "greater-than", "v1.0.0"}, "", 0, "true\n"},
		{"full_word_greaterthanorequal_true", []string{"v2.0.0", "greaterthanorequal", "v1.0.0"}, "", 0, "true\n"},
		{"full_word_greater-than-or-equal_true", []string{"v2.0.0", "greater-than-or-equal", "v1.0.0"}, "", 0, "true\n"},
		{"full_word_equal_true", []string{"v1.0.0", "equal", "v1.0.0"}, "", 0, "true\n"},
		{"full_word_equals_true", []string{"v1.0.0", "equals", "v1.0.0"}, "", 0, "true\n"},
		{"full_word_notequal_true", []string{"v1.0.0", "notequal", "v2.0.0"}, "", 0, "true\n"},
		{"full_word_not-equal_true", []string{"v1.0.0", "not-equal", "v2.0.0"}, "", 0, "true\n"},

		{"spaced_args", []string{"v1.0.0", "<=", "v2.0.0"}, "", 0, "true\n"},

		{"stdin", []string{}, "v1.0.0 < v2.0.0", 0, "true\n"},
		{"stdin_bash_ops", []string{}, "v1.0.0 -le v2.0.0", 0, "true\n"},
		{"stdin_word_ops", []string{}, "v1.0.0 le v2.0.0", 0, "true\n"},
		{"stdin_full_word_ops", []string{}, "v1.0.0 less-than-or-equal v2.0.0", 0, "true\n"},

		{"invalid_format", []string{"v1.0.0v2.0.0"}, "", 2, "Invalid format. Expected <tag1> <op> <tag2>\n"},
		{"invalid_tag1", []string{"invalid<v1.0.0"}, "", 2, "Invalid tag or path: invalid\n"},
		{"invalid_tag2", []string{"v1.0.0<invalid"}, "", 2, "Invalid tag or path: invalid\n"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cmd := exec.Command("./"+binaryName, tc.args...)
			if tc.stdin != "" {
				cmd.Stdin = strings.NewReader(tc.stdin)
			}

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
