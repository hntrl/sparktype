#!/usr/bin/env node
/**
 * Publish the main sparktype npm package.
 *
 * This should be called AFTER all platform packages are published.
 * It updates the main package.json with the version and optionalDependencies,
 * then publishes the main sparktype package.
 *
 * Usage: node ./scripts/publish-npm-main.js <version>
 */

const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const ROOT_DIR = path.resolve(__dirname, "..");
const NPM_DIR = path.join(ROOT_DIR, "distributions/npm");

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
 * Publish a package to npm
 */
function publishPackage(pkgDir, pkgName) {
  try {
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

  if (!version) {
    console.error("Usage: publish-npm-main.js <version>");
    process.exit(1);
  }

  log(`Publishing main sparktype package v${version}...`);

  // Read main package.json
  const packageJsonPath = path.join(NPM_DIR, "package.json");
  const pkg = readPackageJson(packageJsonPath);

  if (!pkg) {
    error(`Main package.json not found: ${packageJsonPath}`);
    process.exit(1);
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
  const result = publishPackage(NPM_DIR, "sparktype");

  if (result.success) {
    if (result.skipped) {
      log(`⚠️  sparktype@${version} already exists, skipped`);
    } else {
      success(`Published sparktype@${version}`);
    }
  } else {
    error("Failed to publish sparktype");
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
