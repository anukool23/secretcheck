#!/usr/bin/env node
// Publishes the six secretcheck npm packages (5 platform packages + the
// main shim package) for a release.
//
// Usage (from repo root, after GoReleaser has populated dist/):
//   VERSION=1.2.3 node .github/scripts/publish-npm.js
//
// Requires NODE_AUTH_TOKEN (set by actions/setup-node from the NPM_TOKEN
// secret) to already be wired into ~/.npmrc via the registry-url input.

const fs = require("fs");
const path = require("path");
const { execFileSync } = require("child_process");

const VERSION = process.env.VERSION;
if (!VERSION) {
  console.error("VERSION env var is required");
  process.exit(1);
}

const ROOT = path.resolve(__dirname, "..", "..");
const NPM_DIR = path.join(ROOT, "packaging", "npm");
const DIST_DIR = path.join(ROOT, "dist");

// goos/goarch are matched as substrings against whatever dist folder name
// GoReleaser actually produced, rather than assuming an exact template
// (GoReleaser has changed this naming across versions, e.g. adding a
// "_v1" suffix to amd64 builds by default) — see findBinary() below.
const PLATFORMS = [
  { dir: "darwin-arm64", pkg: "secretcheck-darwin-arm64", goos: "darwin", goarch: "arm64", binary: "secretcheck" },
  { dir: "darwin-x64", pkg: "secretcheck-darwin-x64", goos: "darwin", goarch: "amd64", binary: "secretcheck" },
  { dir: "linux-arm64", pkg: "secretcheck-linux-arm64", goos: "linux", goarch: "arm64", binary: "secretcheck" },
  { dir: "linux-x64", pkg: "secretcheck-linux-x64", goos: "linux", goarch: "amd64", binary: "secretcheck" },
  { dir: "win32-x64", pkg: "secretcheck-win32-x64", goos: "windows", goarch: "amd64", binary: "secretcheck.exe" },
];

function readJson(file) {
  return JSON.parse(fs.readFileSync(file, "utf8"));
}

function writeJson(file, obj) {
  fs.writeFileSync(file, JSON.stringify(obj, null, 2) + "\n");
}

// Returns true if name@version is already live on the registry. Used to
// make this script safely re-runnable: npm permanently refuses to publish
// over an existing version, so on a partial failure (one package rejected,
// e.g. by npm's anti-abuse scanner) we must skip everything that already
// succeeded rather than re-attempting it.
function alreadyPublished(name, version) {
  try {
    execFileSync("npm", ["view", `${name}@${version}`, "version"], { stdio: "pipe" });
    return true;
  } catch {
    return false;
  }
}

function publish(dir, name, version) {
  if (alreadyPublished(name, version)) {
    console.log(`\n>> ${name}@${version} is already published — skipping.`);
    return;
  }
  console.log(`\n>> npm publish ${dir}`);
  execFileSync("npm", ["publish", "--access", "public"], { cwd: dir, stdio: "inherit" });
}

// Finds the built binary for a given goos/goarch by scanning dist/ for a
// directory whose name contains both, instead of assuming an exact naming
// template. Throws a descriptive error (including what dist/ actually
// contains) if nothing matches, so failures are easy to diagnose.
function findBinary(distDir, goos, goarch, binaryName) {
  if (!fs.existsSync(distDir)) {
    throw new Error(
      `dist/ does not exist at ${distDir} — did the GoReleaser step run/succeed before this one?`
    );
  }

  const entries = fs.readdirSync(distDir, { withFileTypes: true });
  const candidateDirs = entries
    .filter((e) => e.isDirectory())
    .map((e) => e.name)
    .filter((name) => name.includes(goos) && name.includes(goarch));

  for (const dirName of candidateDirs) {
    const candidate = path.join(distDir, dirName, binaryName);
    if (fs.existsSync(candidate)) {
      return candidate;
    }
  }

  const allDirs = entries.filter((e) => e.isDirectory()).map((e) => e.name);
  throw new Error(
    `Could not find a "${binaryName}" binary for ${goos}/${goarch} under ${distDir}.\n` +
      `Directories in dist/: ${allDirs.join(", ") || "(none)"}`
  );
}

// 1. Stage each platform package: copy its binary in, bump its version.
for (const p of PLATFORMS) {
  const pkgDir = path.join(NPM_DIR, "platforms", p.dir);
  const binDir = path.join(pkgDir, "bin");
  fs.mkdirSync(binDir, { recursive: true });

  const src = findBinary(DIST_DIR, p.goos, p.goarch, p.binary);
  const dest = path.join(binDir, p.binary);
  fs.copyFileSync(src, dest);
  fs.chmodSync(dest, 0o755);

  const pkgJsonPath = path.join(pkgDir, "package.json");
  const pkgJson = readJson(pkgJsonPath);
  pkgJson.version = VERSION;
  writeJson(pkgJsonPath, pkgJson);
}

// 2. Bump the main package's version and its optionalDependencies so they
//    point at this exact release.
const mainDir = path.join(NPM_DIR, "secretcheck");
const mainPkgPath = path.join(mainDir, "package.json");
const mainPkg = readJson(mainPkgPath);
mainPkg.version = VERSION;
for (const p of PLATFORMS) {
  mainPkg.optionalDependencies[p.pkg] = VERSION;
}
writeJson(mainPkgPath, mainPkg);

// 3. Publish platform packages first, then the main package, so its
//    optionalDependencies resolve immediately for anyone installing it.
//    Each publish is skipped if that exact name@version is already live,
//    so re-running this script after a partial failure is safe.
//
//    A platform package failing to publish (e.g. npm's anti-spam scanner
//    blocking a package with a bundled .exe) does NOT stop the others or
//    the main package: npm's optionalDependencies mechanism already
//    tolerates a listed optional version that doesn't exist — `npm install`
//    just warns and skips it, it doesn't fail. So one blocked platform
//    should never prevent everyone else from getting a working release.
const failures = [];
for (const p of PLATFORMS) {
  try {
    publish(path.join(NPM_DIR, "platforms", p.dir), p.pkg, VERSION);
  } catch (err) {
    console.error(`\n>> FAILED to publish ${p.pkg}@${VERSION}: ${err.message}`);
    console.error(">> Continuing with the remaining packages — this one can be retried later.");
    failures.push(p.pkg);
  }
}
publish(mainDir, mainPkg.name, VERSION);

if (failures.length > 0) {
  console.error(
    `\n${failures.length} package(s) failed to publish: ${failures.join(", ")}.\n` +
      "Everything else published successfully. Once the underlying issue is resolved " +
      "(e.g. npm support clears an anti-spam flag), re-run this script/workflow with the " +
      "same VERSION — already-published packages are skipped automatically."
  );
  process.exitCode = 1;
} else {
  console.log("\nAll npm packages published.");
}
