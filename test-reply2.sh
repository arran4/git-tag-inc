# Check if there's an existing push tag logic in prepare-release-tag
cat .github/workflows/ci.yml | grep -A 30 "Prepare release tag"
