// Copyright (c) 2025, Arran Ubels
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package gittaginc

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Envs   []string
	Stages []string
}

func parseConfig(data []byte, cfg *Config) error {
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])
			items := strings.Split(val, ",")
			var parsedItems []string
			for _, item := range items {
				item = strings.TrimSpace(item)
				if item != "" {
					parsedItems = append(parsedItems, item)
				}
			}

			if strings.EqualFold(key, "Envs") {
				cfg.Envs = parsedItems
			} else if strings.EqualFold(key, "Stages") {
				cfg.Stages = parsedItems
			}
		}
	}
	return nil
}

var DefaultConfig = Config{
	Envs:   []string{"test", "uat"},
	Stages: []string{"alpha", "beta", "rc", "next"},
}

var ConfiguredEnvsMap map[string]int
var ConfiguredStagesMap map[string]int

func init() {
	LoadConfig()
}

func LoadConfigEx(configUrl string, requireConfig bool) error {
	cfg := DefaultConfig

	found := false
	if configUrl != "" {
		if strings.HasPrefix(configUrl, "http://") || strings.HasPrefix(configUrl, "https://") {
			resp, err := http.Get(configUrl)
			if err != nil {
				if requireConfig {
					return fmt.Errorf("failed to fetch config from %s: %v", configUrl, err)
				}
			} else {
				defer resp.Body.Close()
				data, err := io.ReadAll(resp.Body)
				if err == nil {
					var parsedConfig Config
					if err := parseConfig(data, &parsedConfig); err == nil {
						if len(parsedConfig.Envs) > 0 {
							cfg.Envs = parsedConfig.Envs
						}
						if len(parsedConfig.Stages) > 0 {
							cfg.Stages = parsedConfig.Stages
						}
						found = true
					} else if requireConfig {
						return fmt.Errorf("failed to parse config from %s: %v", configUrl, err)
					}
				} else if requireConfig {
					return fmt.Errorf("failed to read config body from %s: %v", configUrl, err)
				}
			}
		} else {
			data, err := os.ReadFile(configUrl)
			if err != nil {
				if requireConfig {
					return fmt.Errorf("failed to read local config from %s: %v", configUrl, err)
				}
			} else {
				var parsedConfig Config
				if err := parseConfig(data, &parsedConfig); err == nil {
					if len(parsedConfig.Envs) > 0 {
						cfg.Envs = parsedConfig.Envs
					}
					if len(parsedConfig.Stages) > 0 {
						cfg.Stages = parsedConfig.Stages
					}
					found = true
				} else if requireConfig {
					return fmt.Errorf("failed to parse local config from %s: %v", configUrl, err)
				}
			}
		}
	}

	if !found {
		dir, err := os.Getwd()
		if err == nil {
			for {
				for _, filename := range []string{".git-tag-inc.conf", ".gittaginc.conf"} {
					path := filepath.Join(dir, filename)
					data, err := os.ReadFile(path)
					if err == nil {
						var parsedConfig Config
						if err := parseConfig(data, &parsedConfig); err == nil {
							if len(parsedConfig.Envs) > 0 {
								cfg.Envs = parsedConfig.Envs
							}
							if len(parsedConfig.Stages) > 0 {
								cfg.Stages = parsedConfig.Stages
							}
						}
						found = true
						break
					}
				}

				if found {
					break
				}

				// Stop traversing if we hit the repository root
				if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
					break
				}

				parent := filepath.Dir(dir)
				if parent == dir {
					// Reached root of the file system
					break
				}
				dir = parent
			}
		}
	}

	if requireConfig && !found {
		return fmt.Errorf("required config file not found")
	}

	ConfiguredEnvsMap = make(map[string]int)
	for i, env := range cfg.Envs {
		ConfiguredEnvsMap[env] = i
	}

	ConfiguredStagesMap = make(map[string]int)
	for i, stage := range cfg.Stages {
		ConfiguredStagesMap[stage] = i
	}

	return nil
}

func LoadConfig() {
	_ = LoadConfigEx("", false)
}
