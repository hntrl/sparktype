#!/usr/bin/env node

const { execFileSync } = require("child_process");

// Supported platforms - the package name is @sparktype/${platform}-${arch}
const SUPPORTED_PLATFORMS = new Set([
  "darwin-arm64",
  "darwin-x64",
  "linux-arm64",
  "linux-x64",
  "win32-x64",
]);

function getBinaryPath() {
  const platformKey = `${process.platform}-${process.arch}`;

  if (!SUPPORTED_PLATFORMS.has(platformKey)) {
    console.error(`Unsupported platform: ${platformKey}`);
    console.error("");
    console.error("Supported platforms:");
    for (const p of SUPPORTED_PLATFORMS) {
      console.error(`  - ${p}`);
    }
    console.error("");
    console.error("You can download the binary directly from:");
    console.error("  https://github.com/hntrl/sparktype/releases/latest");
    console.error("");
    console.error("Or install via Go:");
    console.error(
      "  go install github.com/hntrl/sparktype/cmd/sparktype@latest"
    );
    process.exit(1);
  }

  const pkgName = `@sparktype/${platformKey}`;

  try {
    // The platform package exports the binary path
    return require(pkgName);
  } catch (err) {
    console.error(`Failed to load platform package: ${pkgName}`);
    console.error("");
    console.error("This usually means the package wasn't installed correctly.");
    console.error("Try reinstalling:");
    console.error("  npm uninstall sparktype && npm install sparktype");
    console.error("");
    console.error("If the problem persists, you can:");
    console.error(
      "  1. Download directly: https://github.com/hntrl/sparktype/releases/latest"
    );
    console.error(
      "  2. Install via Go: go install github.com/hntrl/sparktype/cmd/sparktype@latest"
    );
    process.exit(1);
  }
}

const binaryPath = getBinaryPath();

try {
  execFileSync(binaryPath, process.argv.slice(2), {
    stdio: "inherit",
  });
} catch (err) {
  // execFileSync throws on non-zero exit codes
  if (err.status !== undefined) {
    process.exit(err.status);
  }
  console.error("Failed to run sparktype:", err.message);
  process.exit(1);
}
