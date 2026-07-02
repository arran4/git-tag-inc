// Copyright (c) 2025, Arran Ubels
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package gittaginc

import (
	"encoding/json"
	"os"
	"strings"
)

var ConfiguredEnvs = []string{"test", "uat"}
var ConfiguredEnvsMap = map[string]int{"test": 0, "uat": 1}

type Config struct {
	Envs []string `json:"envs"`
}

func LoadConfig(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return err
	}
	if len(c.Envs) > 0 {
		parseTagReLock.Lock()
		defer parseTagReLock.Unlock()
		ConfiguredEnvs = make([]string, len(c.Envs))
		ConfiguredEnvsMap = make(map[string]int)
		for i, env := range c.Envs {
			lowerEnv := strings.ToLower(env)
			ConfiguredEnvs[i] = lowerEnv
			ConfiguredEnvsMap[lowerEnv] = i
		}
		parseTagRe = nil // invalidate cache
	}
	return nil
}
