package main

import (
	"flag"
	"github.com/arran4/git-tag-inc"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"log"
	"strings"
)

var (
	verbose = flag.Bool("verbose", false, "Extra output")
)

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

func main() {
	flag.Parse()
	log.Printf("Version: %s (%s) by %s commit %s", version, date, builtBy, commit)
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

func FindHighestVersionTag(r *git.Repository) *gittaginc.Tag {
	iter, err := r.Tags()
	if err != nil {
		panic(err)
	}
	var highest *gittaginc.Tag = &gittaginc.Tag{}
	if err := iter.ForEach(func(ref *plumbing.Reference) error {
		if *verbose {
			log.Printf("Ref: %s", ref.Name())
		}
		t := gittaginc.ParseTag(ref.Name().Short())
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
