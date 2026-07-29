# Suggested Prompt for Parser Rewrite

Please address Issue #1 by refactoring the core tagging architecture.

Currently, `tag.go` relies on a rigid, hardcoded regular expression (`getParseTagRe`) to parse version strings, which inherently restricts the flexibility of appending environments (e.g. `-test`, `-uat`) and prerelease stages (`-alpha`, `-beta`).

Your task is to:
1. Replace the regex-based `ParseTag` logic with a completely custom, token-based or struct-driven parsing system.
2. The new parser must allow fully dynamic, configurable suffixes (environments and stages).
3. The system should support reordering, changing, or even chaining these suffixes.
4. Eliminate the `regexp.MustCompile` parsing entirely to prevent the regex complexity trap described in https://arran4.github.io/blog/post/2026/022-llm-expand-regex-test/.
5. Introduce a configuration mechanism (`LoadConfig`) to allow users to define their own custom valid environment suffixes and stages.
6. Ensure that the new parser remains backward-compatible with all existing tests in `tag_test.go`.

This is a sweeping domain change; please ensure all tests compile and pass, and prioritize writing clean, maintainable parsing logic.
