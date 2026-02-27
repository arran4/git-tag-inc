// Copyright (c) 2025, Arran Ubels
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"bytes"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"text/template"
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
	// TODO: consider supporting other naming modes such as "xyzzy",
	// "hybrid" or "octarine" which some teams use internally.
	mode = flag.String("mode", "default", "Naming mode: default or arraneous")

	out io.Writer = os.Stderr
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
		out = io.Discard
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
		fmt.Fprintf(out, "Version: %s (%s) by %s commit %s\n", version, date, builtBy, commit)
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
			fmt.Fprintf(out, "git user.name or user.email not configured\n")
			fmt.Fprintf(out, "Run `git config --global user.name \"Your Name\"` and `git config --global user.email \"you@example.com\"`\n")
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
			fmt.Fprintf(out, "There are uncommited changes in thils repo.\n")
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
			fmt.Fprintf(out, "Failed to find highest similar version tag: %v", err)
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
					fmt.Fprintf(out, "Hash is the same for this and previous tag: (%s) %s and %s\n", lastSimilar, lastSimilarHash, currentHash)
					os.Exit(1)
					return
				}
			}
		}
	}

	highest, err := FindHighestVersionTag(r)
	if err != nil {
		fmt.Fprintf(out, "Failed to find highest version tag: %v", err)
		os.Exit(1)
	}

	fmt.Fprintf(out, "Largest: %s (%s)\n", highest, currentHash)

	if err := highest.Increment(flags, *allowBackwards, *skipForwards); err != nil {
		fmt.Fprintf(out, "%v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(out, "Creating %s\n", highest)
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
		fmt.Fprintf(out, "Dry run finished.\n")
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
	t, err := FindHVersionTag(r, func(last, current *gittaginc.Tag) bool {
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
	return t, err
}

func FindHighestVersionTag(r *git.Repository) (*gittaginc.Tag, error) {
	t, err := FindHVersionTag(r, func(last, current *gittaginc.Tag) bool {
		return last.LessThan(current)
	})
	return t, err
}

func FindHVersionTag(r *git.Repository, stop func(last, current *gittaginc.Tag) bool) (*gittaginc.Tag, error) {
	iter, err := r.Tags()
	if err != nil {
		return nil, err
	}
	var highest *gittaginc.Tag = &gittaginc.Tag{}
	if err := iter.ForEach(func(ref *plumbing.Reference) error {
		if *verbose {
			fmt.Fprintf(out, "Ref: %s\n", ref.Name())
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

//go:embed usage.txt
var usageText string

func Usage() {
	out := flag.CommandLine.Output()

	var buf bytes.Buffer
	flag.VisitAll(func(f *flag.Flag) {
		fmt.Fprintf(&buf, "  -%s", f.Name)
		name, usage := flag.UnquoteUsage(f)
		if len(name) > 0 {
			fmt.Fprintf(&buf, " %s", name)
		}
		if len(usage) > 24 {
			fmt.Fprintf(&buf, "\n    \t")
		} else {
			fmt.Fprintf(&buf, "\t")
		}
		fmt.Fprintf(&buf, "%s", usage)
		if f.DefValue != "" {
			fmt.Fprintf(&buf, " (default %s)", f.DefValue)
		}
		fmt.Fprint(&buf, "\n")
	})

	t, err := template.New("usage").Parse(usageText)
	if err != nil {
		panic(err)
	}

	data := struct {
		ProgramName  string
		Flags        string
		PatchName    string
		ReleaseLines string
	}{
		ProgramName: os.Args[0],
		Flags:       buf.String(),
		PatchName:   "patch",
	}

	if *mode == gittaginc.ModeArraneous {
		data.PatchName = "release"
	} else {
		data.ReleaseLines = "* `release      => v0.0.1-test1 => v0.0.1-test2`\n* `release      => v0.0.1 => v0.0.1.1`\n"
	}

	if err := t.Execute(out, data); err != nil {
		panic(err)
	}
}

func printVersion() {
	fmt.Printf("git-tag-inc version %s\n", version)
	fmt.Printf("commit: %s\n", commit)
	fmt.Printf("branch: %s\n", branch)
	fmt.Printf("built: %s by %s\n", date, builtBy)
	fmt.Printf("repo: %s\n", repo)
	fmt.Printf("credits: Arran Ubels\n")
}
