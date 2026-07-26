"""Execs the bundled secretcheck binary, forwarding argv and exit code.

The actual scanning logic lives in the Go binary shipped inside this
platform-specific wheel (see packaging/pypi/build_wheels.sh); this module
is just a launcher so `pip install secretcheck` gives you a `secretcheck`
command and `python -m secretcheck` both work.
"""

import os
import subprocess
import sys


def _binary_name() -> str:
    return "secretcheck.exe" if os.name == "nt" else "secretcheck"


def main() -> None:
    here = os.path.dirname(os.path.abspath(__file__))
    binary = os.path.join(here, "bin", _binary_name())

    if not os.path.exists(binary):
        sys.stderr.write(
            "secretcheck: no bundled binary found for this platform.\n"
            "This wheel may not match your platform/architecture. Try:\n"
            "  pip install --force-reinstall --no-cache-dir secretcheck\n"
            "or download a binary directly from:\n"
            "  https://github.com/anukool23/secretcheck/releases\n"
        )
        sys.exit(1)

    if os.name != "nt":
        try:
            os.chmod(binary, 0o755)
        except OSError:
            pass

    result = subprocess.run([binary, *sys.argv[1:]])
    sys.exit(result.returncode)


if __name__ == "__main__":
    main()
