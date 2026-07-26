#!/usr/bin/env bash
# Cuts a release: creates an annotated git tag and pushes it, which is all
# that's needed to trigger release-go.yml, release-npm.yml, and
# release-pypi.yml. Those three workflows derive the version from the tag
# itself and rewrite packaging/npm/*/package.json + packaging/pypi/pyproject.toml
# on the CI runner at publish time — nothing in this repo needs its version
# hand-edited before tagging.
#
# Usage:
#   ./scripts/release.sh 1.0.5
#   ./scripts/release.sh 1.0.5 --dry-run
#
# Requires: git, a clean working tree, and to be up to date with origin/main.

set -euo pipefail

VERSION="${1:-}"
DRY_RUN=false
if [[ "${2:-}" == "--dry-run" ]]; then
  DRY_RUN=true
fi

if [[ -z "$VERSION" ]]; then
  echo "Usage: $0 <version> [--dry-run]" >&2
  echo "Example: $0 1.0.5" >&2
  exit 1
fi

# Strip a leading 'v' if the caller included one, so both "1.0.5" and
# "v1.0.5" work the same way.
VERSION="${VERSION#v}"

if ! [[ "$VERSION" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.]+)?$ ]]; then
  echo "error: '$VERSION' doesn't look like semver (expected e.g. 1.0.5)" >&2
  exit 1
fi

TAG="v$VERSION"

if [[ -n "$(git status --porcelain)" ]]; then
  echo "error: working tree is not clean. Commit or stash your changes first:" >&2
  git status --short >&2
  exit 1
fi

BRANCH="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$BRANCH" != "main" ]]; then
  echo "warning: you're on branch '$BRANCH', not 'main'." >&2
  read -r -p "Continue anyway? (y/N) " reply
  [[ "$reply" =~ ^[Yy]$ ]] || exit 1
fi

echo ">> Fetching latest from origin..."
git fetch origin "$BRANCH" --quiet

LOCAL_SHA="$(git rev-parse HEAD)"
REMOTE_SHA="$(git rev-parse "origin/$BRANCH")"
if [[ "$LOCAL_SHA" != "$REMOTE_SHA" ]]; then
  echo "error: local $BRANCH ($LOCAL_SHA) differs from origin/$BRANCH ($REMOTE_SHA)." >&2
  echo "Push or pull first, so the tag points at a commit that's actually on GitHub." >&2
  exit 1
fi

if git rev-parse "$TAG" >/dev/null 2>&1; then
  echo "error: tag $TAG already exists locally." >&2
  exit 1
fi
if git ls-remote --tags origin "refs/tags/$TAG" | grep -q "$TAG"; then
  echo "error: tag $TAG already exists on origin." >&2
  exit 1
fi

echo ">> Will create and push tag: $TAG (at $LOCAL_SHA, branch $BRANCH)"
if $DRY_RUN; then
  echo ">> --dry-run: not actually tagging or pushing."
  exit 0
fi

git tag -a "$TAG" -m "Release $TAG"
git push origin "$TAG"

echo
echo "Pushed $TAG. Three workflows should now be running in parallel:"
echo "  https://github.com/anukool23/secretcheck/actions"
echo
echo "  - release-go.yml    builds + publishes the GitHub Release (go install works once this succeeds)"
echo "  - release-npm.yml   publishes the npm packages"
echo "  - release-pypi.yml  publishes the PyPI wheels"
