package gittaginc

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

type Tag struct {
	Test    *int
	Uat     *int
	Release int
	Major   int
	Minor   int
	Pad     int
}

func (t *Tag) LessThan(other *Tag) bool {
	if t.Major != other.Major {
		return t.Major < other.Major
	}
	if t.Minor != other.Minor {
		return t.Minor < other.Minor
	}
	if t.Release != other.Release {
		return t.Release < other.Release
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
	return false
}

func (t *Tag) String() string {
	ext := ""
	if t.Uat != nil {
		ext = fmt.Sprintf("-uat%0"+fmt.Sprintf("%d", t.Pad)+"d", *t.Uat)
	} else if t.Test != nil {
		ext = fmt.Sprintf("-test%0"+fmt.Sprintf("%d", t.Pad)+"d", *t.Test)
	}
	return fmt.Sprintf("v%d.%d.%d%s", t.Major, t.Minor, t.Release, ext)
}

var ParseTagRe = regexp.MustCompile("^v(?:(\\d+)\\.(?:(\\d+)\\.(?:(\\d+))?)?)(?:(?U:-(.*))((?:0*)(\\d*)))?$")

func ParseTag(tag string) *Tag {
	m := ParseTagRe.FindStringSubmatch(tag)
	t := &Tag{}
	if len(m) == 0 {
		return nil
	}
	if len(m) > 0 {
		t.Major, _ = strconv.Atoi(m[1])
	}
	if len(m) > 1 {
		t.Minor, _ = strconv.Atoi(m[2])
	}
	if len(m) > 2 {
		t.Release, _ = strconv.Atoi(m[3])
	}
	if len(m) > 5 && m[4] != "" {
		v1 := len(m[5])
		v2, _ := strconv.Atoi(m[6])
		switch strings.ToLower(m[4]) {
		case "test":
			t.Pad = v1
			t.Test = &v2
		case "uat":
			t.Pad = v1
			t.Uat = &v2
		default:
			return nil
		}
	}
	return t
}

func (t *Tag) Increment(major bool, minor bool, release bool, uat bool, test bool) {
	puat := t.Uat
	ptest := t.Test
	if major {
		t.Major++
		t.Minor = 0
		t.Release = 0
		t.Uat = nil
		t.Test = nil
		puat = pi(0)
		ptest = pi(0)
	}
	if minor {
		t.Minor++
		t.Release = 0
		t.Uat = nil
		t.Test = nil
		puat = pi(0)
		ptest = pi(0)
	}
	if release {
		if (t.Test == nil || test) && (t.Uat == nil || uat) {
			t.Release += 1
		}
		t.Uat = nil
		t.Test = nil
		puat = pi(0)
		ptest = pi(0)
	}

	if uat {
		z := 1
		if puat != nil {
			z = *puat
			z = z + 1
		} else if ptest != nil {
			z = *ptest
		} else if !release && !major && !minor {
			t.Release += 1
			t.Pad = 2
		} else {
			t.Pad = 2
		}
		t.Uat = &z
		t.Test = nil
	}
	if test {
		z := 1
		if puat != nil {
			z = *puat
			z = z + 1
		} else if ptest != nil {
			z = *ptest
			z = z + 1
		} else if !release && !major && !minor {
			t.Release += 1
			t.Pad = 2
		} else {
			t.Pad = 2
		}
		t.Test = &z
		t.Uat = nil
	}
}
