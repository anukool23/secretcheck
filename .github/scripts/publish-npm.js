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

// goTarget matches GoReleaser's dist folder naming: secretcheck_<goos>_<goarch>.
const PLATFORMS = [
  { dir: "darwin-arm64", pkg: "secretcheck-darwin-arm64", goTarget: "darwin_arm64", binary: "secretcheck" },
  { dir: "darwin-x64", pkg: "secretcheck-darwin-x64", goTarget: "darwin_amd64", binary: "secretcheck" },
  { dir: "linux-arm64", pkg: "secretcheck-linux-arm64", goTarget: "linux_arm64", binary: "secretcheck" },
  { dir: "linux-x64", pkg: "secretcheck-linux-x64", goTarget: "linux_amd64", binary: "secretcheck" },
  { dir: "win32-x64", pkg: "secretcheck-win32-x64", goTarget: "windows_amd64", binary: "secretcheck.exe" },
];

function readJson(file) {
  return JSON.parse(fs.readFileSync(file, "utf8"));
}

function writeJson(file, obj) {
  fs.writeFileSync(file, JSON.stringify(obj, null, 2) + "\n");
}

function publish(dir) {
  console.log(`\n>> npm publish ${dir}`);
  execFileSync("npm", ["publish", "--access", "public"], { cwd: dir, stdio: "inherit" });
}

// 1. Stage each platform package: copy its binary in, bump its version.
for (const p of PLATFORMS) {
  const pkgDir = path.join(NPM_DIR, "platforms", p.dir);
  const binDir = path.join(pkgDir, "bin");
  fs.mkdirSync(binDir, { recursive: true });

  const src = path.join(DIST_DIR, `secretcheck_${p.goTarget}`, p.binary);
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
for (const p of PLATFORMS) {
  publish(path.join(NPM_DIR, "platforms", p.dir));
}
publish(mainDir);

console.log("\nAll npm packages published.");
