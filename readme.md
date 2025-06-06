# git-tag-inc

Increments the version number and tags it. (You will need to push)

# Usage

```
./git-tag-inc [major] [minor] [patch] [release] [alpha|beta|rc] [test|uat]
--version
```

Use `--version` to display build information and credits.

`--mode arraneous` switches to the legacy naming (patch becomes `release`).

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

## Combinations work:
* `patch test   => v0.0.1-test1 => v0.1.0-test1`

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
```

## Manual page

The repository contains a pre-built manual at `man/git-tag-inc.1`.
If you install `go-md2man` you can regenerate it from the Markdown source:

```bash
go-md2man -in=man/git-tag-inc.md -out=man/git-tag-inc.1
```
