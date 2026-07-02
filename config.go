// Copyright (c) 2025, Arran Ubels
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package gittaginc

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var ConfiguredEnvs = []string{"test", "uat"}
var ConfiguredEnvsSemver = []string{"alpha", "beta", "rc", "next"}
var ConfiguredEnvsMap = map[string]int{"test": 0, "uat": 1}
var ConfiguredEnvsSemverMap = map[string]int{"alpha": 0, "beta": 1, "rc": 2, "next": 3}

func FindConfig(filename string) (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		path := filepath.Join(dir, filename)
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", os.ErrNotExist
}

func LoadConfig(filename string) error {
	path, err := FindConfig(filename)
	if err != nil {
		return err
	}
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	var envs []string
	scanner := bufio.NewScanner(file)
	// regex to match Envs: test & uat or Envs(test, uat)
	reColon := regexp.MustCompile(`(?i)^\s*envs:\s*(.*)$`)
	reParen := regexp.MustCompile(`(?i)^\s*envs\((.*)\)\s*$`)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		var content string
		if m := reColon.FindStringSubmatch(line); len(m) > 0 {
			content = m[1]
		} else if m := reParen.FindStringSubmatch(line); len(m) > 0 {
			content = m[1]
		}

		if content != "" {
			parts := regexp.MustCompile(`[\s,&|]+`).Split(content, -1)
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					envs = append(envs, p)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	if len(envs) > 0 {
		parseTagReLock.Lock()
		defer parseTagReLock.Unlock()
		ConfiguredEnvs = make([]string, len(envs))
		ConfiguredEnvsMap = make(map[string]int)
		for i, env := range envs {
			lowerEnv := strings.ToLower(env)
			ConfiguredEnvs[i] = lowerEnv
			ConfiguredEnvsMap[lowerEnv] = i
		}
		parseTagRe = nil // invalidate cache
	}
	return nil
}
