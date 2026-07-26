# npm packaging

This mirrors the esbuild/swc/turbo distribution pattern: `secretcheck` is a
thin JS shim, and the actual binary ships in five tiny per-platform
packages listed as `optionalDependencies`. npm installs only the one that
matches the current OS/CPU.

```
packaging/npm/
  secretcheck/                 published as "secretcheck"
    package.json
    bin/secretcheck.js         shim: finds + execs the right platform binary
  platforms/
    darwin-arm64/package.json  published as "secretcheck-darwin-arm64"
    darwin-x64/package.json    published as "secretcheck-darwin-x64"
    linux-arm64/package.json   published as "secretcheck-linux-arm64"
    linux-x64/package.json     published as "secretcheck-linux-x64"
    win32-x64/package.json     published as "secretcheck-win32-x64"
```

Each `platforms/*/bin/` directory is empty in source control (just a
`.gitkeep`) — the release workflow copies the matching binary from
GoReleaser's `dist/` output into `platforms/<name>/bin/secretcheck[.exe]`
right before publishing that platform package. See
`.github/workflows/release-npm.yml`.

**Version numbers must stay in lockstep.** The main package's
`optionalDependencies` versions, and every platform package's own
`version`, need to match the release version exactly, or npm won't be able
to resolve them together. The release workflow rewrites all of these
automatically from the git tag — you should not need to bump them by hand.

## Manual test (after a release binary exists locally)

```bash
mkdir -p platforms/linux-x64/bin
cp ../../dist/secretcheck_linux_amd64/secretcheck platforms/linux-x64/bin/secretcheck
cd secretcheck && npm install ../platforms/linux-x64 --no-save
node bin/secretcheck.js scan --all
```
