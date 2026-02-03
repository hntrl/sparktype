#!/usr/bin/env node
/**
 * Publish all npm packages (platform-specific and main).
 *
 * Called by goreleaser once after all builds complete:
 *   node ./scripts/publish-npm.js <version>
 *
 * This script:
 * 1. Iterates over all platforms defined in platforms.json
 * 2. Copies binaries and publishes each platform package
 * 3. Updates and publishes the main sparktype package
 */

const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");
const crypto = require("crypto");

const ROOT_DIR = path.resolve(__dirname, "..");
const NPM_DIR = path.join(ROOT_DIR, "distributions/npm");
const PLATFORMS_DIR = path.join(NPM_DIR, "platforms");
const DIST_DIR = path.join(ROOT_DIR, "dist");

// Read shared platform config
const platformsConfig = JSON.parse(
  fs.readFileSync(path.join(__dirname, "platforms.json"), "utf8")
);

function log(msg) {
  console.log(msg);
}

function error(msg) {
  console.error(`::error::${msg}`);
}

function success(msg) {
  console.log(`✅ ${msg}`);
}

/**
 * Find the binary in goreleaser output for a given platform
 */
function findBinary(goOs, goArch, binaryName = "sparktype") {
  const platformKey = `${goOs}_${goArch}`;

  // Try common goreleaser output patterns
  const patterns = [
    `sparktype_${platformKey}_v8.0/${binaryName}`,
    `sparktype_${platformKey}_v1/${binaryName}`,
    `sparktype_${platformKey}/${binaryName}`,
    `sparktype_${platformKey}_v2/${binaryName}`,
    `sparktype_${platformKey}_v3/${binaryName}`,
  ];

  for (const pattern of patterns) {
    const fullPath = path.join(DIST_DIR, pattern);
    if (fs.existsSync(fullPath)) {
      return fullPath;
    }
  }

  // Dynamic search - find any directory containing the platform key
  if (fs.existsSync(DIST_DIR)) {
    const entries = fs.readdirSync(DIST_DIR, { withFileTypes: true });
    for (const entry of entries) {
      if (entry.isDirectory() && entry.name.includes(platformKey)) {
        const binaryPath = path.join(DIST_DIR, entry.name, binaryName);
        if (fs.existsSync(binaryPath)) {
          return binaryPath;
        }
      }
    }
  }

  return null;
}

/**
 * Calculate SHA256 checksum
 */
function sha256(filePath) {
  const content = fs.readFileSync(filePath);
  return crypto.createHash("sha256").update(content).digest("hex");
}

/**
 * Read and parse a package.json file
 */
function readPackageJson(filePath) {
  if (!fs.existsSync(filePath)) {
    return null;
  }
  const content = fs.readFileSync(filePath, "utf8");
  if (!content || content.trim() === "") {
    throw new Error(`File is empty: ${filePath}`);
  }
  try {
    return JSON.parse(content);
  } catch (err) {
    throw new Error(`Failed to parse JSON in ${filePath}: ${err.message}`);
  }
}

/**
 * Write package.json
 */
function writePackageJson(filePath, pkg) {
  fs.writeFileSync(filePath, JSON.stringify(pkg, null, 2) + "\n");
}

/**
 * Create .npmrc file for authentication
 */
function ensureNpmAuth(pkgDir) {
  const token = process.env.NODE_AUTH_TOKEN || process.env.NPM_TOKEN;
  if (!token) {
    return false;
  }
  const npmrcPath = path.join(pkgDir, ".npmrc");
  const npmrcContent = `//registry.npmjs.org/:_authToken=${token}\n`;
  fs.writeFileSync(npmrcPath, npmrcContent);
  return true;
}

/**
 * Sleep for specified milliseconds
 */
function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * Publish a package to npm with retry logic for E409 conflicts
 */
async function publishPackage(pkgDir, pkgName, maxRetries = 5) {
  // Ensure npm authentication is configured
  if (!ensureNpmAuth(pkgDir)) {
    return {
      success: false,
      skipped: false,
      error: "NPM_TOKEN or NODE_AUTH_TOKEN environment variable not set",
    };
  }

  const args = ["publish", "--access", "public"];
  if (process.env.GITHUB_ACTIONS) {
    args.push("--provenance");
  }

  for (let attempt = 1; attempt <= maxRetries; attempt++) {
    try {
      execSync(`npm ${args.join(" ")}`, {
        cwd: pkgDir,
        stdio: ["pipe", "pipe", "pipe"],
        env: { ...process.env },
      });

      return { success: true, skipped: false };
    } catch (err) {
      const output = err.stderr?.toString() || err.stdout?.toString() || "";

      // Check if already published
      if (
        /previously published|cannot publish over|already exists|EPUBLISHCONFLICT/i.test(
          output
        )
      ) {
        return { success: true, skipped: true };
      }

      // Check for E409 conflict - registry still processing previous package
      if (/E409|409 Conflict/i.test(output)) {
        if (attempt < maxRetries) {
          const delaySeconds = attempt * 10; // 10s, 20s, 30s, 40s...
          log(
            `⏳ Registry conflict for ${pkgName}, retrying in ${delaySeconds}s (attempt ${attempt}/${maxRetries})...`
          );
          await sleep(delaySeconds * 1000);
          continue;
        }
      }

      return { success: false, skipped: false, error: output };
    }
  }

  return {
    success: false,
    skipped: false,
    error: `Failed after ${maxRetries} attempts`,
  };
}

/**
 * Publish a single platform package
 */
async function publishPlatformPackage(version, platformId, config) {
  const npmKey = config.npm.key;
  const goOs = config.go.os;
  const goArch = config.go.arch;
  const binaryName = config.binaryName || "sparktype";

  log(`\n📦 Publishing @sparktype/${npmKey}...`);

  // Find binary
  const binaryPath = findBinary(goOs, goArch, binaryName);
  if (!binaryPath) {
    error(`Binary not found for ${goOs}_${goArch}`);

    // List what's in dist for debugging
    if (fs.existsSync(DIST_DIR)) {
      const entries = fs.readdirSync(DIST_DIR);
      log(`Contents of ${DIST_DIR}:`);
      for (const e of entries.slice(0, 15)) {
        log(`  ${e}`);
      }
    }
    return false;
  }

  // Verify binary
  const stats = fs.statSync(binaryPath);
  if (stats.size === 0) {
    error(`Binary is empty: ${binaryPath}`);
    return false;
  }

  const binaryHash = sha256(binaryPath);
  log(
    `Binary: ${binaryPath} (${stats.size} bytes, sha256: ${binaryHash.substring(0, 12)}...)`
  );

  // Prepare platform package
  const platformDir = path.join(PLATFORMS_DIR, npmKey);
  const packageJsonPath = path.join(platformDir, "package.json");

  const pkg = readPackageJson(packageJsonPath);
  if (!pkg) {
    error(`package.json not found: ${packageJsonPath}`);
    return false;
  }

  // Copy binary
  const binDir = path.join(platformDir, "bin");
  fs.mkdirSync(binDir, { recursive: true });
  const destBinary = path.join(binDir, binaryName);
  fs.copyFileSync(binaryPath, destBinary);
  fs.chmodSync(destBinary, 0o755);

  // Update version
  pkg.version = version;
  writePackageJson(packageJsonPath, pkg);

  // Publish
  const result = await publishPackage(platformDir, `@sparktype/${npmKey}`);

  if (result.success) {
    if (result.skipped) {
      log(`⚠️  @sparktype/${npmKey}@${version} already exists, skipped`);
    } else {
      success(`Published @sparktype/${npmKey}@${version}`);
    }
    return true;
  } else {
    error(`Failed to publish @sparktype/${npmKey}`);
    if (result.error) {
      console.error(result.error);
    }
    return false;
  }
}

/**
 * Publish the main sparktype package
 */
async function publishMainPackage(version) {
  log(`\n📦 Publishing main sparktype package...`);

  // Read main package.json
  const packageJsonPath = path.join(NPM_DIR, "package.json");
  const pkg = readPackageJson(packageJsonPath);

  if (!pkg) {
    error(`Main package.json not found: ${packageJsonPath}`);
    return false;
  }

  // Update version
  pkg.version = version;

  // Update optionalDependencies versions
  if (!pkg.optionalDependencies) {
    pkg.optionalDependencies = {};
  }
  for (const config of Object.values(platformsConfig.platforms)) {
    pkg.optionalDependencies[`@sparktype/${config.npm.key}`] = version;
  }

  writePackageJson(packageJsonPath, pkg);
  log(`Updated package.json with version ${version}`);

  // Publish
  const result = await publishPackage(NPM_DIR, "sparktype");

  if (result.success) {
    if (result.skipped) {
      log(`⚠️  sparktype@${version} already exists, skipped`);
    } else {
      success(`Published sparktype@${version}`);
    }
    return true;
  } else {
    error("Failed to publish sparktype");
    if (result.error) {
      console.error(result.error);
    }
    return false;
  }
}

// Delay between publishing packages to avoid npm registry conflicts (in ms)
const PUBLISH_DELAY_MS = 15000; // 15 seconds

async function main() {
  const version = process.argv[2];

  if (!version) {
    console.error("Usage: publish-npm.js <version>");
    console.error("Example: publish-npm.js 0.1.0");
    process.exit(1);
  }

  log(`🚀 Publishing npm packages v${version}`);
  log(`================================================`);

  // Publish all platform packages first (sequentially with delays)
  let allSucceeded = true;
  const platforms = Object.entries(platformsConfig.platforms);

  log(`\nPublishing ${platforms.length} platform packages sequentially...`);

  for (let i = 0; i < platforms.length; i++) {
    const [platformId, config] = platforms[i];

    // Add delay between publishes (except for the first one)
    if (i > 0) {
      log(`\n⏳ Waiting ${PUBLISH_DELAY_MS / 1000}s before next publish...`);
      await sleep(PUBLISH_DELAY_MS);
    }

    const succeeded = await publishPlatformPackage(version, platformId, config);
    if (!succeeded) {
      allSucceeded = false;
      // Continue with other platforms even if one fails
    }
  }

  if (!allSucceeded) {
    error("Some platform packages failed to publish");
    process.exit(1);
  }

  // Wait before publishing main package
  log(`\n⏳ Waiting ${PUBLISH_DELAY_MS / 1000}s before publishing main package...`);
  await sleep(PUBLISH_DELAY_MS);

  // Publish main package after all platforms succeed
  log(`\n================================================`);
  const mainSucceeded = await publishMainPackage(version);

  if (!mainSucceeded) {
    process.exit(1);
  }

  log(`\n================================================`);
  success(`All npm packages published successfully!`);
}

main().catch((err) => {
  error(err.message);
  process.exit(1);
});

