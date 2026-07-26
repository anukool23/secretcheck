# secretcheck

A git pre-commit hook that scans your staged files for secrets — API keys,
tokens, private keys, credentials — and blocks the commit if it finds any.
If it's a false positive (or you really need to commit anyway), you can
override it on the spot.

Written in Go with zero third-party dependencies, distributed as a single
binary. The same binary powers native installs in Go, npm, and PyPI
projects, so one tool works across polyglot repos.

**Full docs:** [docs.anukool.me/secretcheck](https://docs.anukool.me/secretcheck)

## Install

Pick whichever matches how your project already installs its tools — they
all install the exact same binary and behave identically.

**Go**
```bash
go install github.com/anukool23/secretcheck/cmd/secretcheck@latest
```

**npm / Node projects**
```bash
npm install --save-dev secretcheck
```

**Python projects**
```bash
pip install secretcheck
```

**Direct download** — grab a binary from the [Releases page](https://github.com/anukool23/secretcheck/releases) for macOS, Linux, or Windows and put it on your `PATH`.

Then, from inside any git repo:
```bash
secretcheck init
```

This writes a pre-commit hook into `.git/hooks/pre-commit` (or
`.husky/pre-commit` if it detects Husky). From then on, every `git commit`
runs the scan automatically — regardless of which install method you used.

## What it catches

Built-in rules cover AWS keys, GitHub/GitLab tokens, Slack tokens and
webhooks, Google API keys, Stripe live keys, Twilio/Mailgun/SendGrid keys,
npm tokens, private key blocks (`-----BEGIN ... PRIVATE KEY-----`), JWTs,
database connection strings with embedded credentials, and generic
`api_key = "..."` / `password = "..."` style assignments.

## When a secret is found

```
✖ secretcheck found 1 potential secret(s):

config/settings.py
  line 12  AWS Access Key ID  AKIA****************WXYZ
    AWS_KEY = "AKIAABCDEFGHIJKLWXYZ"

Secrets were detected above. Commit anyway? (y/N)
```

You get three ways forward:

1. **Fix it** — remove the secret, `git add` again, re-commit.
2. **Answer the prompt** — type `y` to commit anyway (interactive terminals only).
3. **Bypass explicitly** — for scripts/CI or non-interactive shells:
   ```bash
   SECRETCHECK_ALLOW=1 git commit -m "..."
   # or skip all git hooks entirely
   git commit --no-verify -m "..."
   ```

## Ignoring false positives

**Inline, single line:**
```js
const demoKey = "AKIAABCDEFGHIJKLWXYZ"; // secretcheck-disable-line
```

**Inline, next line:**
```js
// secretcheck-disable-next-line
const demoKey = "AKIAABCDEFGHIJKLWXYZ";
```

**By path**, via `.secretcheckignore` in the repo root (gitignore-style globs):
```
fixtures/**
**/*.fixture.json
docs/examples/**
```

**By config**, via `.secretcheckrc.json` in the repo root:
```json
{
  "ignorePaths": ["fixtures/**"],
  "disableRules": ["generic-api-key"],
  "customRules": [
    { "id": "internal-token", "description": "Internal service token", "pattern": "ITK-[A-Z0-9]{24}" }
  ]
}
```
`pattern` is a Go RE2 regular expression. Set `"caseInsensitive": true` on a
custom rule instead of embedding inline flags.

## CLI

```bash
secretcheck init [--force]    # install the pre-commit hook
secretcheck uninstall         # remove it
secretcheck scan              # scan staged files (what the hook runs)
secretcheck scan --all        # scan every tracked file, e.g. for a one-off audit
secretcheck scan --json       # machine-readable output, no prompt (good for CI)
secretcheck scan --no-prompt  # never prompt; just fail if secrets are found
secretcheck version           # print the version
```

## Project layout

```
cmd/secretcheck/        CLI entry point
internal/
  rules/                 built-in detection patterns + placeholder/redaction helpers
  gitutil/                git plumbing (staged files, index content, hooks dir)
  ignore/                 .secretcheckignore + default excludes (custom glob matcher, no deps)
  config/                 .secretcheckrc.json loading + rule resolution
  scanner/                runs rules against staged/working-tree files
  hook/                   installs/removes the pre-commit hook (native + Husky)
  prompt/                 interactive y/N confirmation, TTY/CI detection
  colors/                 minimal ANSI helper
  report/                 findings output formatting
packaging/
  npm/                    main npm shim package + 5 per-platform binary packages
  pypi/                   pyproject.toml/setup.py + per-platform wheel build script
.goreleaser.yaml           cross-platform build config
.github/workflows/         ci.yml (test on every push/PR), plus three independent
                            release workflows triggered by a version tag:
                            release-go.yml (GoReleaser + GitHub Release),
                            release-npm.yml (npm packages), release-pypi.yml
                            (PyPI wheels) — each builds its own binaries, so a
                            failure in one can never block the other two
```

No third-party Go dependencies — smaller attack surface for a security
tool, no `go.sum` to audit, and no network access needed to build.

## Developing

```bash
go build ./...
go vet ./...
go test ./...
```

## Releasing (for maintainers)

1. Make sure `NPM_TOKEN` and `PYPI_API_TOKEN` are set as repository
   secrets in GitHub Settings → Secrets and variables → Actions.
2. Tag and push:
   ```bash
   git tag v0.1.0
   git push origin v0.1.0
   ```
3. That single tag push triggers three independent workflows in parallel:
   - `release-go.yml` — runs GoReleaser to cross-compile binaries and
     publish the GitHub Release. This is also all the Go module needs —
     `go install github.com/anukool23/secretcheck/cmd/secretcheck@v0.1.0`
     resolves directly from the pushed tag via the Go module proxy.
   - `release-npm.yml` — builds its own copy of the binaries
     (`goreleaser build`, no GitHub Release side effects) and publishes
     the npm packages.
   - `release-pypi.yml` — likewise builds its own binaries and publishes
     the PyPI wheels.

   Each workflow is fully self-contained, so a failure in one (say, npm's
   anti-spam scanner flagging a package) can't block the other two, and
   re-running a failed workflow doesn't re-touch the ones that already
   succeeded.

To test packaging locally without waiting on CI (this stages real files and
will actually publish if you have npm auth configured, so review
`.github/scripts/publish-npm.js` before running it against a real npm login):
```bash
goreleaser build --snapshot --clean   # populates dist/ with binaries for every platform
cd packaging/pypi && ./build_wheels.sh  # builds wheelhouse/*.whl locally, no upload
```

## License

MIT
