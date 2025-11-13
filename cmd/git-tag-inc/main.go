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
	allowBackwards   = flag.Bool("allow-backwards", false, "Allow numeric arguments to decrease version counters")
	skipForwards     = flag.Bool("skip-forwards", false, "Automatically bump the patch when numeric arguments go backwards")
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
	flag.Usage = Usage
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

	if err := highest.Increment(flags, *allowBackwards, *skipForwards); err != nil {
		log.Printf("%v", err)
		os.Exit(1)
	}

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
	log.Printf("Usage of git-tag-inc:")
	log.Printf("./git-tag-inc [--allow-backwards] [--skip-forwards] [major[<n>]] [minor[<n>]] [patch[<n>]] [release[<n>]] [alpha|beta|rc[<n>]] [test|uat[<n>]]")
	log.Printf("--version [--print-version-only]")
	log.Printf("")
	log.Printf("Use --version to display build information and credits.")
	log.Printf("Use --print-version-only to output the next version without tagging.")
	log.Printf("")
	log.Printf("--mode arraneous switches to the legacy naming (patch becomes `release`).")
	log.Printf("")
	log.Printf("Numeric suffixes can be added to any command to set a specific counter. For example,")
	log.Printf("`test5` produces `-test5`, `rc02` produces `-rc02` and `major3` moves directly to")
	log.Printf("`v3.0.0`. When a numeric suffix would decrease a counter compared to the previous tag")
	log.Printf("the command fails unless either `--allow-backwards` is provided or `--skip-forwards`")
	log.Printf("is used. `--allow-backwards` applies the requested number directly, while")
	log.Printf("`--skip-forwards` automatically bumps the patch component first so the resulting tag")
	log.Printf("still increases. For instance, `git-tag-inc --skip-forwards test2` upgrades")
	log.Printf("`v1.0.0-test3` to `v1.0.1-test2`.")
	log.Printf("")
	log.Printf("git-tag-inc then, one or more of:")
	log.Printf("* `major        => v0.0.1-test1 => v1.0.0`")
	log.Printf("* `minor        => v0.0.1-test1 => v0.1.0`")
	patchName := "patch"
	if gittaginc.Mode == "arraneous" {
		patchName = "release"
	}
	log.Printf("* `%s        => v0.0.1-test1 => v0.0.2`", patchName)
	if gittaginc.Mode != "arraneous" {
		log.Printf("* `release      => v0.0.1-test1 => v0.0.1-test2`")
		log.Printf("* `release      => v0.0.1 => v0.0.1.1`")
	}
	log.Printf("* `test         => v0.0.1-test1 => v0.0.1-test2`")
	log.Printf("* `uat          => v0.0.1-uat1  => v0.0.1-uat2`")
	log.Printf("* `alpha        => v0.0.1-alpha1 => v0.0.1-alpha2`")
	log.Printf("* `beta         => v0.0.1-beta1  => v0.0.1-beta2`")
	log.Printf("* `rc           => v0.0.1-rc1    => v0.0.1-rc2`")
	log.Printf("* `rc5          => v0.0.1-rc1    => v0.0.1-rc5`")
	log.Printf("* `major4       => v0.0.1        => v4.0.0`")
	log.Printf("")
	log.Printf("Combinations work:")
	log.Printf("* `patch test   => v0.0.1-test1 => v0.1.0-test1`")
	log.Printf("* `patch rc2    => v0.1.0-rc4  => v0.1.1-rc2`")
	log.Printf("")
	log.Printf("Preventing backwards moves:")
	log.Printf("* `test1` (when the last tag was `test3`) errors unless `--allow-backwards` is supplied.")
	log.Printf("* `--skip-forwards test1` turns the same command into `vX.Y.(Z+1)-test1` automatically.")
	log.Printf("")
	log.Printf("Duplications don't:")
	log.Printf("* `test test    => v0.0.1-test1 => v0.0.1-test2`")
}

func printVersion() {
	fmt.Printf("git-tag-inc version %s\n", version)
	fmt.Printf("commit: %s\n", commit)
	fmt.Printf("branch: %s\n", branch)
	fmt.Printf("built: %s by %s\n", date, builtBy)
	fmt.Printf("repo: %s\n", repo)
	fmt.Printf("credits: Arran Ubels\n")
}
