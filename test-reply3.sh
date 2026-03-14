# The user's comment is: "@jules If we tag elsewhere couldn't we pull in the tags?"
# The user is asking if we can just "pull in the tags" instead of doing a local `git tag`.
# Wait, if we tag elsewhere... do they mean if we push the tag in the `prepare-release-tag` job, then we could just pull the tags here using `actions/checkout` or `git fetch --tags`?
# Yes! But `prepare-release-tag` doesn't push the tag. It just calculates what the tag should be:
cat .github/workflows/ci.yml | grep -A 5 "next_tag=\$(go run ./cmd/git-tag-inc -print-version-only \$level \$suffix)"
