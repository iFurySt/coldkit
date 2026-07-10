"use strict";

const { spawnSync } = require("node:child_process");
const path = require("node:path");

const platformMap = {
  darwin: { packagePlatform: "darwin", goos: "darwin" },
  linux: { packagePlatform: "linux", goos: "linux" },
  win32: { packagePlatform: "win32", goos: "windows" }
};

const archMap = {
  arm64: { packageArch: "arm64", goarch: "arm64" },
  x64: { packageArch: "x64", goarch: "amd64" }
};

function runBinary(name) {
  const platform = platformMap[process.platform];
  const arch = archMap[process.arch];

  if (!platform || !arch) {
    console.error(`coldkit does not ship a ${process.platform}/${process.arch} binary.`);
    process.exit(1);
  }

  const executable = process.platform === "win32" ? `${name}.exe` : name;
  const packageName = `@ifuryst/coldkit-${platform.packagePlatform}-${arch.packageArch}`;
  const binaryPath = resolvePackagedBinary(packageName, executable);
  const result = spawnSync(binaryPath, process.argv.slice(2), { stdio: "inherit" });

  if (result.error) {
    console.error(`Failed to run ${binaryPath}: ${result.error.message}`);
    process.exit(1);
  }

  process.exit(result.status === null ? 1 : result.status);
}

function resolvePackagedBinary(packageName, executable) {
  try {
    const packageJsonPath = require.resolve(`${packageName}/package.json`);
    return path.join(path.dirname(packageJsonPath), "bin", executable);
  } catch (error) {
    const localBinary = path.join(
      __dirname,
      "..",
      "dist-packages",
      packageName,
      "bin",
      executable
    );

    if (error && error.code === "MODULE_NOT_FOUND") {
      return localBinary;
    }

    throw error;
  }
}

module.exports = { runBinary };
