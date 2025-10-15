package gittaginc

import (
	"regexp"
	"strconv"
	"strings"
)

type CmdFlags struct {
	Major        bool
	MajorValue   *int
	Minor        bool
	MinorValue   *int
	Patch        bool
	PatchValue   *int
	Release      bool
	ReleaseValue *int
	Stage        string
	StageValue   *int
	StageDigits  int
	Env          string
	EnvValue     *int
	EnvDigits    int
	Valid        bool
}

func CommandsToFlags(args []string, mode string) CmdFlags {
	c := CmdFlags{Valid: true}
	re := regexp.MustCompile(`^([a-z]+)(\d+)?$`)
	for _, f := range args {
		lower := strings.ToLower(f)
		m := re.FindStringSubmatch(lower)
		if len(m) == 0 {
			c.Valid = false
			return c
		}
		name := m[1]
		var value *int
		if m[2] != "" {
			v, err := strconv.Atoi(m[2])
			if err != nil {
				c.Valid = false
				return c
			}
			value = pi(v)
		}
		switch name {
		case "major":
			c.Major = true
			if value != nil {
				c.MajorValue = value
			}
		case "minor":
			c.Minor = true
			if value != nil {
				c.MinorValue = value
			}
		case "patch":
			if mode == "arraneous" {
				c.Valid = false
				return c
			}
			c.Patch = true
			if value != nil {
				c.PatchValue = value
			}
		case "release":
			if mode == "arraneous" {
				c.Patch = true
				if value != nil {
					c.PatchValue = value
				}
			} else {
				c.Release = true
				if value != nil {
					c.ReleaseValue = value
				}
			}
		case "alpha", "beta", "rc":
			if c.Stage != "" {
				c.Valid = false
				return c
			}
			c.Stage = name
			if value != nil {
				c.StageValue = value
				digits := len(m[2])
				c.StageDigits = digits
			}
		case "test", "uat":
			if c.Env != "" {
				c.Valid = false
				return c
			}
			c.Env = name
			if value != nil {
				c.EnvValue = value
				digits := len(m[2])
				c.EnvDigits = digits
			}
		default:
			c.Valid = false
			return c
		}
	}
	return c
}

func pi(i int) *int { return &i }
