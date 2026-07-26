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

# goTarget -> wheel platform tag. goTarget matches GoReleaser's dist
# folder naming: secretcheck_<goos>_<goarch>.
declare -A TARGETS=(
  [darwin_arm64]="macosx_11_0_arm64"
  [darwin_amd64]="macosx_10_13_x86_64"
  [linux_arm64]="manylinux_2_17_aarch64"
  [linux_amd64]="manylinux_2_17_x86_64"
  [windows_amd64]="win_amd64"
)

python -m pip install --quiet --upgrade build wheel

for goTarget in "${!TARGETS[@]}"; do
  wheelPlat="${TARGETS[$goTarget]}"

  rm -rf src/secretcheck/bin
  mkdir -p src/secretcheck/bin

  if [[ "$goTarget" == windows_* ]]; then
    cp "$DIST_DIR/secretcheck_${goTarget}/secretcheck.exe" src/secretcheck/bin/secretcheck.exe
  else
    cp "$DIST_DIR/secretcheck_${goTarget}/secretcheck" src/secretcheck/bin/secretcheck
    chmod +x src/secretcheck/bin/secretcheck
  fi

  echo "Building wheel for $goTarget -> $wheelPlat"
  SECRETCHECK_WHEEL_PLATFORM="$wheelPlat" python -m build --wheel --outdir "$OUT_DIR"
done

rm -rf src/secretcheck/bin

echo "Built wheels:"
ls -1 "$OUT_DIR"
