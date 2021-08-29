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
	puat := t.uat
	ptest := t.test
	if major {
		t.major++
		t.minor = 0
		t.release = 0
		t.uat = nil
		t.test = nil
		puat = pi(0)
		ptest = pi(0)
	}
	if minor {
		t.minor++
		t.release = 0
		t.uat = nil
		t.test = nil
		puat = pi(0)
		ptest = pi(0)
	}
	if release {
		t.release += 1
		t.uat = nil
		t.test = nil
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
		} else {
			t.release += 1
			t.pad = 2
		}
		t.uat = &z
		t.test = nil
	}
	if test {
		z := 1
		if puat != nil {
			z = *puat
			z = z + 1
		} else if ptest != nil {
			z = *ptest
			z = z + 1
		} else {
			t.release += 1
			t.pad = 2
		}
		t.test = &z
		t.uat = nil
	}
}
