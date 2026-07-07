package gittaginc

import "testing"
import "os"
import "path/filepath"

func TestCustomEnvConfig(t *testing.T) {
	// Override configuration
	parseTagReLock.Lock()
	originalEnvs := ConfiguredEnvs
	originalMap := ConfiguredEnvsMap
	parseTagReLock.Unlock()

	tempDir := t.TempDir()
	confPath := filepath.Join(tempDir, "test_config.conf")
	err := os.WriteFile(confPath, []byte("Envs: staging, prod"), 0644)
	if err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	err = LoadConfig(confPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	defer func() {
		parseTagReLock.Lock()
		ConfiguredEnvs = originalEnvs
		ConfiguredEnvsMap = originalMap
		parseTagRe = nil
		parseTagReLock.Unlock()
	}()

	// Parse test
	tag := ParseTag("v1.0.0-staging1")
	if tag == nil || tag.EnvName != "staging" || *tag.Env != 1 {
		t.Fatalf("failed to parse custom env: %+v", tag)
	}

	// Format test
	if got := tag.String(); got != "v1.0.0-staging1" {
		t.Fatalf("expected v1.0.0-staging1, got %s", got)
	}

	// CommandsToFlags test
	flags := CommandsToFlags([]string{"prod2"}, "default")
	if flags.Env != "prod" || *flags.EnvValue != 2 {
		t.Fatalf("failed to parse custom env flag: %+v", flags)
	}

	// Increment test
	err = tag.Increment(flags, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := tag.String(); got != "v1.0.0-prod02" {
		t.Fatalf("expected v1.0.0-prod02, got %s", got)
	}
}

func TestFindConfig(t *testing.T) {
	tempDir := t.TempDir()
	repoDir := filepath.Join(tempDir, "repo")
	srcDir := filepath.Join(repoDir, "src")
	os.MkdirAll(srcDir, 0755)

	confPath := filepath.Join(repoDir, ".git-tag-inc.conf")
	err := os.WriteFile(confPath, []byte("Envs: qa, int"), 0644)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	path, err := FindConfig(srcDir, ".git-tag-inc.conf")
	if err != nil {
		t.Fatalf("expected to find config, got %v", err)
	}
	if path != confPath {
		t.Fatalf("expected path %s, got %s", confPath, path)
	}
}
