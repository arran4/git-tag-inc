// Copyright (c) 2025, Arran Ubels
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package gittaginc

import (
	"fmt"
	"strconv"
	"strings"
)

func ptr(i int) *int {
	return &i
}

type Tag struct {
	Hash string
	Mode string

	StageName string
	Stage     *int
	StagePad  int

	EnvName string
	Env *int
	Pad  int

	Patch   int
	Release *int
	Major   int
	Minor   int
}

func (t *Tag) Clone() *Tag {
	if t == nil {
		return nil
	}
	clone := &Tag{
		Hash:      t.Hash,
		Mode:      t.Mode,
		StageName: t.StageName,
		StagePad:  t.StagePad,
		Pad:       t.Pad,
		Patch:     t.Patch,
		Major:     t.Major,
		Minor:     t.Minor,
	}
	if t.Stage != nil {
		v := *t.Stage
		clone.Stage = &v
	}
	if t.Env != nil {
		v := *t.Env
		clone.Env = &v
	}
	clone.EnvName = t.EnvName
	if t.Release != nil {
		v := *t.Release
		clone.Release = &v
	}
	return clone
}

func (t *Tag) CopyFrom(other *Tag) {
	if other == nil {
		*t = Tag{}
		return
	}
	clone := other.Clone()
	*t = *clone
}

func stageRank(n string) int {
	if n == "" {
		return len(ConfiguredStagesMap) // Release
	}
	if rank, ok := ConfiguredStagesMap[strings.ToLower(n)]; ok {
		return rank
	}
	return len(ConfiguredStagesMap) + 1 // Other
}

func (t *Tag) LessThan(other *Tag) bool {
	if t.Major != other.Major {
		return t.Major < other.Major
	}
	if t.Minor != other.Minor {
		return t.Minor < other.Minor
	}
	if t.Patch != other.Patch {
		return t.Patch < other.Patch
	}

	if stageRank(t.StageName) != stageRank(other.StageName) {
		return stageRank(t.StageName) < stageRank(other.StageName)
	}
	if t.Stage != nil || other.Stage != nil {
		tv := 0
		ov := 0
		if t.Stage != nil {
			tv = *t.Stage
		}
		if other.Stage != nil {
			ov = *other.Stage
		}
		if tv != ov {
			return tv < ov
		}
	}
	var tv *int = t.Env
	var ov *int = other.Env

	if tv == nil && ov == nil {
		// Both have no env. Check release next.
	} else if tv == nil {
		return false
	} else if ov == nil {
		return true
	} else if *tv != *ov {
		return *tv < *ov
	} else {
		// Both have env and value is same. Compare env priorities.
		tRank, tOk := ConfiguredEnvsMap[t.EnvName]
		oRank, oOk := ConfiguredEnvsMap[other.EnvName]

		if !tOk { tRank = len(ConfiguredEnvsMap) }
		if !oOk { oRank = len(ConfiguredEnvsMap) }

		if tRank != oRank {
			// Higher rank number means lower priority (from 0 to n)
			// Actually wait, let's look at original logic. If UAT vs Test, Test < UAT was returning true?
			// if other.Uat != nil && t.Test != nil { return true } means Test < Uat
			// So lower rank number (test=0, uat=1) means less than!
			return tRank < oRank
		}
	}

	rv := 0
	ovv := 0
	if t.Release != nil {
		rv = *t.Release
	}
	if other.Release != nil {
		ovv = *other.Release
	}
	if rv != ovv {
		return rv < ovv
	}
	return false
}

func (t *Tag) String() string {
	ext := ""
	if t.Mode == ModeLegacy || t.Mode == ModeArraneous {
		if t.Stage != nil {
			ext += fmt.Sprintf("-%s%0*d", t.StageName, t.StagePad, *t.Stage)
		}
		if t.Env != nil {
			ext += fmt.Sprintf("-%s%0*d", t.EnvName, t.Pad, *t.Env)
		}
		if t.Release != nil {
			ext += fmt.Sprintf(".%d", *t.Release)
		}
	} else {
		if t.Stage != nil {
			ext += fmt.Sprintf("-%s.%0*d", t.StageName, t.StagePad, *t.Stage)
		}
		if t.Env != nil {
			if ext == "" {
				ext += fmt.Sprintf("-%s.", t.EnvName)
			} else {
				ext += fmt.Sprintf(".%s.", t.EnvName)
			}
			ext += fmt.Sprintf("%0*d", t.Pad, *t.Env)
		}
		if t.Release != nil {
			if ext == "" {
				ext += fmt.Sprintf("-%d", *t.Release)
			} else {
				ext += fmt.Sprintf(".%d", *t.Release)
			}
		}
	}
	return fmt.Sprintf("v%d.%d.%d%s", t.Major, t.Minor, t.Patch, ext)
}



func ParseTag(tag string) (*Tag, error) {
	if !strings.HasPrefix(tag, "v") {
		return nil, fmt.Errorf("missing 'v' prefix")
	}

	// Remove 'v' prefix
	s := tag[1:]
	t := &Tag{}

	// Parse Major
	dotIdx := strings.Index(s, ".")
	if dotIdx == -1 {
		return nil, fmt.Errorf("missing minor version component")
	}
	majorStr := s[:dotIdx]
	var err error
	t.Major, err = strconv.Atoi(majorStr)
	if err != nil {
		return nil, fmt.Errorf("Version component `major` value `%s` invalid", majorStr)
	}
	s = s[dotIdx+1:]

	// Parse Minor
	dotIdx = strings.Index(s, ".")
	if dotIdx == -1 {
		return nil, fmt.Errorf("missing patch version component")
	}
	minorStr := s[:dotIdx]
	t.Minor, err = strconv.Atoi(minorStr)
	if err != nil {
		return nil, fmt.Errorf("Version component `minor` value `%s` invalid", minorStr)
	}
	s = s[dotIdx+1:]

	// Parse Patch
	patchEndIdx := len(s)
	for i, c := range s {
		if c < '0' || c > '9' {
			patchEndIdx = i
			break
		}
	}
	if patchEndIdx == 0 {
		return nil, fmt.Errorf("Version component `patch` value `%s` invalid", s)
	}
	patchStr := s[:patchEndIdx]
	t.Patch, err = strconv.Atoi(patchStr)
	if err != nil {
		return nil, fmt.Errorf("Version component `patch` value `%s` invalid", patchStr)
	}

	s = s[patchEndIdx:]

	// If nothing remains, it's just vX.Y.Z
	if len(s) == 0 {
		t.Mode = ModeLegacy // Default to legacy, although mode only really matters for extensions
		return t, nil
	}

	if strings.Contains(s, ".") {
		t.Mode = ModeSemver
	} else {
		t.Mode = ModeLegacy
	}

	// Helper function to extract a component and its digits
	// Returns: componentName, padLen, value, remaining string
	extractComponent := func(str string) (string, int, *int, string, error) {
		if len(str) == 0 {
			return "", 0, nil, "", nil
		}

		// MUST start with separator
		if str[0] != '-' && str[0] != '.' {
			return "", 0, nil, str, nil
		}

		sub := str[1:]
		if len(sub) == 0 {
			return "", 0, nil, str, fmt.Errorf("missing component after separator")
		}

		// Find where letters end and digits begin
		letterEndIdx := 0
		for i, c := range sub {
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				letterEndIdx = i + 1
			} else {
				break
			}
		}

		// Check for release component (just digits)
		if letterEndIdx == 0 {
			digitEndIdx := 0
			for i, c := range sub {
				if c >= '0' && c <= '9' {
					digitEndIdx = i + 1
				} else {
					break
				}
			}
			if digitEndIdx > 0 {
				val, _ := strconv.Atoi(sub[:digitEndIdx])
				return "release", 0, ptr(val), sub[digitEndIdx:], nil
			}
			return "", 0, nil, str, fmt.Errorf("invalid release component")
		}

		compName := strings.ToLower(sub[:letterEndIdx])
		sub = sub[letterEndIdx:]

		// Check for separator between word and digits (e.g. . in rc.1 or - in rc-1, or none in rc1)
		// The regex `(?:(?:-|\.?)((?:0*)(\d+)))` means separator is optional (`-`, `.`, or empty).
		if len(sub) > 0 && (sub[0] == '-' || sub[0] == '.') {
			sub = sub[1:]
		}

		digitEndIdx := 0
		for i, c := range sub {
			if c >= '0' && c <= '9' {
				digitEndIdx = i + 1
			} else {
				break
			}
		}

		if digitEndIdx == 0 {
			// MUST have digits! "(\d+)" requires at least 1 digit.
			return "", 0, nil, str, fmt.Errorf("missing digits for component `%s`", compName)
		}

		digitsStr := sub[:digitEndIdx]
		val, _ := strconv.Atoi(digitsStr)
		padLen := len(digitsStr) // t.Pad and t.StagePad are the total length of the match for `((?:0*)(\d+))`

		return compName, padLen, ptr(val), sub[digitEndIdx:], nil
	}

// Parse arbitrary suffix components
	for len(s) > 0 {
		compName, padLen, valPtr, nextS, err := extractComponent(s)
		if err != nil {
			return nil, err
		}

		if _, ok := ConfiguredStagesMap[compName]; ok {
			if t.StageName == "" {
				t.StageName = compName
				t.StagePad = padLen
				if valPtr != nil {
					t.Stage = valPtr
				}
			} else {
				return nil, fmt.Errorf("duplicate stage component `%s`", compName)
			}
		} else if _, ok := ConfiguredEnvsMap[compName]; ok {
			if t.EnvName == "" {
				t.EnvName = compName
				t.Pad = padLen
				if valPtr != nil {
					t.Env = valPtr
				}
			} else {
				return nil, fmt.Errorf("duplicate env component `%s`", compName)
			}
		} else if compName == "release" {
			if valPtr != nil {
				t.Release = valPtr
			}
		} else {
			return nil, fmt.Errorf("unknown component `%s`", compName)
		}

		s = nextS
	}

	return t, nil
}

func (t *Tag) applyIncrement(flags CmdFlags) {
	prevStage := t.Stage
	prevStageName := strings.ToLower(t.StageName)
	prevStagePad := t.StagePad
	prevEnv := t.Env
	prevEnvType := strings.ToLower(t.EnvName)
	prevPad := t.Pad

	if flags.Major {
		target := t.Major + 1
		if flags.MajorValue != nil {
			target = *flags.MajorValue
		}
		t.Major = target
		t.Minor = 0
		t.Patch = 0
		t.Release = nil
		t.Stage = nil
		t.StageName = ""
		t.StagePad = 0
		t.Env = nil
		t.EnvName = ""
		prevStage = nil
		prevStageName = ""
		prevEnv = nil
		prevEnvType = ""
	}
	if flags.Minor {
		target := t.Minor + 1
		if flags.MinorValue != nil {
			target = *flags.MinorValue
		}
		t.Minor = target
		t.Patch = 0
		t.Release = nil
		t.Stage = nil
		t.StageName = ""
		t.StagePad = 0
		t.Env = nil
		t.EnvName = ""
		prevStage = nil
		prevStageName = ""
		prevEnv = nil
		prevEnvType = ""
	}
	if flags.Patch {
		target := t.Patch
		if flags.PatchValue != nil {
			target = *flags.PatchValue
		} else if (t.Env == nil || flags.Env != "") && (t.Stage == nil || flags.Stage != "") {
			target = t.Patch + 1
		}
		t.Patch = target
		t.Stage = nil
		t.StageName = ""
		t.StagePad = 0
		t.Env = nil
		t.EnvName = ""
		t.Release = nil
		prevStage = nil
		prevStageName = ""
		prevEnv = nil
		prevEnvType = ""
	}
	if flags.Stage != "" {
		stageName := strings.ToLower(flags.Stage)
		stagePad := 2
		sameStage := prevStage != nil && prevStageName == stageName
		if sameStage {
			stagePad = prevStagePad
		}
		if flags.StageDigits > 0 {
			requestedPad := flags.StageDigits
			if sameStage {
				if requestedPad > stagePad {
					stagePad = requestedPad
				}
			} else {
				if requestedPad > stagePad {
					stagePad = requestedPad
				} else if requestedPad >= stagePad {
					stagePad = requestedPad
				}
			}
		}
		z := 1
		if flags.StageValue != nil {
			z = *flags.StageValue
		} else if prevStage != nil && prevStageName == stageName {
			z = *prevStage + 1
		} else if !flags.Major && !flags.Minor && !flags.Patch {
			t.Patch += 1
		}
		t.Stage = ptr(z)
		t.StagePad = stagePad
		t.StageName = stageName
		prevEnv = nil
		prevEnvType = ""
		prevPad = 0
		t.Env = nil
		t.EnvName = ""
		t.Release = nil
	}

	if flags.Env != "" {
		envName := strings.ToLower(flags.Env)
		envPad := 2
		sameEnv := prevEnv != nil && prevEnvType == envName
		if sameEnv {
			envPad = prevPad
		}
		if flags.EnvDigits > 0 {
			requestedPad := flags.EnvDigits
			if sameEnv {
				if requestedPad > envPad {
					envPad = requestedPad
				}
			} else {
				if requestedPad > envPad {
					envPad = requestedPad
				} else if requestedPad >= envPad {
					envPad = requestedPad
				}
			}
		}
		z := 1
		if prevEnv != nil {
			if prevEnvType == envName {
				z = *prevEnv + 1
			} else {
				z = *prevEnv
			}
		} else if !flags.Major && !flags.Minor && !flags.Patch && flags.Stage == "" && prevStage == nil {
			t.Patch += 1
		}
		if flags.EnvValue != nil {
			z = *flags.EnvValue
		}
		t.Pad = envPad
		t.EnvName = envName
		t.Env = ptr(z)
		t.Release = nil
	}

	if flags.Release {
		target := 1
		if flags.ReleaseValue != nil {
			target = *flags.ReleaseValue
		} else if t.Release != nil {
			target = *t.Release + 1
		}
		t.Release = ptr(target)
	}
}

func (t *Tag) Increment(flags CmdFlags, allowBackwards bool, skipForwards bool) error {
	original := t.Clone()
	if original == nil {
		return fmt.Errorf("no tag to increment")
	}

	currentFlags := flags
	t.applyIncrement(currentFlags)

	decreases := detectDecreases(original, t, currentFlags)
	if len(decreases) == 0 {
		if allowBackwards {
			return nil
		}
		if original.String() == t.String() {
			newTag := t.String()
			t.CopyFrom(original)
			return fmt.Errorf("resulting tag %s is unchanged from previous", newTag)
		}
		return nil
	}

	if allowBackwards {
		return nil
	}

	if skipForwards && !flags.Major && !flags.Minor && !flags.Patch {
		t.CopyFrom(original)
		autoFlags := flags
		autoFlags.Patch = true
		autoFlags.PatchValue = ptr(original.Patch + 1)
		currentFlags = autoFlags
		t.applyIncrement(currentFlags)
		decreases = detectDecreases(original, t, currentFlags)
		if len(decreases) == 0 {
			return nil
		}
	}

	newTag := t.String()
	originalTag := original.String()
	msg := formatDecreases(decreases)
	t.CopyFrom(original)
	return fmt.Errorf("%s; use --allow-backwards to force (previous %s, requested %s)", msg, originalTag, newTag)
}

type decrease struct {
	component string
	previous  int
	current   int
}

func envInfo(tag *Tag) (string, *int) {
	if tag.Env != nil {
		return tag.EnvName, tag.Env
	}
	return "", nil
}

func detectDecreases(original, current *Tag, flags CmdFlags) []decrease {
	var result []decrease

	checkInt := func(component string, previous, current int, target *int, valid bool) {
		if valid && target != nil && current < previous {
			result = append(result, decrease{component: component, previous: previous, current: current})
		}
	}

	checkPtr := func(component string, previous, current *int, target *int, valid bool) {
		if valid && target != nil && previous != nil && current != nil {
			if *current < *previous {
				result = append(result, decrease{component: component, previous: *previous, current: *current})
			}
		}
	}

	checkInt("major", original.Major, current.Major, flags.MajorValue, true)
	checkInt("minor", original.Minor, current.Minor, flags.MinorValue, current.Major == original.Major)
	checkInt("patch", original.Patch, current.Patch, flags.PatchValue, current.Major == original.Major && current.Minor == original.Minor)

	baseSame := current.Major == original.Major && current.Minor == original.Minor && current.Patch == original.Patch

	if flags.StageValue != nil {
		stageName := strings.ToLower(flags.Stage)
		valid := baseSame && stageName != "" && original.Stage != nil && current.Stage != nil && strings.ToLower(original.StageName) == stageName
		checkPtr(stageName, original.Stage, current.Stage, flags.StageValue, valid)
	}

	if flags.EnvValue != nil {
		envName := strings.ToLower(flags.Env)
		if envName != "" {
			origEnvName, origEnvVal := envInfo(original)
			currEnvName, currEnvVal := envInfo(current)
			valid := baseSame && origEnvVal != nil && currEnvVal != nil && origEnvName == envName && currEnvName == envName
			checkPtr(envName, origEnvVal, currEnvVal, flags.EnvValue, valid)
		}
	}

	validRelease := baseSame && original.Release != nil && current.Release != nil
	checkPtr("release", original.Release, current.Release, flags.ReleaseValue, validRelease)

	return result
}

func formatDecreases(decreases []decrease) string {
	parts := make([]string, 0, len(decreases))
	for _, d := range decreases {
		parts = append(parts, fmt.Sprintf("%s from %d to %d", d.component, d.previous, d.current))
	}
	return fmt.Sprintf("numeric argument(s) went backwards: %s", strings.Join(parts, ", "))
}
