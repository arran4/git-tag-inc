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

func LoadConfig(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	var c Config
	if err := json.Unmarshal(b, &c); err != nil {
		return
	}
	if len(c.Envs) > 0 {
		parseTagReLock.Lock()
		defer parseTagReLock.Unlock()
		ConfiguredEnvs = c.Envs
		ConfiguredEnvsMap = make(map[string]int)
		for i, env := range ConfiguredEnvs {
			ConfiguredEnvsMap[strings.ToLower(env)] = i
		}
		parseTagRe = nil // invalidate cache
	}
}
