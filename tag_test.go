package gittaginc

import (
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func TestTag_LessThan(t1 *testing.T) {
	type args struct {
		other *Tag
	}
	tests := []struct {
		name     string
		tag      *Tag
		args     args
		lessThan bool
	}{
		{name: "Simple revision", tag: ParseTag("v1.2.3"), args: args{other: ParseTag("v1.2.4")}, lessThan: true},
		{name: "Simple revision reverse", tag: ParseTag("v1.2.4"), args: args{other: ParseTag("v1.2.3")}, lessThan: false},
		{name: "Simple major", tag: ParseTag("v1.2.3"), args: args{other: ParseTag("v4.2.3")}, lessThan: true},
		{name: "Simple major reverse", tag: ParseTag("v4.2.3"), args: args{other: ParseTag("v1.2.3")}, lessThan: false},
		{name: "Simple minor", tag: ParseTag("v1.3.3"), args: args{other: ParseTag("v1.4.3")}, lessThan: true},
		{name: "Simple minor reverse", tag: ParseTag("v1.4.3"), args: args{other: ParseTag("v1.3.3")}, lessThan: false},
		{name: "Simple equal", tag: ParseTag("v1.2.3"), args: args{other: ParseTag("v1.2.3")}, lessThan: false},
		{name: "Simple equal uat", tag: ParseTag("v1.2.3-uat1"), args: args{other: ParseTag("v1.2.3-uat1")}, lessThan: false},
		{name: "Simple equal test", tag: ParseTag("v1.2.3-test1"), args: args{other: ParseTag("v1.2.3-test1")}, lessThan: false},
		{name: "Simple uat ", tag: ParseTag("v1.2.3-uat1"), args: args{other: ParseTag("v1.2.3-uat2")}, lessThan: true},
		{name: "Simple uat reverse", tag: ParseTag("v1.2.3-uat2"), args: args{other: ParseTag("v1.2.3-uat1")}, lessThan: false},
		{name: "Simple test ", tag: ParseTag("v1.2.3-test1"), args: args{other: ParseTag("v1.2.3-test2")}, lessThan: true},
		{name: "Simple test reverse", tag: ParseTag("v1.2.3-test2"), args: args{other: ParseTag("v1.2.3-test1")}, lessThan: false},
		{name: "Simple equal uat test", tag: ParseTag("v1.2.3-test1"), args: args{other: ParseTag("v1.2.3-uat1")}, lessThan: true},
		{name: "Simple equal uat test reverse", tag: ParseTag("v1.2.3-uat1"), args: args{other: ParseTag("v1.2.3-test1")}, lessThan: false},
		{name: "Simple greater uat lower test", tag: ParseTag("v1.2.3-test1"), args: args{other: ParseTag("v1.2.3-uat2")}, lessThan: true},
		{name: "Simple greater uat lower test reverse", tag: ParseTag("v1.2.3-uat2"), args: args{other: ParseTag("v1.2.3-test1")}, lessThan: false},
		{name: "Simple greater test lower uat", tag: ParseTag("v1.2.3-test2"), args: args{other: ParseTag("v1.2.3-uat1")}, lessThan: false},
		{name: "Simple greater test lower uat reverse", tag: ParseTag("v1.2.3-uat1"), args: args{other: ParseTag("v1.2.3-test2")}, lessThan: true},
		{name: "Simple test same version", tag: ParseTag("v1.2.3-test1"), args: args{other: ParseTag("v1.2.3")}, lessThan: true},
		{name: "Simple test same version reverse", tag: ParseTag("v1.2.3"), args: args{other: ParseTag("v1.2.3-test1")}, lessThan: false},
		{name: "Simple uat same version", tag: ParseTag("v1.2.3-uat1"), args: args{other: ParseTag("v1.2.3")}, lessThan: true},
		{name: "Simple uat same version reverse", tag: ParseTag("v1.2.3"), args: args{other: ParseTag("v1.2.3-uat1")}, lessThan: false},
		{name: "Simple test previous version", tag: ParseTag("v1.2.3-test1"), args: args{other: ParseTag("v1.2.4")}, lessThan: true},
		{name: "Simple test previous version reverse", tag: ParseTag("v1.2.4"), args: args{other: ParseTag("v1.2.3-test1")}, lessThan: false},
		{name: "Simple uat previous version", tag: ParseTag("v1.2.3-uat1"), args: args{other: ParseTag("v1.2.4")}, lessThan: true},
		{name: "Simple uat previous version reverse", tag: ParseTag("v1.2.4"), args: args{other: ParseTag("v1.2.3-uat1")}, lessThan: false},
		{name: "Simple test next version", tag: ParseTag("v1.2.3-test1"), args: args{other: ParseTag("v1.2.2")}, lessThan: false},
		{name: "Simple test next version reverse", tag: ParseTag("v1.2.2"), args: args{other: ParseTag("v1.2.3-test1")}, lessThan: true},
		{name: "Simple uat next version", tag: ParseTag("v1.2.3-uat1"), args: args{other: ParseTag("v1.2.2")}, lessThan: false},
		{name: "Simple uat next version reverse", tag: ParseTag("v1.2.2"), args: args{other: ParseTag("v1.2.3-uat1")}, lessThan: true},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			if got := tt.tag.LessThan(tt.args.other); got != tt.lessThan {
				t1.Errorf("LessThan() = %v, lessThan %v", got, tt.lessThan)
			}
		})
	}
}

func TestTag_String(t1 *testing.T) {
	type fields struct {
		test    *int
		uat     *int
		release int
		major   int
		minor   int
		pad     int
	}
	pi := func(i int) *int {
		return &i
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{name: "Empty ", fields: fields{}, want: "v0.0.0"},
		{name: "Version serialized properly", fields: fields{test: nil, uat: nil, release: 1, major: 2, minor: 3, pad: 0}, want: "v2.3.1"},
		{name: "Version with uat and pad of 0", fields: fields{test: nil, uat: pi(1), release: 2, major: 1, minor: 3, pad: 0}, want: "v1.3.2-uat1"},
		{name: "Version with test and pad of 0", fields: fields{test: pi(1), uat: nil, release: 2, major: 1, minor: 3, pad: 0}, want: "v1.3.2-test1"},
		{name: "Version with uat and pad of 2", fields: fields{test: nil, uat: pi(1), release: 2, major: 1, minor: 3, pad: 2}, want: "v1.3.2-uat01"},
		{name: "Version with test and pad of 2", fields: fields{test: pi(1), uat: nil, release: 2, major: 1, minor: 3, pad: 2}, want: "v1.3.2-test01"},
		{name: "Version with uat and pad of 4", fields: fields{test: nil, uat: pi(1), release: 2, major: 1, minor: 3, pad: 4}, want: "v1.3.2-uat0001"},
		{name: "Version with test and pad of 4", fields: fields{test: pi(1), uat: nil, release: 2, major: 1, minor: 3, pad: 4}, want: "v1.3.2-test0001"},
		{name: "Version with test and uat of 2 - uat wins", fields: fields{test: pi(6), uat: pi(5), release: 2, major: 1, minor: 3, pad: 2}, want: "v1.3.2-uat05"},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tag{
				Test:    tt.fields.test,
				Uat:     tt.fields.uat,
				Release: tt.fields.release,
				Major:   tt.fields.major,
				Minor:   tt.fields.minor,
				Pad:     tt.fields.pad,
			}
			if got := t.String(); got != tt.want {
				t1.Errorf("String() = %v, lessThan %v", got, tt.want)
			}
		})
	}
}

func TestTag_Increment(t1 *testing.T) {
	type fields struct {
		test    *int
		uat     *int
		release int
		major   int
		minor   int
		pad     int
	}
	type args struct {
		major   bool
		minor   bool
		release bool
		uat     bool
		test    bool
	}
	pi := func(i int) *int {
		return &i
	}
	tests := []struct {
		name     string
		fields   *Tag
		expected *Tag
		args     args
	}{
		{name: "Simple version no flags", fields: &Tag{Test: nil, Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 2}, expected: &Tag{Test: nil, Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 2}, args: args{major: false, minor: false, release: false, uat: false, test: false}},
		{name: "Simple version with test - no flags", fields: &Tag{Test: pi(1), Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 2}, expected: &Tag{Test: pi(1), Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 2}, args: args{major: false, minor: false, release: false, uat: false, test: false}},
		{name: "Simple version uat flag - no flags", fields: &Tag{Test: nil, Uat: pi(1), Release: 1, Major: 1, Minor: 1, Pad: 2}, expected: &Tag{Test: nil, Uat: pi(1), Release: 1, Major: 1, Minor: 1, Pad: 2}, args: args{major: false, minor: false, release: false, uat: false, test: false}},
		{name: "Simple version with test - major ", fields: &Tag{Test: pi(1), Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 2}, expected: &Tag{Test: nil, Uat: nil, Release: 0, Major: 2, Minor: 0, Pad: 2}, args: args{major: true, minor: false, release: false, uat: false, test: false}},
		{name: "Simple version with test - minor ", fields: &Tag{Test: pi(1), Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 2}, expected: &Tag{Test: nil, Uat: nil, Release: 0, Major: 1, Minor: 2, Pad: 2}, args: args{major: false, minor: true, release: false, uat: false, test: false}},
		{name: "Simple version with test - release ",
			fields:   &Tag{Test: pi(1), Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 2},
			expected: &Tag{Test: nil, Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 2},
			args:     args{major: false, minor: false, release: true, uat: false, test: false}},
		{name: "Simple version with test - uat ", fields: &Tag{Test: pi(1), Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 2}, expected: &Tag{Test: nil, Uat: pi(1), Release: 1, Major: 1, Minor: 1, Pad: 2}, args: args{major: false, minor: false, release: false, uat: true, test: false}},
		{name: "Simple version with test - test ", fields: &Tag{Test: pi(1), Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 2}, expected: &Tag{Test: pi(2), Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 2}, args: args{major: false, minor: false, release: false, uat: false, test: true}},
		{name: "Simple version with uat - major ", fields: &Tag{Test: nil, Uat: pi(1), Release: 1, Major: 1, Minor: 1, Pad: 2}, expected: &Tag{Test: nil, Uat: nil, Release: 0, Major: 2, Minor: 0, Pad: 2}, args: args{major: true, minor: false, release: false, uat: false, test: false}},
		{name: "Simple version with uat - minor ", fields: &Tag{Test: nil, Uat: pi(1), Release: 1, Major: 1, Minor: 1, Pad: 2}, expected: &Tag{Test: nil, Uat: nil, Release: 0, Major: 1, Minor: 2, Pad: 2}, args: args{major: false, minor: true, release: false, uat: false, test: false}},
		{name: "Simple version with uat - release ", fields: &Tag{Test: nil, Uat: pi(1), Release: 1, Major: 1, Minor: 1, Pad: 2}, expected: &Tag{Test: nil, Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 2}, args: args{major: false, minor: false, release: true, uat: false, test: false}},
		{name: "Simple version with uat - uat ", fields: &Tag{Test: nil, Uat: pi(1), Release: 1, Major: 1, Minor: 1, Pad: 2}, expected: &Tag{Test: nil, Uat: pi(2), Release: 1, Major: 1, Minor: 1, Pad: 2}, args: args{major: false, minor: false, release: false, uat: true, test: false}},
		{name: "Simple version with uat - test ", fields: &Tag{Test: nil, Uat: pi(1), Release: 1, Major: 1, Minor: 1, Pad: 2}, expected: &Tag{Test: pi(2), Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 2}, args: args{major: false, minor: false, release: false, uat: false, test: true}},
		{name: "Simple version - major ", fields: &Tag{Test: nil, Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 0}, expected: &Tag{Test: nil, Uat: nil, Release: 0, Major: 2, Minor: 0, Pad: 0}, args: args{major: true, minor: false, release: false, uat: false, test: false}},
		{name: "Simple version - minor ", fields: &Tag{Test: nil, Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 0}, expected: &Tag{Test: nil, Uat: nil, Release: 0, Major: 1, Minor: 2, Pad: 0}, args: args{major: false, minor: true, release: false, uat: false, test: false}},
		{name: "Simple version - release ", fields: &Tag{Test: nil, Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 0}, expected: &Tag{Test: nil, Uat: nil, Release: 2, Major: 1, Minor: 1, Pad: 0}, args: args{major: false, minor: false, release: true, uat: false, test: false}},
		{name: "Simple version - uat ", fields: &Tag{Test: nil, Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 0}, expected: &Tag{Test: nil, Uat: pi(1), Release: 2, Major: 1, Minor: 1, Pad: 2}, args: args{major: false, minor: false, release: false, uat: true, test: false}},
		{name: "Simple version - test ", fields: &Tag{Test: nil, Uat: nil, Release: 1, Major: 1, Minor: 1, Pad: 0}, expected: &Tag{Test: pi(1), Uat: nil, Release: 2, Major: 1, Minor: 1, Pad: 2}, args: args{major: false, minor: false, release: false, uat: false, test: true}},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &Tag{
				Test:    tt.fields.Test,
				Uat:     tt.fields.Uat,
				Release: tt.fields.Release,
				Major:   tt.fields.Major,
				Minor:   tt.fields.Minor,
				Pad:     tt.fields.Pad,
			}
			t.Increment(tt.args.major, tt.args.minor, tt.args.release, tt.args.uat, tt.args.test)
			commands := FlagsToCommands(tt.args.major, tt.args.minor, tt.args.release, tt.args.uat, tt.args.test)
			assert.Equal(t1, tt.expected.String(), t.String(), "Converting %s to %s using flags %v but got %s", tt.fields.String(), tt.expected.String(), commands, t.String())
		})
	}
}

func FlagsToCommands(major bool, minor bool, release bool, uat bool, test bool) []string {
	result := []string{}
	if major {
		result = append(result, "major")
	}
	if minor {
		result = append(result, "minor")
	}
	if release {
		result = append(result, "release")
	}
	if uat {
		result = append(result, "uat")
	}
	if test {
		result = append(result, "test")
	}
	return result
}

func TestParseTag(t *testing.T) {
	type args struct {
		tag string
	}
	pi := func(i int) *int { return &i }
	tests := []struct {
		name string
		args args
		want *Tag
	}{
		{name: "Empty fails", args: args{tag: ""}, want: nil},
		{name: "Invalid fails", args: args{tag: "asdfasdfsad"}, want: nil},
		{name: "Close but not actually fails 1", args: args{tag: "v1..1.1"}, want: nil},
		{name: "Close but not actually fails 2", args: args{tag: "vv1.1.1"}, want: nil},
		{name: "Close but not actually fails 3", args: args{tag: "v1.2.3-utesttest45"}, want: nil},
		{name: "Basic v1.2.3", args: args{tag: "v1.2.3"}, want: &Tag{Test: nil, Uat: nil, Release: 3, Major: 1, Minor: 2, Pad: 0}},
		{name: "Basic uat v1.2.3-uat45", args: args{tag: "v1.2.3-uat45"}, want: &Tag{Test: nil, Uat: pi(45), Release: 3, Major: 1, Minor: 2, Pad: 2}},
		{name: "Basic uat large pad test v1.2.3-uat0045", args: args{tag: "v1.2.3-uat0045"}, want: &Tag{Test: nil, Uat: pi(45), Release: 3, Major: 1, Minor: 2, Pad: 4}},
		{name: "Basic test v1.2.3-test45", args: args{tag: "v1.2.3-test45"}, want: &Tag{Test: pi(45), Uat: nil, Release: 3, Major: 1, Minor: 2, Pad: 2}},
		{name: "Basic test large pad test v1.2.3-test0045", args: args{tag: "v1.2.3-test0045"}, want: &Tag{Test: pi(45), Uat: nil, Release: 3, Major: 1, Minor: 2, Pad: 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ParseTag(tt.args.tag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseTag() = %v, lessThan %v", got, tt.want)
			}
		})
	}
}

func TestWhole(t *testing.T) {
	tests := []struct {
		name     string
		inStr    string
		outStr   string
		commands []string
	}{
		{name: "test to release", inStr: "v0.0.1-test99", outStr: "v0.0.2-test01", commands: []string{"release", "test"}},
		{name: "test to major", inStr: "v0.0.1-test99", outStr: "v1.0.0-test01", commands: []string{"major", "test"}},
		{name: "test to minor", inStr: "v0.0.1-test99", outStr: "v0.1.0-test01", commands: []string{"minor", "test"}},
		{name: "test to uat", inStr: "v0.0.8-test01", outStr: "v0.0.8-uat01", commands: []string{"uat"}},
		{name: "uat to release", inStr: "v0.0.8-uat01", outStr: "v0.0.8", commands: []string{"release"}},
		{name: "test to release", inStr: "v0.0.8-test01", outStr: "v0.0.8", commands: []string{"release"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tag := ParseTag(tt.inStr)
			tag.Increment(CommandsToFlags(tt.commands))
			got := tag.String()
			if !reflect.DeepEqual(got, tt.outStr) {
				t.Errorf("'%s'.Increment(%v) => %v, Should be %v", tt.inStr, tt.commands, got, tt.outStr)
			}
		})
	}
}
