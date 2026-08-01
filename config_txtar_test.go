package gittaginc

import (
	"encoding/json"
	"golang.org/x/tools/txtar"
	"os"
	"path/filepath"
	"testing"
)

func TestParseConfig_Txtar(t *testing.T) {
	files, err := os.ReadDir("testdata")
	if err != nil {
		t.Fatalf("Failed to read testdata directory: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) == ".txtar" && file.Name() != "config_tests.txtar" {
			t.Run(file.Name(), func(t *testing.T) {
				ar, err := txtar.ParseFile(filepath.Join("testdata", file.Name()))
				if err != nil {
					t.Fatalf("Failed to parse txtar file %s: %v", file.Name(), err)
				}

				var confData, jsonData []byte
				for _, f := range ar.Files {
					if f.Name == "input.conf" {
						confData = f.Data
					} else if f.Name == "expected.json" {
						jsonData = f.Data
					}
				}

				if len(confData) == 0 || len(jsonData) == 0 {
					t.Fatalf("Missing either input.conf or expected.json in %s", file.Name())
				}

				cfg := Config{}
				err = parseConfig(confData, &cfg)
				if err != nil {
					t.Fatalf("parseConfig failed: %v", err)
				}

				var expectedConfig Config
				err = json.Unmarshal(jsonData, &expectedConfig)
				if err != nil {
					t.Fatalf("failed to parse expected JSON: %v", err)
				}

				actualJSON, _ := json.MarshalIndent(cfg, "", "  ")
				expectedJSON, _ := json.MarshalIndent(expectedConfig, "", "  ")

				if string(actualJSON) != string(expectedJSON) {
					t.Errorf("Config mismatch.\nExpected:\n%s\nGot:\n%s", string(expectedJSON), string(actualJSON))
				}
			})
		}
	}
}
