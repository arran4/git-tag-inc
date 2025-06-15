package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/arran4/git-tag-inc"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/pkg/errors"
)

var (
	verbose          = flag.Bool("verbose", false, "Extra output")
	showVersion      = flag.Bool("version", false, "Print version information")
	dry              = flag.Bool("dry", false, "Dry run")
	printVersionOnly = flag.Bool("print-version-only", false, "Print next version only")
	ignore           = flag.Bool("ignore", true, "Ignore uncommitted files")
	repeating        = flag.Bool("repeating", false, "Allow new tags to repeat a previous")
	// TODO: consider supporting other naming modes such as "xyzzy",
	// "hybrid" or "octarine" which some teams use internally.
	mode = flag.String("mode", "default", "Naming mode: default or arraneous")
)

// nolint: gochecknoglobals
var (
	version = "dev"
	commit  = ""
	branch  = ""
	date    = ""
	builtBy = ""
	repo    = "https://github.com/arran4/git-tag-inc"
)

func main() {
	flag.Parse()
	if *printVersionOnly {
		*dry = true
		log.SetOutput(io.Discard)
	}
	if *showVersion {
		printVersion()
		return
	}
	gittaginc.Mode = *mode
	if !*verbose {
		log.SetFlags(0)
	}
	if *verbose {
		log.Printf("Version: %s (%s) by %s commit %s", version, date, builtBy, commit)
	}
	r, err := git.PlainOpen(".")
	if err != nil {
		panic(err)
	}

	cfg, cfgErr := r.ConfigScoped(config.SystemScope)
	var tagger *object.Signature
	if cfgErr == nil {
		if cfg.User.Name == "" || cfg.User.Email == "" {
			log.Printf("git user.name or user.email not configured")
			log.Printf("Run `git config --global user.name \"Your Name\"` and `git config --global user.email \"you@example.com\"`")
			os.Exit(1)
			return
		}
		tagger = &object.Signature{
			Name:  cfg.User.Name,
			Email: cfg.User.Email,
			When:  time.Now(),
		}
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

	flags := gittaginc.CommandsToFlags(flag.Args(), *mode)
	if !flags.Valid || (!flags.Major && !flags.Minor && !flags.Patch && !flags.Release && flags.Env == "" && flags.Stage == "") {
		Usage()
		return
	}
	currentHash, err := GetHash(r, nil)
	if err != nil {
		panic(err)
	}
	if !*repeating && currentHash != "" {
		lastSimilar := FindHighestSimilarVersionTag(r, flags.Env)
		if lastSimilar != nil {
			lastSimilarHash, err := GetHash(r, lastSimilar)
			if err != nil {
				switch {
				case errors.Is(err, plumbing.ErrObjectNotFound):
				default:
					panic(err)
				}
			} else {
				if len(lastSimilarHash) > 0 && lastSimilarHash == currentHash {
					log.Printf("Hash is the same for this and previous tag: (%s) %s and %s", lastSimilar, lastSimilarHash, currentHash)
					os.Exit(1)
					return
				}
			}
		}
	}

	highest := FindHighestVersionTag(r)

	log.Printf("Largest: %s (%s)", highest, currentHash)

	highest.Increment(flags.Major, flags.Minor, flags.Patch, flags.Stage, flags.Env, flags.Release)

	log.Printf("Creating %s", highest)
	if *printVersionOnly {
		fmt.Println(highest.String())
		return
	}

	h, err := r.Head()
	if err != nil {
		panic(err)
	}
	if !*dry {
		_, err = r.CreateTag(highest.String(), h.Hash(), &git.CreateTagOptions{
			Message: highest.String(),
			Tagger:  tagger,
		})
	} else {
		log.Printf("Dry run finished.")
	}
	if err != nil {
		panic(err)
	}
}

func GetHash(r *git.Repository, lastSimilar *gittaginc.Tag) (string, error) {
	var err error
	var ref *plumbing.Reference
	var to *object.Tag
	if lastSimilar != nil {
		ref, err = r.Tag(lastSimilar.String())
		if err == git.ErrTagNotFound {
			return "", nil
		} else if err != nil {
			return "", err
		}
		to, err = r.TagObject(ref.Hash())
		if err == git.ErrTagNotFound {
			return "", nil
		} else if err != nil {
			return "", err
		}
		return to.Target.String(), nil
	} else {
		ref, err = r.Head()
		if err == git.ErrTagNotFound {
			return "", nil
		} else if err != nil {
			return "", err
		}
		return ref.Hash().String(), nil
	}
}

func FindHighestSimilarVersionTag(r *git.Repository, env string) *gittaginc.Tag {
	return FindHVersionTag(r, func(last, current *gittaginc.Tag) bool {
		if env == "test" && current.Test == nil {
			return false
		}
		if env == "uat" && current.Uat == nil {
			return false
		}
		if env == "" && (current.Uat != nil || current.Test != nil) {
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
	patchName := "patch"
	if gittaginc.Mode == "arraneous" {
		patchName = "release"
	}
	log.Printf("  - %s        => v0.0.1-test1 => v0.0.2       ", patchName)
	if gittaginc.Mode != "arraneous" {
		log.Printf("  - release      => v0.0.1-test1 => v0.0.1-test2 ")
		log.Printf("  - release      => v0.0.1 => v0.0.1.1")
	}
	log.Printf("  - test         => v0.0.1-test1 => v0.0.1-test2 ")
	log.Printf("  - uat          => v0.0.1-uat1 => v0.0.1-uat2  ")
	log.Printf("  - alpha        => v0.0.1-alpha1 => v0.0.1-alpha2 ")
	log.Printf("  - beta         => v0.0.1-beta1 => v0.0.1-beta2 ")
	log.Printf("  - rc           => v0.0.1-rc1 => v0.0.1-rc2 ")
	log.Printf("Combinations work:")
	log.Printf("  - patch test   => v0.0.1-test1 => v0.1.0-test1  ")
	log.Printf("Duplications don't:")
	log.Printf("  - test test    => v0.0.1-test1 => v0.0.1-test2  ")
	log.Printf("Use --mode arraneous to switch to legacy naming")
}

func printVersion() {
	fmt.Printf("git-tag-inc version %s\n", version)
	fmt.Printf("commit: %s\n", commit)
	fmt.Printf("branch: %s\n", branch)
	fmt.Printf("built: %s by %s\n", date, builtBy)
	fmt.Printf("repo: %s\n", repo)
	fmt.Printf("credits: Arran Ubels\n")
}
