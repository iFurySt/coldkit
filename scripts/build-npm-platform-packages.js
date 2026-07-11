#!/usr/bin/env node
"use strict";

const { execFileSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

const rootDir = path.resolve(__dirname, "..");
const rootPackage = require(path.join(rootDir, "package.json"));
const outDir = path.join(rootDir, "npm", "dist-packages");

const targets = [
  { nodePlatform: "darwin", nodeArch: "x64", goos: "darwin", goarch: "amd64" },
  { nodePlatform: "darwin", nodeArch: "arm64", goos: "darwin", goarch: "arm64" },
  { nodePlatform: "linux", nodeArch: "x64", goos: "linux", goarch: "amd64" },
  { nodePlatform: "linux", nodeArch: "arm64", goos: "linux", goarch: "arm64" },
  { nodePlatform: "win32", nodeArch: "x64", goos: "windows", goarch: "amd64" },
  { nodePlatform: "win32", nodeArch: "arm64", goos: "windows", goarch: "arm64" }
];

fs.rmSync(outDir, { recursive: true, force: true });

for (const target of targets) {
  const packageName = `@ifuryst/coldkit-${target.nodePlatform}-${target.nodeArch}`;
  const packageDir = path.join(outDir, packageName);
  const binDir = path.join(packageDir, "bin");
  const extension = target.goos === "windows" ? ".exe" : "";

  fs.mkdirSync(binDir, { recursive: true });

  buildGoBinary(target, path.join(binDir, `ck${extension}`), "./cmd/ck");
  buildGoBinary(target, path.join(binDir, `ck-mcp${extension}`), "./cmd/ck-mcp");

  fs.writeFileSync(
    path.join(packageDir, "package.json"),
    `${JSON.stringify(platformPackage(packageName, target), null, 2)}\n`
  );
  fs.copyFileSync(path.join(rootDir, "LICENSE"), path.join(packageDir, "LICENSE"));
  fs.copyFileSync(path.join(rootDir, "README.md"), path.join(packageDir, "README.md"));
}

function buildGoBinary(target, output, pkg) {
  execFileSync("go", ["build", "-trimpath", "-ldflags=-s -w", "-o", output, pkg], {
    cwd: rootDir,
    env: {
      ...process.env,
      CGO_ENABLED: "0",
      GOOS: target.goos,
      GOARCH: target.goarch
    },
    stdio: "inherit"
  });
}

function platformPackage(packageName, target) {
  const extension = target.goos === "windows" ? ".exe" : "";

  return {
    name: packageName,
    version: rootPackage.version,
    description: `${rootPackage.description} (${target.nodePlatform}/${target.nodeArch})`,
    license: rootPackage.license,
    repository: rootPackage.repository,
    homepage: rootPackage.homepage,
    bugs: rootPackage.bugs,
    os: [target.nodePlatform],
    cpu: [target.nodeArch],
    bin: {
      ck: `bin/ck${extension}`,
      "ck-mcp": `bin/ck-mcp${extension}`
    },
    files: ["bin/", "LICENSE", "README.md"],
    publishConfig: rootPackage.publishConfig
  };
}
