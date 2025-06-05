package gittaginc

import (
	"strings"
)

type CmdFlags struct {
	Major   bool
	Minor   bool
	Patch   bool
	Release bool
	Stage   string
	Env     string
	Valid   bool
}

func CommandsToFlags(args []string, mode string) CmdFlags {
	c := CmdFlags{Valid: true}
	for _, f := range args {
		switch strings.ToLower(f) {
		case "major":
			c.Major = true
		case "minor":
			c.Minor = true
		case "patch":
			if mode == "arraneous" {
				c.Valid = false
				return c
			}
			c.Patch = true
		case "release":
			if mode == "arraneous" {
				c.Patch = true
			} else {
				c.Release = true
			}
		case "alpha", "beta", "rc":
			if c.Stage != "" {
				c.Valid = false
				return c
			}
			c.Stage = strings.ToLower(f)
		case "test", "uat":
			if c.Env != "" {
				c.Valid = false
				return c
			}
			c.Env = strings.ToLower(f)
		default:
			c.Valid = false
			return c
		}
	}
	return c
}

func pi(i int) *int { return &i }
