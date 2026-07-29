package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/arran4/git-tag-inc"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func findHighestVersionTag(r *git.Repository) (*gittaginc.Tag, error) {
	iter, err := r.Tags()
	if err != nil {
		return nil, err
	}
	var highest *gittaginc.Tag
	if err := iter.ForEach(func(ref *plumbing.Reference) error {
		t := gittaginc.ParseTag(ref.Name().Short())
		if t == nil {
			return nil
		}
		t.Hash = ref.Hash().String()
		if highest == nil || highest.LessThan(t) {
			highest = t
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return highest, nil
}

func getTagFromStrOrPath(input string) *gittaginc.Tag {
	// check if it's a file path
	if strings.HasPrefix(input, "/") || strings.HasPrefix(input, "./") || strings.HasPrefix(input, "../") {
		absPath, err := filepath.Abs(input)
		if err == nil {
			r, err := git.PlainOpen(absPath)
			if err == nil {
				// Get the highest tag in this repo
				tag, _ := findHighestVersionTag(r)
				if tag != nil {
					return tag
				}
			}
		}
	}
	return gittaginc.ParseTag(input)
}

func main() {
	var input string
	if len(os.Args) > 1 {
		input = strings.Join(os.Args[1:], " ")
	} else {
		// Read from stdin
		stat, _ := os.Stdin.Stat()
		if (stat.Mode() & os.ModeCharDevice) == 0 {
			bytes, err := io.ReadAll(os.Stdin)
			if err == nil {
				input = string(bytes)
			}
		}
	}

	input = strings.TrimSpace(input)

	if input == "" {
		fmt.Fprintf(os.Stderr, "Usage: git-tag-cmp <tag1> <op> <tag2>\n")
		os.Exit(2)
	}

	// First try to parse with standard operators
	// We need to order the operators from longest to shortest in regex
	re := regexp.MustCompile(`(?i)^(.+?)(more-recent-than-or-equal|morerecentthanorequal|older-than-or-equal|olderthanorequal|less-than-or-equal|lessthanorequal|greater-than-or-equal|greaterthanorequal|more-recent-than|morerecentthan|older-than|olderthan|newer-than|newerthan|less-than|lessthan|greater-than|greaterthan|not-equal|notequal|equals|equal|<=|>=|<|>|==|!=|-lt|-le|-gt|-ge|-eq|-ne|lt|le|gt|ge|eq|ne)(.+?)$`)
	matches := re.FindStringSubmatch(input)

	var tag1Str, op, tag2Str string

	if len(matches) == 4 {
		tag1Str = strings.TrimSpace(matches[1])
		op = strings.ToLower(matches[2])
		tag2Str = strings.TrimSpace(matches[3])
	} else {
		// Try parsing space separated
		parts := strings.Fields(input)
		if len(parts) == 3 {
			tag1Str = parts[0]
			op = strings.ToLower(parts[1])
			tag2Str = parts[2]
		} else {
			fmt.Fprintf(os.Stderr, "Invalid format. Expected <tag1> <op> <tag2>\n")
			os.Exit(2)
		}
	}

	tag1 := getTagFromStrOrPath(tag1Str)
	if tag1 == nil {
		fmt.Fprintf(os.Stderr, "Invalid tag or path: %s\n", tag1Str)
		os.Exit(2)
	}

	tag2 := getTagFromStrOrPath(tag2Str)
	if tag2 == nil {
		fmt.Fprintf(os.Stderr, "Invalid tag or path: %s\n", tag2Str)
		os.Exit(2)
	}

	lessThan := tag1.LessThan(tag2)
	greaterThan := tag2.LessThan(tag1)
	equal := !lessThan && !greaterThan

	result := false
	switch op {
	case "<", "-lt", "lt", "less-than", "lessthan", "older-than", "olderthan":
		result = lessThan
	case "<=", "-le", "le", "less-than-or-equal", "lessthanorequal", "older-than-or-equal", "olderthanorequal":
		result = lessThan || equal
	case ">", "-gt", "gt", "greater-than", "greaterthan", "newer-than", "newerthan", "more-recent-than", "morerecentthan":
		result = greaterThan
	case ">=", "-ge", "ge", "greater-than-or-equal", "greaterthanorequal", "newer-than-or-equal", "newerthanorequal", "more-recent-than-or-equal", "morerecentthanorequal":
		result = greaterThan || equal
	case "==", "-eq", "eq", "equal", "equals":
		result = equal
	case "!=", "-ne", "ne", "not-equal", "notequal":
		result = !equal
	default:
		fmt.Fprintf(os.Stderr, "Unknown operator: %s\n", op)
		os.Exit(2)
	}

	fmt.Println(result)
	if result {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
