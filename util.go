package gittaginc

import (
	"strings"
)

func CommandsToFlags(args []string) (bool, bool, bool, bool, bool) {
	test := false
	uat := false
	release := false
	major := false
	minor := false
	d := 0
	for _, f := range args {
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
			return false, false, false, false, false
		}
		d++
	}
	return major, minor, release, uat, test
}

func pi(i int) *int { return &i }
