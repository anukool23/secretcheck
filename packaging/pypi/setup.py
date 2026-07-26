"""
Custom build hook: forces the wheel's platform tag to whatever platform the
bundled Go binary was built for (via the SECRETCHECK_WHEEL_PLATFORM env
var), and marks the wheel as platform-specific ("not pure Python") so pip
picks the matching one instead of trying to build from source.

Everything else (name, version, entry point, package data) is declared in
pyproject.toml; this file only exists for this one piece of build-time
customization that pyproject.toml alone can't express.
"""

import os

from setuptools import setup
from setuptools.dist import Distribution

try:
    from wheel.bdist_wheel import bdist_wheel as _bdist_wheel

    class bdist_wheel(_bdist_wheel):  # noqa: N801 - setuptools convention
        def finalize_options(self):
            _bdist_wheel.finalize_options(self)
            self.root_is_pure = False

        def get_tag(self):
            _python, _abi, plat = _bdist_wheel.get_tag(self)
            plat = os.environ.get("SECRETCHECK_WHEEL_PLATFORM", plat)
            return "py3", "none", plat

    cmdclass = {"bdist_wheel": bdist_wheel}
except ImportError:
    cmdclass = {}


class BinaryDistribution(Distribution):
    """Tells setuptools this package contains a compiled binary, even
    though there's no C extension to build."""

    def has_ext_modules(self):
        return True


setup(distclass=BinaryDistribution, cmdclass=cmdclass)
