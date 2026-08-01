package gittaginc

import (
	_ "embed"
	"encoding/json"
	"golang.org/x/tools/txtar"
	"strings"
	"testing"
)

//go:embed testdata/config_tests.txtar
var configTestsTxtar []byte

func TestParseConfig_Txtar(t *testing.T) {
	ar := txtar.Parse(configTestsTxtar)

	// Group files by prefix (e.g., test1.conf and test1.json)
	tests := make(map[string]struct {
		conf []byte
		json []byte
	})

	for _, f := range ar.Files {
		parts := strings.Split(f.Name, ".")
		if len(parts) < 2 {
			continue
		}
		name := parts[0]
		ext := parts[1]

		entry := tests[name]
		if ext == "conf" {
			entry.conf = f.Data
		} else if ext == "json" {
			entry.json = f.Data
		}
		tests[name] = entry
	}

	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			if len(data.conf) == 0 || len(data.json) == 0 {
				t.Fatalf("Missing either .conf or .json for test %s", name)
			}

			cfg := Config{}
			err := parseConfig(data.conf, &cfg)
			if err != nil {
				t.Fatalf("parseConfig failed: %v", err)
			}

			var expectedConfig Config
			err = json.Unmarshal(data.json, &expectedConfig)
			if err != nil {
				t.Fatalf("failed to parse expected JSON: %v", err)
			}

			// Normalize empty slices to nil for deep equal if needed, but our parseConfig appends to nil.
			// Let's compare JSON string output instead.
			actualJSON, _ := json.MarshalIndent(cfg, "", "  ")
			expectedJSON, _ := json.MarshalIndent(expectedConfig, "", "  ")

			if string(actualJSON) != string(expectedJSON) {
				t.Errorf("Config mismatch.\nExpected:\n%s\nGot:\n%s", string(expectedJSON), string(actualJSON))
			}
		})
	}
}
