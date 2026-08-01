// Copyright (c) 2025, Arran Ubels
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package gittaginc

import (
	"strconv"
	"strings"
)


type CmdFlags struct {
	Major        bool
	MajorValue   *int
	Minor        bool
	MinorValue   *int
	Patch        bool
	PatchValue   *int
	Release      bool
	ReleaseValue *int
	Stage        string
	StageValue   *int
	StageDigits  int
	Env          string
	EnvValue     *int
	EnvDigits    int
	Valid        bool
	Mode         string
}

func CommandsToFlags(args []string, mode string) CmdFlags {
	c := CmdFlags{Valid: true, Mode: mode}
	for _, f := range args {
		lower := strings.ToLower(f)

		// Extract name and digits manually instead of regexp
		letterEndIdx := 0
		for i, char := range lower {
			if char >= 'a' && char <= 'z' {
				letterEndIdx = i + 1
			} else {
				break
			}
		}
		if letterEndIdx == 0 {
			c.Valid = false
			return c
		}

		name := lower[:letterEndIdx]
		digitsStr := lower[letterEndIdx:]
		var value *int
		if digitsStr != "" {
			v, err := strconv.Atoi(digitsStr)
			if err != nil {
				c.Valid = false
				return c
			}
			value = &v
		}

		if name == "major" {
			c.Major = true
			if value != nil {
				c.MajorValue = value
			}
		} else if name == "minor" {
			c.Minor = true
			if value != nil {
				c.MinorValue = value
			}
		} else if name == "patch" {
			if mode == "arraneous" {
				c.Valid = false
				return c
			}
			c.Patch = true
			if value != nil {
				c.PatchValue = value
			}
		} else if name == "release" {
			if mode == "arraneous" {
				c.Patch = true
				if value != nil {
					c.PatchValue = value
				}
			} else {
				c.Release = true
				if value != nil {
					c.ReleaseValue = value
				}
			}
		} else if _, ok := ConfiguredStagesMap[name]; ok {
			if c.Stage != "" {
				c.Valid = false
				return c
			}
			c.Stage = name
			if value != nil {
				c.StageValue = value
				c.StageDigits = len(digitsStr)
			}
		} else if _, ok := ConfiguredEnvsMap[name]; ok {
			if c.Env != "" {
				c.Valid = false
				return c
			}
			c.Env = name
			if value != nil {
				c.EnvValue = value
				c.EnvDigits = len(digitsStr)
			}
		} else {
			c.Valid = false
			return c
		}
	}
	return c
}
