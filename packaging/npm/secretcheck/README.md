# secretcheck

A git pre-commit hook that scans staged files for secrets — API keys,
tokens, private keys, credentials — and blocks the commit if it finds any,
with an explicit option to override.

```bash
npm install --save-dev secretcheck
secretcheck init      # installs the pre-commit hook in the current repo
```

This package is a thin shim; the actual scanner is a prebuilt Go binary
installed automatically for your platform via `optionalDependencies`.

Full docs, rule list, and config options: https://docs.anukool.me/secretcheck
Source: https://github.com/anukool23/secretcheck
