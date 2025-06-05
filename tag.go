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

func (t *Tag) Increment(major bool, minor bool, patch bool, stage string, env string, rel bool) {
	prevStage := t.Stage
	prevStageName := t.StageName
	prevEnv := t.Uat
	if prevEnv == nil {
		prevEnv = t.Test
	}
	if major {
		t.Major++
		t.Minor = 0
		t.Patch = 0
		t.Release = nil
		t.Stage = nil
		t.StageName = ""
		t.Uat = nil
		t.Test = nil
		prevStage = nil
		prevEnv = nil
	}
	if minor {
		t.Minor++
		t.Patch = 0
		t.Release = nil
		t.Stage = nil
		t.StageName = ""
		t.Uat = nil
		t.Test = nil
		prevStage = nil
		prevEnv = nil
	}
	if patch {
		if (t.Test == nil || env != "") && (t.Uat == nil || env != "") && (t.Stage == nil || stage != "") {
			t.Patch += 1
		}
		t.Stage = nil
		t.StageName = ""
		t.Uat = nil
		t.Test = nil
		t.Release = nil
		prevStage = nil
		prevEnv = nil
	}
	if stage != "" {
		z := 1
		if prevStage != nil && prevStageName == stage {
			z = *prevStage + 1
		} else if !major && !minor && !patch {
			t.Patch += 1
		}
		t.Stage = &z
		t.StagePad = 2
		t.StageName = stage
		prevEnv = nil
		t.Uat = nil
		t.Test = nil
		t.Release = nil
	}

	if env != "" {
		z := 1
		if prevEnv != nil {
			if t.Uat != nil && env == "uat" {
				z = *prevEnv + 1
			} else if t.Test != nil && env == "test" {
				z = *prevEnv + 1
			} else {
				z = *prevEnv
			}
		} else if !major && !minor && !patch && stage == "" && prevStage == nil {
			t.Patch += 1
		}
		if env == "uat" {
			t.Uat = &z
			t.Test = nil
		} else if env == "test" {
			t.Test = &z
			t.Uat = nil
		}
		t.Pad = 2
		t.Release = nil
	}

	if rel {
		if t.Release != nil {
			*t.Release = *t.Release + 1
		} else {
			v := 1
			t.Release = &v
		}
	}
}
