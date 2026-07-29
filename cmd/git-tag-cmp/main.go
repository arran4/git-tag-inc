// Copyright (c) 2025, Arran Ubels
// All rights reserved.
//
// This source code is licensed under the BSD-style license found in the
// LICENSE file in the root directory of this source tree.

package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/arran4/git-tag-inc"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: git-tag-cmp <tag1><op><tag2>\n")
		os.Exit(2)
	}

	arg := os.Args[1]

	// Regular expression to extract tag1, operator, and tag2
	re := regexp.MustCompile(`^(.+?)(<=|>=|<|>|==|!=)(.+?)$`)
	matches := re.FindStringSubmatch(arg)
	if len(matches) != 4 {
		fmt.Fprintf(os.Stderr, "Invalid format. Expected <tag1><op><tag2>\n")
		os.Exit(2)
	}

	tag1Str := strings.TrimSpace(matches[1])
	op := matches[2]
	tag2Str := strings.TrimSpace(matches[3])

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
	case "<":
		result = lessThan
	case "<=":
		result = lessThan || equal
	case ">":
		result = greaterThan
	case ">=":
		result = greaterThan || equal
	case "==":
		result = equal
	case "!=":
		result = !equal
	}

	fmt.Println(result)
	if result {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}
