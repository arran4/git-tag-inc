package gittaginc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Tag struct {
	test    *int
	uat     *int
	release int
	major   int
	minor   int
	pad     int
}

func (t *Tag) LessThan(other *Tag) bool {
	if t.major < other.major {
		return true
	}
	if t.minor < other.minor {
		return true
	}
	if t.release < other.release {
		return true
	}
	if t.release > other.release {
		return false
	}
	var tv *int = nil
	if t.uat != nil {
		tv = t.uat
	} else if t.test != nil {
		tv = t.test
	}
	var ov *int = nil
	if other.uat != nil {
		ov = other.uat
	} else if other.test != nil {
		ov = other.test
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
		if other.uat != nil && t.test != nil {
			return true
		}
	}
	return false
}

func (t *Tag) String() string {
	ext := ""
	if t.uat != nil {
		ext = fmt.Sprintf("-uat%0"+fmt.Sprintf("%d", t.pad)+"d", *t.uat)
	} else if t.test != nil {
		ext = fmt.Sprintf("-test%0"+fmt.Sprintf("%d", t.pad)+"d", *t.test)
	}
	return fmt.Sprintf("v%d.%d.%d%s", t.major, t.minor, t.release, ext)
}

var ParseTagRe = regexp.MustCompile("^v(?:(\\d+)\\.(?:(\\d+)\\.(?:(\\d+))?)?)(?:(?U:-(.*))((?:0*)(\\d*)))?$")

func ParseTag(tag string) *Tag {
	m := ParseTagRe.FindStringSubmatch(tag)
	t := &Tag{}
	if len(m) == 0 {
		return nil
	}
	if len(m) > 0 {
		t.major, _ = strconv.Atoi(m[1])
	}
	if len(m) > 1 {
		t.minor, _ = strconv.Atoi(m[2])
	}
	if len(m) > 2 {
		t.release, _ = strconv.Atoi(m[3])
	}
	if len(m) > 5 && m[4] != "" {
		v1 := len(m[5])
		v2, _ := strconv.Atoi(m[6])
		switch strings.ToLower(m[4]) {
		case "test":
			t.pad = v1
			t.test = &v2
		case "uat":
			t.pad = v1
			t.uat = &v2
		default:
			return nil
		}
	}
	return t
}

func (t *Tag) Increment(major bool, minor bool, release bool, uat bool, test bool) {
	if major {
		t.major++
		t.minor = 0
		t.release = 0
		t.uat = nil
		t.test = nil
	}
	if minor {
		t.minor++
		t.release = 0
		t.uat = nil
		t.test = nil
	}
	var variant *int = nil
	if t.uat != nil {
		variant = t.uat
	}
	if t.test != nil {
		if variant != nil && *variant < *t.test || variant == nil {
			variant = t.test
		}
	}
	if release || (variant == nil && (uat || test) && !(minor || major)) {
		t.release += 1
		t.uat = nil
		t.test = nil
	}

	if uat {
		z := 1
		if variant != nil {
			z = *variant
			if t.test == nil {
				z = z + 1
			}
		} else {
			t.pad = 2
		}
		t.uat = &z
		t.test = nil
	}
	if test {
		z := 1
		if variant != nil {
			z = *variant + 1
		} else {
			t.pad = 2
		}
		t.test = &z
		t.uat = nil
	}
}
