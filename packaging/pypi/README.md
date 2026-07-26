# secretcheck

A git pre-commit hook that scans staged files for secrets — API keys,
tokens, private keys, credentials — and blocks the commit if it finds any,
with an explicit option to override.

```bash
pip install secretcheck
secretcheck init      # installs the pre-commit hook in the current repo
```

This package bundles a prebuilt Go binary for your platform; `secretcheck`
and `python -m secretcheck` both just launch it. Full docs, rule list, and
config options: https://github.com/anukool23/secretcheck
