package main

import (
	"flag"
	"github.com/arran4/git-tag-inc"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"log"
)

var (
	verbose = flag.Bool("verbose", false, "Extra output")
	dry     = flag.Bool("dry", false, "Dry run")
	check   = flag.Bool("check", true, "Check if there are uncommitted files in repo before running")
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
	if !*verbose {
		log.SetFlags(0)
	}
	log.Printf("Version: %s (%s) by %s commit %s", version, date, builtBy, commit)
	r, err := git.PlainOpen(".")
	if err != nil {
		panic(err)
	}

	if *check {
		wt, err := r.Worktree()
		if err != nil {
			panic(err)
		}
		s, err := wt.Status()
		if err != nil {
			panic(err)
		}
		if !s.IsClean() {
			log.Printf("There are uncommited changes in thils repo.")
			return
		}
	}

	major, minor, release, uat, test := gittaginc.CommandsToFlags(flag.Args())
	if !test && !uat && !release && !major && !minor {
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
	if !*dry {
		_, err = r.CreateTag(highest.String(), h.Hash(), &git.CreateTagOptions{
			Message: highest.String(),
		})
	} else {
		log.Printf("Dry run finished.")
	}
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
