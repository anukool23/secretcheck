#!/usr/bin/env bash
# Builds all five platform wheels from the Go binaries GoReleaser already
# produced in ../../dist/. Every wheel is built on a single machine (no
# cross-compilation of native code needed, since we're only tagging and
# repackaging a prebuilt static binary) — see setup.py for how the
# platform tag is forced.
#
# Usage: ./build_wheels.sh
# Run from packaging/pypi/, after `goreleaser build --snapshot --clean`
# (or a real release) has populated ../../dist/.

set -euo pipefail

DIST_DIR="../../dist"
OUT_DIR="wheelhouse"

rm -rf "$OUT_DIR"
mkdir -p "$OUT_DIR"

# goTarget -> wheel platform tag. goTarget is split into goos/goarch and
# matched as a substring against whatever dist folder GoReleaser actually
# produced (rather than assuming an exact naming template) — GoReleaser has
# changed this across versions, e.g. adding a "_v1" suffix to amd64 builds
# by default.
declare -A TARGETS=(
  [darwin_arm64]="macosx_11_0_arm64"
  [darwin_amd64]="macosx_10_13_x86_64"
  [linux_arm64]="manylinux_2_17_aarch64"
  [linux_amd64]="manylinux_2_17_x86_64"
  [windows_amd64]="win_amd64"
)

# Finds the dist/ subdirectory whose name contains both goos and goarch,
# and echoes the path to the given binary inside it. Fails loudly with the
# contents of dist/ if nothing matches, so mismatches are easy to diagnose.
find_binary() {
  local goos="$1" goarch="$2" binname="$3"
  local match
  match=$(find "$DIST_DIR" -mindepth 1 -maxdepth 1 -type d -name "*${goos}*${goarch}*" | head -n1)

  if [[ -z "$match" ]]; then
    echo "ERROR: no dist/ directory matching goos=${goos} goarch=${goarch} under ${DIST_DIR}" >&2
    echo "Contents of ${DIST_DIR}:" >&2
    ls -1 "$DIST_DIR" >&2 || echo "  (dist/ does not exist — did GoReleaser run before this script?)" >&2
    exit 1
  fi

  if [[ ! -f "$match/$binname" ]]; then
    echo "ERROR: found ${match} but it has no ${binname}" >&2
    ls -1 "$match" >&2
    exit 1
  fi

  echo "$match/$binname"
}

python -m pip install --quiet --upgrade build wheel

for goTarget in "${!TARGETS[@]}"; do
  wheelPlat="${TARGETS[$goTarget]}"
  goos="${goTarget%_*}"
  goarch="${goTarget#*_}"

  rm -rf src/secretcheck/bin
  mkdir -p src/secretcheck/bin

  if [[ "$goTarget" == windows_* ]]; then
    src_binary=$(find_binary "$goos" "$goarch" "secretcheck.exe")
    cp "$src_binary" src/secretcheck/bin/secretcheck.exe
  else
    src_binary=$(find_binary "$goos" "$goarch" "secretcheck")
    cp "$src_binary" src/secretcheck/bin/secretcheck
    chmod +x src/secretcheck/bin/secretcheck
  fi

  echo "Building wheel for $goTarget -> $wheelPlat"
  SECRETCHECK_WHEEL_PLATFORM="$wheelPlat" python -m build --wheel --outdir "$OUT_DIR"
done

rm -rf src/secretcheck/bin

echo "Built wheels:"
ls -1 "$OUT_DIR"
