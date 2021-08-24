# git-tag-inc

Increments the version number and tags it. (You will need to push)

# Usage

```
./git-tag-inc [major] [minor] [release] [test] [uat] 
```

## git-tag-inc then, one or more of:
* `major        => v0.0.1-test1 => v1.0.0`
* `minor        => v0.0.1-test1 => v0.1.0`
* `release      => v0.0.1-test1 => v0.0.2`
* `test         => v0.0.1-test1 => v0.0.1-test2`
* `test         => v0.0.1-uat1  => v0.0.1-test2`
* `uat          => v0.0.1-test3 => v0.0.1-uat3`
* `uat          => v0.0.1-uat1  => v0.0.1-uat2`  

## Combinations work:
* `release test => v0.0.1-test1 => v0.1.0-test1`  

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
$ git-tag-inc.exe release
Largest: v1.1.0-test1
Creating v1.1.1
```
