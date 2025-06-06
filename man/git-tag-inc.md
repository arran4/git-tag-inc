# git-tag-inc(1) Manual

## Name
`git-tag-inc` - increment git version tags

## Synopsis
```
git-tag-inc [options] [command...]
```

## Description
`git-tag-inc` detects the highest semantic version tag in the repository and
creates the next tag. Commands control which part of the version is bumped and
which stage or environment counters are updated.

Supported stages include `alpha`, `beta` and `rc`. Environment counters `test`
and `uat` are also available.

## Commands
- `major`  – bump the major version (resets minor and patch)
- `minor`  – bump the minor version (resets patch)
- `patch`  – bump the patch version
- `release` – bump the release number. In `--mode arraneous` this behaves as
  `patch`
- `alpha`, `beta`, `rc` – start or bump the named pre-release stage
- `test`, `uat` – start or bump the named environment counter

## Options
- `--verbose` – print additional output
- `--version` – show build information
- `--dry` – display the tag that would be created
- `--ignore` – ignore uncommitted files (default)
- `--repeating` – allow new tags to repeat the last commit hash
- `--mode=MODE` – switch between `default` and `arraneous` naming

## Examples
Create a new test tag based on the highest existing version:

```bash
$ git-tag-inc test
```

Bump minor version and create an alpha pre-release:

```bash
$ git-tag-inc minor alpha
```

Bump patch and create a UAT tag:

```bash
$ git-tag-inc patch uat
```

Perform multiple increments at once:

```bash
$ git-tag-inc minor major test
```

## See also
`git-tag(1)`

## Author
Arran Ubels <arran@ubels.com.au>

