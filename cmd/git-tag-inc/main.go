package main

import (
	"flag"
	"github.com/arran4/git-tag-inc"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"log"
	"os"
)

var (
	verbose   = flag.Bool("verbose", false, "Extra output")
	dry       = flag.Bool("dry", false, "Dry run")
	ignore    = flag.Bool("ignore", true, "Ignore uncommitted files")
	repeating = flag.Bool("repeating", false, "Allow new tags to repeat a previous")
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

	if !*ignore {
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
			os.Exit(1)
			return
		}
	}

	major, minor, release, uat, test := gittaginc.CommandsToFlags(flag.Args())
	if !test && !uat && !release && !major && !minor {
		Usage()
		return
	}
	if (uat || test) && !*repeating {
		lastSimilar := FindHighestSimilarVersionTag(r, test, uat)
		var lstrh string
		lstr, err := r.Tag(lastSimilar.String())
		if err == git.ErrTagNotFound {

		} else if err != nil {
			panic(err)
		} else {
			lstrh = lstr.Hash().String()
		}
		ch, err := r.Head()
		if err != nil {
			panic(err)
		}
		chh := ch.Hash().String()
		if lstrh == chh {
			log.Printf("Hash is the same for this and previous tag: %s", lastSimilar.String())
			os.Exit(1)
			return
		}
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

func FindHighestSimilarVersionTag(r *git.Repository, test bool, uat bool) *gittaginc.Tag {
	if !uat && !test {
		return &gittaginc.Tag{}
	}
	return FindHVersionTag(r, func(last, current *gittaginc.Tag) bool {
		if test && current.Test == nil {
			return false
		}
		if uat && current.Uat == nil {
			return false
		}
		return last.LessThan(current)
	})
}

func FindHighestVersionTag(r *git.Repository) *gittaginc.Tag {
	return FindHVersionTag(r, func(last, current *gittaginc.Tag) bool {
		return last.LessThan(current)
	})
}

func FindHVersionTag(r *git.Repository, stop func(last, current *gittaginc.Tag) bool) *gittaginc.Tag {
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
		if stop(highest, t) {
			highest = t
		}
		return nil
	}); err != nil {
		panic(err)
	}
	return highest
}

func Usage() {
	flag.Usage()
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
