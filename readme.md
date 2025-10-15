# git-tag-inc

Increments the version number and tags it. (You will need to push)

# Usage

```
./git-tag-inc [--allow-backwards] [--skip-forwards] [major[<n>]] [minor[<n>]] [patch[<n>]] [release[<n>]] [alpha|beta|rc[<n>]] [test|uat[<n>]]
--version [--print-version-only]
```

Use `--version` to display build information and credits.
Use `--print-version-only` to output the next version without tagging.

`--mode arraneous` switches to the legacy naming (patch becomes `release`).

Numeric suffixes can be added to any command to set a specific counter. For example,
`test5` produces `-test5`, `rc02` produces `-rc02` and `major3` moves directly to
`v3.0.0`. When a numeric suffix would decrease a counter compared to the previous tag
the command fails unless either `--allow-backwards` is provided or `--skip-forwards`
is used. `--allow-backwards` applies the requested number directly, while
`--skip-forwards` automatically bumps the patch component first so the resulting tag
still increases. For instance, `git-tag-inc --skip-forwards test2` upgrades
`v1.0.0-test3` to `v1.0.1-test2`.

## git-tag-inc then, one or more of:
* `major        => v0.0.1-test1 => v1.0.0`
* `minor        => v0.0.1-test1 => v0.1.0`
* `patch        => v0.0.1-test1 => v0.0.2`
* `release      => v0.0.1-test1 => v0.0.1-test2`
* `release      => v0.0.1 => v0.0.1.1`
* `test         => v0.0.1-test1 => v0.0.1-test2`
* `uat          => v0.0.1-uat1  => v0.0.1-uat2`
* `alpha        => v0.0.1-alpha1 => v0.0.1-alpha2`
* `beta         => v0.0.1-beta1  => v0.0.1-beta2`
* `rc           => v0.0.1-rc1    => v0.0.1-rc2`
* `rc5          => v0.0.1-rc1    => v0.0.1-rc5`
* `major4       => v0.0.1        => v4.0.0`

## Combinations work:
* `patch test   => v0.0.1-test1 => v0.1.0-test1`
* `patch rc2    => v0.1.0-rc4  => v0.1.1-rc2`

## Preventing backwards moves:
* `test1` (when the last tag was `test3`) errors unless `--allow-backwards` is supplied.
* `--skip-forwards test1` turns the same command into `vX.Y.(Z+1)-test1` automatically.

```bash
$ git-tag-inc --allow-backwards test2
# v1.0.0-test3 -> v1.0.0-test2
$ git-tag-inc --skip-forwards test2
# v1.0.0-test3 -> v1.0.1-test2
$ git-tag-inc --allow-backwards major1
# v3.0.0 -> v1.0.0
$ git-tag-inc --skip-forwards release2
# v1.2.3-test3.5 -> v1.2.4.2
```

## Duplications don't:
* `test test    => v0.0.1-test1 => v0.0.1-test2`

# Install

You can use the packages provided. Put them in your `$PATH` or `%path%` depending on OS. You can also use:
```
$ git clone github.com/arran4/git-tag-inc
$ cd git-tag-inc
$ go install .
```

# Example:

```
$ git-tag-inc.exe test
Largest: v0.0.1-test1
Creating v0.0.1-test2

$ git-tag-inc.exe uat
Largest: v0.0.1-test2
Creating v0.0.1-uat2

$ git-tag-inc.exe uat
Largest: v0.0.1-uat2
Creating v0.0.1-uat3

$ git-tag-inc.exe test
Largest: v0.0.1-uat3
Creating v0.0.1-test4

$ git-tag-inc.exe minor
Largest: v0.0.1-test4
Creating v0.1.0

$ git-tag-inc.exe minor test
Largest: v0.1.0
Creating v0.2.0-test01

$ git-tag-inc.exe minor major test
Largest: v0.2.0-test1
Creating v1.1.0-test01

$ git-tag-inc.exe patch
Largest: v1.1.0-test1
Creating v1.1.1

$ git-tag-inc.exe release
Largest: v1.1.1-test1
Creating v1.1.1-test1.1

$ git-tag-inc.exe --skip-forwards test2
Largest: v1.1.1-test3
Creating v1.1.2-test2
```

## Manual page

The repository contains a pre-built manual at `man/git-tag-inc.1`.
If you install `go-md2man` you can regenerate it from the Markdown source:

```bash
go-md2man -in=man/git-tag-inc.md -out=man/git-tag-inc.1
```
