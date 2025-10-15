package gittaginc

import (
	"reflect"
	"testing"
)

func TestParseTag(t *testing.T) {
	tests := []struct {
		tag  string
		want *Tag
	}{
		{"", nil},
		{"garbage", nil},
		{"v1..1", nil},
		{"vv1.1.1", nil},
		{"v1.2.3-uattest", nil},
		{"v1.2", nil},
		{"v1.0.0-", nil},
		{"v1.0.0-test", nil},
		{"v1.0.0-beta-uat1", nil},
		{"v1.0.0-alpha01uat01", nil},
		{"v1.0.0-beta01-foo01", nil},
		{"v1.0.0-unknown1", nil},
		{"v1.2.3", &Tag{Major: 1, Minor: 2, Patch: 3}},
		{"v1.2.3-test45", &Tag{Major: 1, Minor: 2, Patch: 3, Test: pi(45), Pad: 2}},
		{"v1.2.3-uat0045", &Tag{Major: 1, Minor: 2, Patch: 3, Uat: pi(45), Pad: 4}},
		{"v1.2.3-alpha1", &Tag{Major: 1, Minor: 2, Patch: 3, StageName: "alpha", Stage: pi(1), StagePad: 0}},
		{"v1.2.3-beta02-test03", &Tag{Major: 1, Minor: 2, Patch: 3, StageName: "beta", Stage: pi(2), StagePad: 2, Test: pi(3), Pad: 2}},
		{"v1.0.0-rc01-test02", &Tag{Major: 1, Minor: 0, Patch: 0, StageName: "rc", Stage: pi(1), StagePad: 2, Test: pi(2), Pad: 2}},
		{"v1.0.0-beta007-uat012", &Tag{Major: 1, Minor: 0, Patch: 0, StageName: "beta", Stage: pi(7), StagePad: 3, Uat: pi(12), Pad: 3}},
		{"v1.0.0-beta1-test2.3", &Tag{Major: 1, Minor: 0, Patch: 0, StageName: "beta", Stage: pi(1), StagePad: 0, Test: pi(2), Pad: 0, Release: pi(3)}},
	}
	for _, tt := range tests {
		got := ParseTag(tt.tag)
		if !reflect.DeepEqual(got, tt.want) {
			if got == nil || tt.want == nil {
				t.Errorf("ParseTag(%s) = %#v, want %#v", tt.tag, got, tt.want)
			} else if got.String() != tt.want.String() {
				t.Errorf("ParseTag(%s) = %s, want %s", tt.tag, got.String(), tt.want.String())
			}
		}
	}
}

func TestString(t *testing.T) {
	cases := []struct {
		tag  *Tag
		want string
	}{
		{&Tag{Major: 0, Minor: 0, Patch: 0}, "v0.0.0"},
		{&Tag{Major: 1, Minor: 2, Patch: 3}, "v1.2.3"},
		{&Tag{Major: 1, Minor: 0, Patch: 0, StageName: "rc", Stage: pi(1), StagePad: 2}, "v1.0.0-rc01"},
		{&Tag{Major: 0, Minor: 1, Patch: 2, Test: pi(3), Pad: 2}, "v0.1.2-test03"},
		{&Tag{Major: 2, Minor: 3, Patch: 4, StageName: "beta", Stage: pi(2), StagePad: 2, Uat: pi(1), Pad: 2}, "v2.3.4-beta02-uat01"},
		{&Tag{Major: 5, Minor: 6, Patch: 7, StageName: "beta", Stage: pi(2), StagePad: 3, Test: pi(10), Pad: 3}, "v5.6.7-beta002-test010"},
		{&Tag{Major: 1, Minor: 0, Patch: 1, StageName: "alpha", Stage: pi(1), StagePad: 2, Test: pi(1), Pad: 2, Release: pi(2)}, "v1.0.1-alpha01-test01.2"},
	}
	for _, tt := range cases {
		if got := tt.tag.String(); got != tt.want {
			t.Errorf("%v got %s want %s", tt.tag, got, tt.want)
		}
	}
}

func TestIncrement(t *testing.T) {
	tests := []struct {
		name     string
		start    string
		cmds     []string
		expected string
	}{
		{"patch", "v0.0.1", []string{"patch"}, "v0.0.2"},
		{"stage then env", "v0.0.2", []string{"alpha", "test"}, "v0.0.3-alpha01-test01"},
		{"major reset", "v0.0.3-alpha01-test01", []string{"major"}, "v1.0.0"},
		{"minor", "v1.0.0", []string{"minor"}, "v1.1.0"},
		{"env bump", "v1.1.0-test01", []string{"test"}, "v1.1.0-test02"},
		{"stage bump", "v1.1.0-alpha01", []string{"alpha"}, "v1.1.0-alpha02"},
		{"env switch", "v1.1.0-test02", []string{"uat"}, "v1.1.0-uat02"},
		{"patch clears stage", "v1.1.0-beta01", []string{"patch"}, "v1.1.0"},
		{"patch drops env", "v1.1.0-test01", []string{"patch"}, "v1.1.0"},
		{"patch with env", "v1.1.0-uat01", []string{"patch", "test"}, "v1.1.1-test01"},
		{"stage resets env", "v1.1.0-test01", []string{"alpha"}, "v1.1.1-alpha01"},
		{"env resets stage", "v1.1.1-alpha01", []string{"uat"}, "v1.1.1-alpha01-uat01"},
		{"change stage", "v1.1.1-alpha02", []string{"beta"}, "v1.1.2-beta01"},
		{"release bump", "v1.1.1-alpha01-test01", []string{"release"}, "v1.1.1-alpha01-test01.1"},
		{"release again", "v1.1.1-alpha01-test01.1", []string{"release"}, "v1.1.1-alpha01-test01.2"},
		{"explicit env", "v1.1.1-test02", []string{"test5"}, "v1.1.1-test05"},
		{"new env single digit defaults", "v1.1.0", []string{"test2"}, "v1.1.1-test02"},
		{"env retains existing width", "v1.1.0-test004", []string{"test5"}, "v1.1.0-test005"},
		{"env without padding stays unpadded", "v1.1.0-test3", []string{"test"}, "v1.1.0-test4"},
		{"switch env single digit defaults", "v1.1.0-test02", []string{"uat2"}, "v1.1.0-uat02"},
		{"explicit stage new base", "v1.1.1-rc03", []string{"patch", "rc2"}, "v1.1.2-rc02"},
		{"stage retains existing width", "v1.1.0-rc004", []string{"rc5"}, "v1.1.0-rc005"},
		{"new stage single digit defaults", "v1.1.0", []string{"rc2"}, "v1.1.0-rc02"},
		{"explicit release", "v1.1.1-test01.3", []string{"release5"}, "v1.1.1-test01.5"},
		{"explicit major", "v1.1.1", []string{"major5"}, "v5.0.0"},
		{"explicit minor", "v5.0.0", []string{"minor7"}, "v5.7.0"},
		{"explicit patch", "v5.7.0", []string{"patch9"}, "v5.7.9"},
	}
	for _, tt := range tests {
		tag := ParseTag(tt.start)
		flags := CommandsToFlags(tt.cmds, "default")
		if err := tag.Increment(flags, false, false); err != nil {
			t.Fatalf("unexpected error incrementing %s with %v: %v", tt.start, tt.cmds, err)
		}
		if got := tag.String(); got != tt.expected {
			t.Errorf("%s Increment(%v) got %s want %s", tt.start, tt.cmds, got, tt.expected)
		}
	}
}

func TestLessThan(t *testing.T) {
	cases := []struct {
		a    string
		b    string
		want bool
	}{
		{"v1.0.0-alpha1-test1", "v1.0.0-beta1-test1", true},
		{"v1.0.0-test1", "v1.0.0-uat1", true},
		{"v1.0.1", "v1.0.0", false},
		{"v1.1.0", "v2.0.0", true},
		{"v1.0.0-rc1", "v1.0.0", true},
		{"v1.0.0-test1", "v1.0.0", true},
		{"v1.0.0-uat1", "v1.0.0-test1", false},
		{"v1.0.0-beta1-test1", "v1.0.0-alpha1-test2", false},
		{"v1.0.0-test1.1", "v1.0.0-test1.2", true},
	}
	for _, tt := range cases {
		l := ParseTag(tt.a)
		r := ParseTag(tt.b)
		if got := l.LessThan(r); got != tt.want {
			t.Errorf("%s < %s got %v want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestIncrementSequence(t *testing.T) {
	tag := ParseTag("v0.0.1")
	seq := [][]string{
		{"patch"},
		{"release"},
		{"release"},
		{"alpha"},
		{"release"},
	}
	wants := []string{
		"v0.0.2",
		"v0.0.2.1",
		"v0.0.2.2",
		"v0.0.3-alpha01",
		"v0.0.3-alpha01.1",
	}
	for i, cmds := range seq {
		f := CommandsToFlags(cmds, "default")
		if err := tag.Increment(f, false, false); err != nil {
			t.Fatalf("step %d unexpected error: %v", i, err)
		}
		if got := tag.String(); got != wants[i] {
			t.Fatalf("step %d got %s want %s", i, got, wants[i])
		}
	}
}

func TestIncrementBackwardsProtection(t *testing.T) {
	t.Run("env counters", func(t *testing.T) {
		original := ParseTag("v1.0.0-test3")
		backwards := CommandsToFlags([]string{"test2"}, "default")
		if err := original.Increment(backwards, false, false); err == nil {
			t.Fatalf("expected error when decrementing without flags")
		}
		if got := original.String(); got != "v1.0.0-test3" {
			t.Fatalf("tag mutated on error got %s", got)
		}

		allow := ParseTag("v1.0.0-test3")
		if err := allow.Increment(backwards, true, false); err != nil {
			t.Fatalf("allow backwards returned error: %v", err)
		}
		if got := allow.String(); got != "v1.0.0-test2" {
			t.Fatalf("allow backwards produced %s", got)
		}

		skip := ParseTag("v1.0.0-test3")
		if err := skip.Increment(backwards, false, true); err != nil {
			t.Fatalf("skip forwards returned error: %v", err)
		}
		if got := skip.String(); got != "v1.0.1-test02" {
			t.Fatalf("skip forwards produced %s", got)
		}

		withRelease := ParseTag("v1.0.0-test3.1")
		if err := withRelease.Increment(backwards, false, true); err != nil {
			t.Fatalf("skip forwards with release returned error: %v", err)
		}
		if got := withRelease.String(); got != "v1.0.1-test02" {
			t.Fatalf("skip forwards with release produced %s", got)
		}
	})

	t.Run("stages", func(t *testing.T) {
		stage := ParseTag("v1.0.0-rc3")
		stageFlags := CommandsToFlags([]string{"rc2"}, "default")
		if err := stage.Increment(stageFlags, false, false); err == nil {
			t.Fatalf("expected error when decrementing stage without flags")
		}
		allowStage := ParseTag("v1.0.0-rc3")
		if err := allowStage.Increment(stageFlags, true, false); err != nil {
			t.Fatalf("allow backwards stage returned error: %v", err)
		}
		if got := allowStage.String(); got != "v1.0.0-rc2" {
			t.Fatalf("allow backwards stage produced %s", got)
		}
		skipStage := ParseTag("v1.0.0-rc3")
		if err := skipStage.Increment(stageFlags, false, true); err != nil {
			t.Fatalf("skip forwards stage returned error: %v", err)
		}
		if got := skipStage.String(); got != "v1.0.1-rc02" {
			t.Fatalf("skip forwards stage produced %s", got)
		}
	})

	t.Run("core version numbers", func(t *testing.T) {
		patchFlags := CommandsToFlags([]string{"patch3"}, "default")
		patch := ParseTag("v2.3.4")
		if err := patch.Increment(patchFlags, false, false); err == nil {
			t.Fatalf("expected patch decrement error")
		}
		if err := patch.Increment(patchFlags, true, false); err != nil {
			t.Fatalf("allow patch decrement failed: %v", err)
		}
		if got := patch.String(); got != "v2.3.3" {
			t.Fatalf("allow patch decrement produced %s", got)
		}

		minorFlags := CommandsToFlags([]string{"minor1"}, "default")
		minor := ParseTag("v2.3.4")
		if err := minor.Increment(minorFlags, false, false); err == nil {
			t.Fatalf("expected minor decrement error")
		}
		if err := minor.Increment(minorFlags, true, false); err != nil {
			t.Fatalf("allow minor decrement failed: %v", err)
		}
		if got := minor.String(); got != "v2.1.0" {
			t.Fatalf("allow minor decrement produced %s", got)
		}

		majorFlags := CommandsToFlags([]string{"major1"}, "default")
		major := ParseTag("v3.0.0")
		if err := major.Increment(majorFlags, false, false); err == nil {
			t.Fatalf("expected major decrement error")
		}
		if err := major.Increment(majorFlags, true, false); err != nil {
			t.Fatalf("allow major decrement failed: %v", err)
		}
		if got := major.String(); got != "v1.0.0" {
			t.Fatalf("allow major decrement produced %s", got)
		}

		releaseFlags := CommandsToFlags([]string{"release2"}, "default")
		release := ParseTag("v1.2.3-test3.5")
		if err := release.Increment(releaseFlags, false, false); err == nil {
			t.Fatalf("expected release decrement error")
		}
		if err := release.Increment(releaseFlags, true, false); err != nil {
			t.Fatalf("allow release decrement failed: %v", err)
		}
		if got := release.String(); got != "v1.2.3-test3.2" {
			t.Fatalf("allow release decrement produced %s", got)
		}

		skipRelease := ParseTag("v1.2.3-test3.5")
		if err := skipRelease.Increment(releaseFlags, false, true); err != nil {
			t.Fatalf("skip release decrement returned error: %v", err)
		}
		if got := skipRelease.String(); got != "v1.2.4.2" {
			t.Fatalf("skip release decrement produced %s", got)
		}

		skipPatch := ParseTag("v2.3.4")
		if err := skipPatch.Increment(patchFlags, false, true); err == nil {
			t.Fatalf("skip forwards should not allow patch decrement when patch provided")
		}
	})
}

func TestCommandsToFlags(t *testing.T) {
	good := CommandsToFlags([]string{"major", "patch", "release", "test"}, "default")
	if !good.Major || !good.Patch || !good.Release || good.Env != "test" || !good.Valid {
		t.Fatalf("unexpected flags %#v", good)
	}
	combo := CommandsToFlags([]string{"patch", "release"}, "default")
	if !combo.Patch || !combo.Release || !combo.Valid {
		t.Fatalf("combo parsing failed %#v", combo)
	}
	numbers := CommandsToFlags([]string{"test12", "rc03", "patch42", "major10", "minor5", "release7"}, "default")
	if numbers.EnvValue == nil || *numbers.EnvValue != 12 || numbers.EnvDigits != 2 {
		t.Fatalf("expected env numeric parsing %#v", numbers)
	}
	if numbers.StageValue == nil || *numbers.StageValue != 3 || numbers.StageDigits != 2 {
		t.Fatalf("expected stage numeric parsing %#v", numbers)
	}
	envDigits := CommandsToFlags([]string{"test2"}, "default")
	if envDigits.EnvValue == nil || *envDigits.EnvValue != 2 || envDigits.EnvDigits != 1 {
		t.Fatalf("expected env digits default padding %#v", envDigits)
	}
	stageDigits := CommandsToFlags([]string{"rc2"}, "default")
	if stageDigits.StageValue == nil || *stageDigits.StageValue != 2 || stageDigits.StageDigits != 1 {
		t.Fatalf("expected stage digits default padding %#v", stageDigits)
	}
	if numbers.PatchValue == nil || *numbers.PatchValue != 42 {
		t.Fatalf("expected patch numeric parsing %#v", numbers)
	}
	if numbers.MajorValue == nil || *numbers.MajorValue != 10 || numbers.MinorValue == nil || *numbers.MinorValue != 5 {
		t.Fatalf("expected major/minor numeric parsing %#v", numbers)
	}
	if numbers.ReleaseValue == nil || *numbers.ReleaseValue != 7 {
		t.Fatalf("expected release numeric parsing %#v", numbers)
	}
	arr := CommandsToFlags([]string{"release", "uat"}, "arraneous")
	if !arr.Patch || arr.Env != "uat" || !arr.Valid {
		t.Fatalf("arraneous parsing failed %#v", arr)
	}
	dup := CommandsToFlags([]string{"test", "uat"}, "default")
	if dup.Valid {
		t.Fatalf("expected invalid duplicate env")
	}
	dupStage := CommandsToFlags([]string{"alpha", "beta"}, "default")
	if dupStage.Valid {
		t.Fatalf("expected invalid duplicate stage")
	}
	relOnly := CommandsToFlags([]string{"release"}, "default")
	if !relOnly.Release || !relOnly.Valid {
		t.Fatalf("release only failed %#v", relOnly)
	}
	wrong2 := CommandsToFlags([]string{"patch"}, "arraneous")
	if wrong2.Valid {
		t.Fatalf("expected invalid patch in arraneous")
	}
}
