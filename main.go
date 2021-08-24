package main

import (
	"flag"
	"fmt"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"log"
	"regexp"
	"strconv"
	"strings"
) // with go modules enabled (GO111MODULE=on or outside GOPATH)

var (
	verbose = flag.Bool("verbose", false, "Extra output")
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

func main() {
	flag.Parse()
	r, err := git.PlainOpen(".git")
	if err != nil {
		panic(err)
	}
	test := false
	uat := false
	release := false
	major := false
	minor := false
	d := 0
	for _, f := range flag.Args() {
		switch strings.ToLower(f) {
		case "test":
			test = true
		case "uat":
			uat = true
		case "release":
			release = true
		case "major":
			major = true
		case "minor":
			minor = true
		default:
			Usage()
			return
		}
		d++
	}
	if d == 0 {
		Usage()
		return
	}
	highest := FindHighestVersionTag(r)
	log.Printf("Largest: %s", highest)
	highest.Increment(major, minor, release, uat, test)

	log.Printf("Creating %s", highest)

	h, err := r.Head()
	if err != nil {
		panic(err)
	}
	_, err = r.CreateTag(highest.String(), h.Hash(), &git.CreateTagOptions{
		Message: highest.String(),
	})
	if err != nil {
		panic(err)
	}
}

func FindHighestVersionTag(r *git.Repository) *Tag {
	iter, err := r.Tags()
	if err != nil {
		panic(err)
	}
	var highest *Tag = &Tag{}
	if err := iter.ForEach(func(ref *plumbing.Reference) error {
		if *verbose {
			log.Printf("Ref: %s", ref.Name())
		}
		t := ParseTag(ref.Name().Short())
		if t == nil {
			return nil
		}
		if highest.LessThan(t) {
			highest = t
		}
		return nil
	}); err != nil {
		panic(err)
	}
	return highest
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

func Usage() {
	log.Printf("You're using this wrong")
	log.Printf("git-tag-inc then, one or more of: ")
	log.Printf("  - major        => v0.0.1-test1 => v1.0.0       ")
	log.Printf("  - minor        => v0.0.1-test1 => v0.1.0       ")
	log.Printf("  - release      => v0.0.1-test1 => v0.0.2       ")
	log.Printf("  - test         => v0.0.1-test1 => v0.0.1-test2 ")
	log.Printf("  - test         => v0.0.1-uat1  => v0.0.1-test2 ")
	log.Printf("  - uat          => v0.0.1-test3 => v0.0.1-uat3  ")
	log.Printf("  - uat          => v0.0.1-uat1  => v0.0.1-uat2  ")
	log.Printf("Combinations work:")
	log.Printf("  - release test => v0.0.1-test1 => v0.1.0-test1  ")
	log.Printf("Duplications don't:")
	log.Printf("  - test test    => v0.0.1-test1 => v0.0.1-test2  ")
}
