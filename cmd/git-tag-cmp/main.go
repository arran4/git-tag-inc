package main

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/arran4/git-tag-inc"
)

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
	re := regexp.MustCompile(`^(.+?)(<=|>=|<|>|==|!=|-lt|-le|-gt|-ge|-eq|-ne)(.+?)$`)
	matches := re.FindStringSubmatch(input)

	var tag1Str, op, tag2Str string

	if len(matches) == 4 {
		tag1Str = strings.TrimSpace(matches[1])
		op = matches[2]
		tag2Str = strings.TrimSpace(matches[3])
	} else {
		// Try parsing space separated
		parts := strings.Fields(input)
		if len(parts) == 3 {
			tag1Str = parts[0]
			op = parts[1]
			tag2Str = parts[2]
		} else {
			fmt.Fprintf(os.Stderr, "Invalid format. Expected <tag1> <op> <tag2>\n")
			os.Exit(2)
		}
	}

	tag1 := gittaginc.ParseTag(tag1Str)
	if tag1 == nil {
		fmt.Fprintf(os.Stderr, "Invalid tag: %s\n", tag1Str)
		os.Exit(2)
	}

	tag2 := gittaginc.ParseTag(tag2Str)
	if tag2 == nil {
		fmt.Fprintf(os.Stderr, "Invalid tag: %s\n", tag2Str)
		os.Exit(2)
	}

	lessThan := tag1.LessThan(tag2)
	greaterThan := tag2.LessThan(tag1)
	equal := !lessThan && !greaterThan

	result := false
	switch op {
	case "<", "-lt":
		result = lessThan
	case "<=", "-le":
		result = lessThan || equal
	case ">", "-gt":
		result = greaterThan
	case ">=", "-ge":
		result = greaterThan || equal
	case "==", "-eq":
		result = equal
	case "!=", "-ne":
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
