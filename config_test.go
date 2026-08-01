package gittaginc

import (
	"reflect"
	"testing"
)

func TestParseConfig(t *testing.T) {
	data := []byte(`
# This is a comment
Envs = staging, prod, test
Stages=alpha,beta,rc
`)
	cfg := Config{}
	err := parseConfig(data, &cfg)
	if err != nil {
		t.Fatalf("parseConfig failed: %v", err)
	}

	expectedEnvs := []string{"staging", "prod", "test"}
	if !reflect.DeepEqual(cfg.Envs, expectedEnvs) {
		t.Errorf("expected envs %v, got %v", expectedEnvs, cfg.Envs)
	}

	expectedStages := []string{"alpha", "beta", "rc"}
	if !reflect.DeepEqual(cfg.Stages, expectedStages) {
		t.Errorf("expected stages %v, got %v", expectedStages, cfg.Stages)
	}
}
