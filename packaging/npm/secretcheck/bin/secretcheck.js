#!/usr/bin/env node
// Thin shim: locates the prebuilt secretcheck binary for the current
// platform (installed as an optionalDependency) and execs it, forwarding
// argv/stdio/exit code. No third-party dependencies.

const path = require("path");
const { spawnSync } = require("child_process");

const PLATFORM_PACKAGES = {
  "darwin-arm64": "secretcheck-darwin-arm64",
  "darwin-x64": "secretcheck-darwin-x64",
  "linux-arm64": "secretcheck-linux-arm64",
  "linux-x64": "secretcheck-linux-x64",
  "win32-x64": "secretcheck-windows-x64",
};

function resolveBinary() {
  const key = `${process.platform}-${process.arch}`;
  const pkgName = PLATFORM_PACKAGES[key];

  if (!pkgName) {
    console.error(
      `secretcheck: unsupported platform "${process.platform}/${process.arch}".\n` +
        "Supported platforms: darwin/arm64, darwin/x64, linux/arm64, linux/x64, win32/x64.\n" +
        "You can also install a binary directly from GitHub Releases:\n" +
        "  https://github.com/anukool23/secretcheck/releases"
    );
    process.exit(1);
  }

  let pkgJsonPath;
  try {
    pkgJsonPath = require.resolve(`${pkgName}/package.json`);
  } catch (err) {
    console.error(
      `secretcheck: could not find the "${pkgName}" package.\n` +
        "This usually means optional dependencies were skipped during install\n" +
        "(e.g. --no-optional, --omit=optional, or a lockfile mismatch).\n" +
        `Try: npm install ${pkgName} --save-optional`
    );
    process.exit(1);
  }

  const pkgDir = path.dirname(pkgJsonPath);
  const binName = process.platform === "win32" ? "secretcheck.exe" : "secretcheck";
  return path.join(pkgDir, "bin", binName);
}

const binPath = resolveBinary();
const result = spawnSync(binPath, process.argv.slice(2), { stdio: "inherit" });

if (result.error) {
  console.error(`secretcheck: failed to execute ${binPath}: ${result.error.message}`);
  process.exit(1);
}

process.exit(result.status === null ? 1 : result.status);
