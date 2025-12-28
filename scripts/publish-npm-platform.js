#!/usr/bin/env node
/**
 * Publish a single npm platform package based on the goreleaser artifact.
 *
 * Called by goreleaser publisher once per archive:
 *   node ./scripts/publish-npm-platform.js 0.1.0 sparktype_0.1.0_darwin_arm64.tar.gz
 *
 * This script:
 * 1. Parses the artifact name to determine platform
 * 2. Extracts/copies the binary to the platform package
 * 3. Publishes just that platform package
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
 * Parse artifact name to extract OS and arch
 * e.g., "sparktype_0.1.0_darwin_arm64.tar.gz" -> { os: "darwin", arch: "arm64" }
 */
function parseArtifact(artifactName) {
  // Pattern: sparktype_VERSION_OS_ARCH.ext
  const match = artifactName.match(/^sparktype_[^_]+_([^_]+)_([^.]+)\./);
  if (!match) {
    return null;
  }
  return { os: match[1], arch: match[2] };
}

/**
 * Find the platform config that matches the given Go OS/arch
 */
function findPlatformConfig(goOs, goArch) {
  for (const [platformId, config] of Object.entries(
    platformsConfig.platforms
  )) {
    if (config.go.os === goOs && config.go.arch === goArch) {
      return { platformId, config };
    }
  }
  return null;
}

/**
 * Find the binary in goreleaser output
 */
function findBinary(goOs, goArch, binaryName = "sparktype") {
  const platformKey = `${goOs}_${goArch}`;

  const patterns = [
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

  // Dynamic search
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
  return JSON.parse(fs.readFileSync(filePath, "utf8"));
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
 * Publish a package to npm
 */
function publishPackage(pkgDir, pkgName) {
  try {
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

    execSync(`npm ${args.join(" ")}`, {
      cwd: pkgDir,
      stdio: ["pipe", "pipe", "pipe"],
      env: { ...process.env },
    });

    return { success: true, skipped: false };
  } catch (err) {
    const output = err.stderr?.toString() || err.stdout?.toString() || "";

    if (
      /previously published|cannot publish over|already exists|EPUBLISHCONFLICT/i.test(
        output
      )
    ) {
      return { success: true, skipped: true };
    }

    return { success: false, skipped: false, error: output };
  }
}

async function main() {
  const version = process.argv[2];
  const artifact = process.argv[3];

  if (!version || !artifact) {
    console.error("Usage: publish-npm-platform.js <version> <artifact>");
    console.error(
      "Example: publish-npm-platform.js 0.1.0 sparktype_0.1.0_darwin_arm64.tar.gz"
    );
    process.exit(1);
  }

  log(`Processing artifact: ${artifact}`);

  // Parse artifact to get OS/arch
  const parsed = parseArtifact(artifact);
  if (!parsed) {
    log(`Skipping non-platform artifact: ${artifact}`);
    process.exit(0);
  }

  const { os: goOs, arch: goArch } = parsed;
  log(`Detected platform: ${goOs}/${goArch}`);

  // Find matching platform config
  const platformMatch = findPlatformConfig(goOs, goArch);
  if (!platformMatch) {
    error(`No npm platform config found for ${goOs}/${goArch}`);
    process.exit(1);
  }

  const { platformId, config } = platformMatch;
  const npmKey = config.npm.key;
  const binaryName = config.binaryName || "sparktype";

  log(`Publishing @sparktype/${npmKey}...`);

  // Find binary
  const binaryPath = findBinary(goOs, goArch, binaryName);
  if (!binaryPath) {
    error(`Binary not found for ${goOs}_${goArch}`);

    // List what's in dist for debugging
    if (fs.existsSync(DIST_DIR)) {
      const entries = fs.readdirSync(DIST_DIR);
      error(`Contents of ${DIST_DIR}:`);
      for (const e of entries.slice(0, 10)) {
        error(`  ${e}`);
      }
    }
    process.exit(1);
  }

  // Verify binary
  const stats = fs.statSync(binaryPath);
  if (stats.size === 0) {
    error(`Binary is empty: ${binaryPath}`);
    process.exit(1);
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
    process.exit(1);
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
  const result = publishPackage(platformDir, `@sparktype/${npmKey}`);

  if (result.success) {
    if (result.skipped) {
      log(`⚠️  @sparktype/${npmKey}@${version} already exists, skipped`);
    } else {
      success(`Published @sparktype/${npmKey}@${version}`);
    }
  } else {
    error(`Failed to publish @sparktype/${npmKey}`);
    if (result.error) {
      console.error(result.error);
    }
    process.exit(1);
  }
}

main().catch((err) => {
  error(err.message);
  process.exit(1);
});
