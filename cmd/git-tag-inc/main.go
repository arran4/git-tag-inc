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
	force            = flag.Bool("force", false, "Force the operation (implies --allow-backwards, --repeating, --ignore)")
	mode             = flag.String("mode", "default", "Naming mode: default or arraneous")
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

	if *force {
		*allowBackwards = true
		*repeating = true
		*ignore = true
	}

	if *printVersionOnly {
		*dry = true
		log.SetOutput(io.Discard)
	}
	if *showVersion {
		printVersion()
		return
	}
	if !*verbose {
		log.SetFlags(0)
	}
	if *verbose {
		log.Printf("Version: %s (%s) by %s commit %s", version, date, builtBy, commit)
	}
	r, err := git.PlainOpen(".")
	if err != nil {
		if errors.Is(err, git.ErrRepositoryNotExists) {
			log.Printf("Error: %v. Are you in a git repository?", err)
		} else {
			log.Printf("Error opening repository: %v", err)
		}
		os.Exit(1)
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
			log.Printf("Failed to get worktree: %v", err)
			os.Exit(1)
		}
		s, err := wt.Status()
		if err != nil {
			log.Printf("Failed to get worktree status: %v", err)
			os.Exit(1)
		}
		if !s.IsClean() {
			log.Printf("There are uncommitted changes in this repo.")
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
		log.Printf("Failed to get current hash: %v", err)
		os.Exit(1)
	}
	if !*repeating && currentHash != "" {
		lastSimilar, err := FindHighestSimilarVersionTag(r, flags.Env)
		if err != nil {
			log.Printf("Failed to find highest similar version tag: %v", err)
			os.Exit(1)
		}
		if lastSimilar != nil {
			lastSimilarHash, err := GetHash(r, lastSimilar)
			if err != nil {
				switch {
				case errors.Is(err, plumbing.ErrObjectNotFound):
				default:
					log.Printf("Failed to get hash for similar version: %v", err)
					os.Exit(1)
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

	highest, err := FindHighestVersionTag(r)
	if err != nil {
		log.Printf("Failed to find highest version tag: %v", err)
		os.Exit(1)
	}

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
		log.Printf("Failed to get HEAD: %v", err)
		os.Exit(1)
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
		log.Printf("Failed to create tag: %v", err)
		os.Exit(1)
	}
}

func GetHash(r *git.Repository, lastSimilar *gittaginc.Tag) (string, error) {
	var err error
	var ref *plumbing.Reference
	var to *object.Tag
	if lastSimilar != nil {
		var hash plumbing.Hash
		if lastSimilar.Hash != "" {
			hash = plumbing.NewHash(lastSimilar.Hash)
		} else {
			ref, err = r.Tag(lastSimilar.String())
			if err == git.ErrTagNotFound {
				return "", nil
			} else if err != nil {
				return "", err
			}
			hash = ref.Hash()
		}
		to, err = r.TagObject(hash)
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

func FindHighestSimilarVersionTag(r *git.Repository, env string) (*gittaginc.Tag, error) {
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

func FindHighestVersionTag(r *git.Repository) (*gittaginc.Tag, error) {
	return FindHVersionTag(r, func(last, current *gittaginc.Tag) bool {
		return last.LessThan(current)
	})
}

func FindHVersionTag(r *git.Repository, stop func(last, current *gittaginc.Tag) bool) (*gittaginc.Tag, error) {
	iter, err := r.Tags()
	if err != nil {
		return nil, err
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
		t.Hash = ref.Hash().String()
		if stop(highest, t) {
			highest = t
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return highest, nil
}

func Usage() {
	out := flag.CommandLine.Output()
	fmt.Fprintf(out, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(out, "%s [--allow-backwards] [--skip-forwards] [major[<n>]] [minor[<n>]] [patch[<n>]] [release[<n>]] [alpha|beta|rc[<n>]] [test|uat[<n>]]\n", os.Args[0])
	fmt.Fprintf(out, "\nFlags:\n")
	flag.PrintDefaults()
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Use --version to display build information and credits.\n")
	fmt.Fprintf(out, "Use --print-version-only to output the next version without tagging.\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "--mode %s switches to the legacy naming (patch becomes `release`).\n", gittaginc.ModeArraneous)
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Numeric suffixes can be added to any command to set a specific counter. For example,\n")
	fmt.Fprintf(out, "`test5` produces `-test5`, `rc02` produces `-rc02` and `major3` moves directly to\n")
	fmt.Fprintf(out, "`v3.0.0`. When a numeric suffix would decrease a counter compared to the previous tag\n")
	fmt.Fprintf(out, "the command fails unless either `--allow-backwards` is provided or `--skip-forwards`\n")
	fmt.Fprintf(out, "is used. `--allow-backwards` applies the requested number directly, while\n")
	fmt.Fprintf(out, "`--skip-forwards` automatically bumps the patch component first so the resulting tag\n")
	fmt.Fprintf(out, "still increases. For instance, `git-tag-inc --skip-forwards test2` upgrades\n")
	fmt.Fprintf(out, "`v1.0.0-test3` to `v1.0.1-test2`.\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "git-tag-inc then, one or more of:\n")
	fmt.Fprintf(out, "* `major        => v0.0.1-test1 => v1.0.0`\n")
	fmt.Fprintf(out, "* `minor        => v0.0.1-test1 => v0.1.0`\n")
	patchName := "patch"
	if *mode == gittaginc.ModeArraneous {
		patchName = "release"
	}
	fmt.Fprintf(out, "* `%s        => v0.0.1-test1 => v0.0.2`\n", patchName)
	if *mode != gittaginc.ModeArraneous {
		fmt.Fprintf(out, "* `release      => v0.0.1-test1 => v0.0.1-test2`\n")
		fmt.Fprintf(out, "* `release      => v0.0.1 => v0.0.1.1`\n")
	}
	fmt.Fprintf(out, "* `test         => v0.0.1-test1 => v0.0.1-test2`\n")
	fmt.Fprintf(out, "* `uat          => v0.0.1-uat1  => v0.0.1-uat2`\n")
	fmt.Fprintf(out, "* `alpha        => v0.0.1-alpha1 => v0.0.1-alpha2`\n")
	fmt.Fprintf(out, "* `beta         => v0.0.1-beta1  => v0.0.1-beta2`\n")
	fmt.Fprintf(out, "* `rc           => v0.0.1-rc1    => v0.0.1-rc2`\n")
	fmt.Fprintf(out, "* `rc5          => v0.0.1-rc1    => v0.0.1-rc5`\n")
	fmt.Fprintf(out, "* `major4       => v0.0.1        => v4.0.0`\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Combinations work:\n")
	fmt.Fprintf(out, "* `patch test   => v0.0.1-test1 => v0.1.0-test1`\n")
	fmt.Fprintf(out, "* `patch rc2    => v0.1.0-rc4  => v0.1.1-rc2`\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Preventing backwards moves:\n")
	fmt.Fprintf(out, "* `test1` (when the last tag was `test3`) errors unless `--allow-backwards` is supplied.\n")
	fmt.Fprintf(out, "* `--skip-forwards test1` turns the same command into `vX.Y.(Z+1)-test1` automatically.\n")
	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "Duplications don't:\n")
	fmt.Fprintf(out, "* `test test    => v0.0.1-test1 => v0.0.1-test2`\n")
}

func printVersion() {
	fmt.Printf("git-tag-inc version %s\n", version)
	fmt.Printf("commit: %s\n", commit)
	fmt.Printf("branch: %s\n", branch)
	fmt.Printf("built: %s by %s\n", date, builtBy)
	fmt.Printf("repo: %s\n", repo)
	fmt.Printf("credits: Arran Ubels\n")
}
