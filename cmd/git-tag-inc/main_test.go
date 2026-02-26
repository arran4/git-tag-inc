// Copyright (c) 2025, Arran Ubels
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
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
			Usage()
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
