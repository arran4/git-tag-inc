package gittaginc

import "testing"
import "testing/fstest"

func TestCustomEnvConfig(t *testing.T) {
	// Override configuration
	parseTagReLock.Lock()
	originalEnvs := ConfiguredEnvs
	originalMap := ConfiguredEnvsMap
	parseTagReLock.Unlock()

	mockFS := fstest.MapFS{
		"project/testdata/test_config.conf": &fstest.MapFile{
			Data: []byte("Envs: staging, prod"),
		},
	}
	err := LoadConfigFS(mockFS, "project/testdata/subfolder", "test_config.conf")
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
	mockFS := fstest.MapFS{
		"repo/.git-tag-inc.conf": &fstest.MapFile{
			Data: []byte("Envs: qa, int"),
		},
		"repo/src/main.go": &fstest.MapFile{
			Data: []byte("package main"),
		},
	}
	path, err := FindConfig(mockFS, "repo/src", ".git-tag-inc.conf")
	if err != nil {
		t.Fatalf("expected to find config, got %v", err)
	}
	if path != "repo/.git-tag-inc.conf" {
		t.Fatalf("expected path repo/.git-tag-inc.conf, got %s", path)
	}
}
