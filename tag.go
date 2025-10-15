package gittaginc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Mode controls which naming scheme to use.
// Valid values: "default", "arraneous".
var Mode string

type Tag struct {
	StageName string
	Stage     *int
	StagePad  int

	Test *int
	Uat  *int
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
	if t.Test != nil {
		v := *t.Test
		clone.Test = &v
	}
	if t.Uat != nil {
		v := *t.Uat
		clone.Uat = &v
	}
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
	switch strings.ToLower(n) {
	case "alpha":
		return 0
	case "beta":
		return 1
	case "rc":
		return 2
	default:
		if n == "" {
			return 3
		}
		return 4
	}
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

	var tv *int = nil
	if t.Uat != nil {
		tv = t.Uat
	} else if t.Test != nil {
		tv = t.Test
	}
	var ov *int = nil
	if other.Uat != nil {
		ov = other.Uat
	} else if other.Test != nil {
		ov = other.Test
	}
	if tv == nil {
		return false
	}
	if ov == nil {
		return true
	}
	if *tv < *ov {
		return true
	}
	if *tv == *ov {
		if other.Uat != nil && t.Test != nil {
			return true
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
	if t.Stage != nil {
		ext += fmt.Sprintf("-%s%0*d", t.StageName, t.StagePad, *t.Stage)
	}
	if t.Uat != nil {
		ext += fmt.Sprintf("-uat%0*d", t.Pad, *t.Uat)
	} else if t.Test != nil {
		ext += fmt.Sprintf("-test%0*d", t.Pad, *t.Test)
	}
	if t.Release != nil {
		ext += fmt.Sprintf(".%d", *t.Release)
	}
	return fmt.Sprintf("v%d.%d.%d%s", t.Major, t.Minor, t.Patch, ext)
}

var ParseTagRe = regexp.MustCompile(`^v(\d+)\.(\d+)\.(\d+)(?:-((?:alpha|beta|rc))((?:0*)(\d+)))?(?:-((?:test|uat))((?:0*)(\d+)))?(?:\.(\d+))?$`)

func ParseTag(tag string) *Tag {
	m := ParseTagRe.FindStringSubmatch(tag)
	t := &Tag{}
	if len(m) == 0 {
		return nil
	}
	t.Major, _ = strconv.Atoi(m[1])
	t.Minor, _ = strconv.Atoi(m[2])
	t.Patch, _ = strconv.Atoi(m[3])
	if m[4] != "" {
		t.StageName = strings.ToLower(m[4])
		t.StagePad = len(m[5])
		v, _ := strconv.Atoi(m[6])
		t.Stage = &v
	}
	if m[7] != "" {
		t.Pad = len(m[8])
		v, _ := strconv.Atoi(m[9])
		switch strings.ToLower(m[7]) {
		case "test":
			t.Test = &v
		case "uat":
			t.Uat = &v
		default:
			return nil
		}
	}
	if len(m) >= 11 && m[10] != "" {
		v, _ := strconv.Atoi(m[10])
		t.Release = &v
	}
	return t
}

func (t *Tag) applyIncrement(flags CmdFlags) {
	prevStage := t.Stage
	prevStageName := strings.ToLower(t.StageName)
	prevStagePad := t.StagePad
	prevEnv := t.Uat
	prevEnvType := "uat"
	if prevEnv == nil {
		prevEnv = t.Test
		prevEnvType = "test"
	}
	if prevEnv == nil {
		prevEnvType = ""
	}
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
		t.Uat = nil
		t.Test = nil
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
		t.Uat = nil
		t.Test = nil
		prevStage = nil
		prevStageName = ""
		prevEnv = nil
		prevEnvType = ""
	}
	if flags.Patch {
		target := t.Patch
		if flags.PatchValue != nil {
			target = *flags.PatchValue
		} else if (t.Test == nil || flags.Env != "") && (t.Uat == nil || flags.Env != "") && (t.Stage == nil || flags.Stage != "") {
			target = t.Patch + 1
		}
		t.Patch = target
		t.Stage = nil
		t.StageName = ""
		t.StagePad = 0
		t.Uat = nil
		t.Test = nil
		t.Release = nil
		prevStage = nil
		prevStageName = ""
		prevEnv = nil
		prevEnvType = ""
	}
	if flags.Stage != "" {
		stageName := strings.ToLower(flags.Stage)
		stagePad := 2
		if flags.StageDigits > 0 {
			stagePad = flags.StageDigits
		} else if prevStage != nil && prevStageName == stageName && prevStagePad > 0 {
			stagePad = prevStagePad
		}
		z := 1
		if flags.StageValue != nil {
			z = *flags.StageValue
		} else if prevStage != nil && prevStageName == stageName {
			z = *prevStage + 1
		} else if !flags.Major && !flags.Minor && !flags.Patch {
			t.Patch += 1
		}
		t.Stage = pi(z)
		t.StagePad = stagePad
		t.StageName = stageName
		prevEnv = nil
		prevEnvType = ""
		prevPad = 0
		t.Uat = nil
		t.Test = nil
		t.Release = nil
	}

	if flags.Env != "" {
		envName := strings.ToLower(flags.Env)
		envPad := 2
		if flags.EnvDigits > 0 {
			envPad = flags.EnvDigits
		} else if prevEnv != nil && prevPad > 0 {
			envPad = prevPad
		}
		z := 1
		if prevEnv != nil {
			if prevEnvType == "uat" && envName == "uat" {
				z = *prevEnv + 1
			} else if prevEnvType == "test" && envName == "test" {
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
		if envName == "uat" {
			t.Uat = pi(z)
			t.Test = nil
		} else {
			t.Test = pi(z)
			t.Uat = nil
		}
		t.Release = nil
	}

	if flags.Release {
		target := 1
		if flags.ReleaseValue != nil {
			target = *flags.ReleaseValue
		} else if t.Release != nil {
			target = *t.Release + 1
		}
		t.Release = pi(target)
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
		autoFlags.PatchValue = pi(original.Patch + 1)
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
	if tag.Uat != nil {
		return "uat", tag.Uat
	}
	if tag.Test != nil {
		return "test", tag.Test
	}
	return "", nil
}

func detectDecreases(original, current *Tag, flags CmdFlags) []decrease {
	var result []decrease

	if flags.MajorValue != nil && *flags.MajorValue < original.Major {
		result = append(result, decrease{component: "major", previous: original.Major, current: *flags.MajorValue})
	}

	if flags.MinorValue != nil && current.Major == original.Major && *flags.MinorValue < original.Minor {
		result = append(result, decrease{component: "minor", previous: original.Minor, current: *flags.MinorValue})
	}

	if flags.PatchValue != nil && current.Major == original.Major && current.Minor == original.Minor && *flags.PatchValue < original.Patch {
		result = append(result, decrease{component: "patch", previous: original.Patch, current: *flags.PatchValue})
	}

	baseSame := current.Major == original.Major && current.Minor == original.Minor && current.Patch == original.Patch

	if flags.StageValue != nil && baseSame {
		stageName := strings.ToLower(flags.Stage)
		if stageName != "" && original.Stage != nil && current.Stage != nil && strings.ToLower(original.StageName) == stageName {
			if *current.Stage < *original.Stage {
				result = append(result, decrease{component: stageName, previous: *original.Stage, current: *current.Stage})
			}
		}
	}

	if flags.EnvValue != nil && baseSame {
		envName := strings.ToLower(flags.Env)
		if envName != "" {
			origEnvName, origEnvVal := envInfo(original)
			currEnvName, currEnvVal := envInfo(current)
			if origEnvVal != nil && currEnvVal != nil && origEnvName == envName && currEnvName == envName {
				if *currEnvVal < *origEnvVal {
					result = append(result, decrease{component: envName, previous: *origEnvVal, current: *currEnvVal})
				}
			}
		}
	}

	if flags.ReleaseValue != nil && baseSame && original.Release != nil && current.Release != nil {
		if *current.Release < *original.Release {
			result = append(result, decrease{component: "release", previous: *original.Release, current: *current.Release})
		}
	}

	return result
}

func formatDecreases(decreases []decrease) string {
	parts := make([]string, 0, len(decreases))
	for _, d := range decreases {
		parts = append(parts, fmt.Sprintf("%s from %d to %d", d.component, d.previous, d.current))
	}
	return fmt.Sprintf("numeric argument(s) went backwards: %s", strings.Join(parts, ", "))
}
