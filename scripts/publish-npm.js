#!/usr/bin/env node
/**
 * Publish npm packages for sparktype.
 *
 * This script:
 * 1. Reads platform config from shared platforms.json
 * 2. Updates existing package.json files (version + optionalDependencies)
 * 3. Verifies binaries exist and are valid
 * 4. Publishes packages in the correct order with proper error handling
 *
 * The package.json files in distributions/npm are the source of truth for
 * package metadata (name, description, keywords, etc.). This script only
 * updates the version field (and optionalDependencies for the main package).
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

function warn(msg) {
  console.warn(`⚠️  ${msg}`);
}

function success(msg) {
  console.log(`✅ ${msg}`);
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
 * Write package.json with consistent formatting
 */
function writePackageJson(filePath, pkg) {
  fs.writeFileSync(filePath, JSON.stringify(pkg, null, 2) + "\n");
}

/**
 * Find the binary for a platform in goreleaser output
 */
function findBinary(goOs, goArch, binaryName = "sparktype") {
  const platformKey = `${goOs}_${goArch}`;

  // Try common goreleaser output patterns
  const patterns = [
    `sparktype_${platformKey}_v1/${binaryName}`,
    `sparktype_${platformKey}/${binaryName}`,
  ];

  for (const pattern of patterns) {
    const fullPath = path.join(DIST_DIR, pattern);
    if (fs.existsSync(fullPath)) {
      return fullPath;
    }
  }

  return null;
}

/**
 * Calculate SHA256 checksum of a file
 */
function sha256(filePath) {
  const content = fs.readFileSync(filePath);
  return crypto.createHash("sha256").update(content).digest("hex");
}

/**
 * Publish a package to npm
 * Returns: { success: boolean, skipped: boolean, error?: string }
 */
function publishPackage(pkgDir, pkgName) {
  try {
    const args = ["publish", "--access", "public"];

    // Add provenance if running in GitHub Actions
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

    // Check if it's an "already exists" error
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

  if (!version) {
    console.error("Usage: publish-npm.js <version>");
    process.exit(1);
  }

  log(`Publishing npm packages version ${version}...`);

  // Verify dist directory exists
  if (!fs.existsSync(DIST_DIR)) {
    error(`Dist directory not found: ${DIST_DIR}`);
    process.exit(1);
  }

  // Step 1: Prepare platform packages
  log("\nPreparing platform packages...");
  const preparedPlatforms = [];

  for (const [platformId, config] of Object.entries(
    platformsConfig.platforms
  )) {
    const npmKey = config.npm.key;
    const platformDir = path.join(PLATFORMS_DIR, npmKey);
    const packageJsonPath = path.join(platformDir, "package.json");

    // Read existing package.json
    const pkg = readPackageJson(packageJsonPath);
    if (!pkg) {
      error(`package.json not found for platform ${npmKey}`);
      error(`  Expected at: ${packageJsonPath}`);
      process.exit(1);
    }

    // Find binary
    const binaryName = config.binaryName || "sparktype";
    const binaryPath = findBinary(config.go.os, config.go.arch, binaryName);

    if (!binaryPath) {
      error(`Binary not found for ${platformId}`);
      error(
        `  Looked for ${config.go.os}_${config.go.arch} binary in ${DIST_DIR}`
      );
      process.exit(1);
    }

    // Verify binary is not empty
    const stats = fs.statSync(binaryPath);
    if (stats.size === 0) {
      error(`Binary for ${platformId} is empty: ${binaryPath}`);
      process.exit(1);
    }

    // Log binary info
    const binaryHash = sha256(binaryPath);
    log(
      `  ${npmKey}: ${path.basename(binaryPath)} (${stats.size} bytes, sha256: ${binaryHash.substring(0, 12)}...)`
    );

    // Copy binary to package
    const binDir = path.join(platformDir, "bin");
    fs.mkdirSync(binDir, { recursive: true });
    const destBinary = path.join(binDir, binaryName);
    fs.copyFileSync(binaryPath, destBinary);
    fs.chmodSync(destBinary, 0o755);

    // Update only the version in package.json
    pkg.version = version;
    writePackageJson(packageJsonPath, pkg);

    preparedPlatforms.push({ npmKey, dir: platformDir });
  }

  success(`Prepared ${preparedPlatforms.length} platform packages`);

  // Step 2: Update main package.json
  log("\nPreparing main package...");
  const mainPackageJsonPath = path.join(NPM_DIR, "package.json");
  const mainPkg = readPackageJson(mainPackageJsonPath);

  if (!mainPkg) {
    error(`Main package.json not found at: ${mainPackageJsonPath}`);
    process.exit(1);
  }

  // Update version
  mainPkg.version = version;

  // Update optionalDependencies versions
  if (!mainPkg.optionalDependencies) {
    mainPkg.optionalDependencies = {};
  }
  for (const config of Object.values(platformsConfig.platforms)) {
    mainPkg.optionalDependencies[`@sparktype/${config.npm.key}`] = version;
  }

  writePackageJson(mainPackageJsonPath, mainPkg);
  success(`Updated main package.json to version ${version}`);

  // Step 3: Publish platform packages
  log("\nPublishing platform packages...");
  let failed = false;

  for (const { npmKey, dir } of preparedPlatforms) {
    const pkgName = `@sparktype/${npmKey}`;
    process.stdout.write(`  Publishing ${pkgName}... `);

    const result = publishPackage(dir, pkgName);

    if (result.success) {
      if (result.skipped) {
        console.log("already exists, skipped");
      } else {
        console.log("done");
      }
    } else {
      console.log("FAILED");
      error(`Failed to publish ${pkgName}`);
      if (result.error) {
        console.error(result.error);
      }
      failed = true;
    }
  }

  if (failed) {
    error("One or more platform packages failed to publish");
    process.exit(1);
  }

  success("All platform packages published");

  // Step 4: Publish main package
  log("\nPublishing main sparktype package...");
  const mainResult = publishPackage(NPM_DIR, "sparktype");

  if (mainResult.success) {
    if (mainResult.skipped) {
      warn("Main package already exists, skipped");
    } else {
      success("Main package published");
    }
  } else {
    error("Failed to publish main sparktype package");
    if (mainResult.error) {
      console.error(mainResult.error);
    }
    process.exit(1);
  }

  log("");
  success(`npm packages v${version} published successfully!`);
}

main().catch((err) => {
  error(err.message);
  process.exit(1);
});
