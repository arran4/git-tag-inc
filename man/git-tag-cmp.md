# git-tag-cmp(1) Manual

## Name
`git-tag-cmp` - compare two git version tags or repositories

## Synopsis
```
git-tag-cmp <tag1|repo1> <op> <tag2|repo2>
git-tag-cmp "<tag1|repo1> <op> <tag2|repo2>"
```

## Description
`git-tag-cmp` takes two semantic version tags (or paths to git repositories) and a comparison operator and returns an exit code indicating whether the statement is true (0) or false (1), as well as printing "true" or "false" to stdout.

If a path to a git repository (starting with `/`, `./`, or `../`) is provided instead of a tag, `git-tag-cmp` will open the repository and use its highest version tag for the comparison.

This utility parses tags strictly using the same logic and ranking as `git-tag-inc(1)`.

## Operands
- `<tag1|repo1>`: A semantic version tag string (e.g. `v1.2.3-alpha1`) or a path to a git repository.
- `<op>`: The comparison operator. Supports standard mathematical symbols, bash-style operators, and word equivalents. See **Operators** below.
- `<tag2|repo2>`: A semantic version tag string (e.g. `v1.2.3-beta1`) or a path to a git repository.

## Operators
- `<` | `-lt` | `lt` | `less-than` | `lessthan` | `older-than` | `olderthan`
- `<=` | `-le` | `le` | `less-than-or-equal` | `lessthanorequal` | `older-than-or-equal` | `olderthanorequal`
- `>` | `-gt` | `gt` | `greater-than` | `greaterthan` | `newer-than` | `newerthan` | `more-recent-than` | `morerecentthan`
- `>=` | `-ge` | `ge` | `greater-than-or-equal` | `greaterthanorequal` | `newer-than-or-equal` | `newerthanorequal` | `more-recent-than-or-equal` | `morerecentthanorequal`
- `==` | `-eq` | `eq` | `equal` | `equals`
- `!=` | `-ne` | `ne` | `not-equal` | `notequal`

## Input Forms
The tool supports multiple input forms for convenience:
1. **Multiple arguments:** `git-tag-cmp v1.0.0 lt v2.0.0`
2. **Single string:** `git-tag-cmp "v1.0.0<=v2.0.0"`
3. **Standard input (stdin):** `echo "v1.0.0 < v2.0.0" | git-tag-cmp`

## Examples
Compare two specific tags:

```bash
$ git-tag-cmp v1.2.3 <= v2.0.0
true
```

Use bash-style operators:

```bash
$ git-tag-cmp v2.0.0 -gt v1.0.0
true
```

Compare two repositories to see which is newer:

```bash
$ git-tag-cmp ../project1-old gt ../project1-new
false
```

Evaluate dynamically generated tags from a script:

```bash
if git-tag-cmp "$TAG1" newer-than "$TAG2"; then
  echo "TAG1 is newer"
fi
```

## See also
`git-tag-inc(1)`, `git-tag(1)`

## Author
Arran Ubels <arran@ubels.com.au>
