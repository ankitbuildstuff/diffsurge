#!/usr/bin/env node

/**
 * preinstall.js — Runs during `npm install diffsurge`
 *
 * 1. Checks if Go is available (for local `go run` fallback)
 * 2. If no Go, downloads the pre-built binary for the current platform
 *    from GitHub Releases into bin/surge-engine
 */

const os = require("os");
const fs = require("fs");
const path = require("path");
const https = require("https");
const zlib = require("zlib");
const { execSync } = require("child_process");

const REPO = "ankitbuildstuff/diffsurge";
const BIN_DIR = path.join(__dirname, "..", "bin");

// ── platform mapping ──────────────────────────────────────────────────────

function getPlatformSuffix() {
  const platform = os.platform();
  const arch = os.arch();

  const mapping = {
    "darwin-x64": "darwin-amd64",
    "darwin-arm64": "darwin-arm64",
    "linux-x64": "linux-amd64",
    "linux-arm64": "linux-arm64",
    "win32-x64": "windows-amd64.exe",
  };

  const key = `${platform}-${arch}`;
  return mapping[key] || null;
}

// ── helpers ────────────────────────────────────────────────────────────────

function hasGo() {
  try {
    execSync("go version", { stdio: "ignore" });
    return true;
  } catch {
    return false;
  }
}

function getVersion() {
  const pkg = JSON.parse(
    fs.readFileSync(path.join(__dirname, "..", "package.json"), "utf8")
  );
  return pkg.version;
}

function download(url) {
  return new Promise((resolve, reject) => {
    https
      .get(url, (res) => {
        if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
          return download(res.headers.location).then(resolve, reject);
        }
        if (res.statusCode !== 200) {
          return reject(new Error(`HTTP ${res.statusCode} for ${url}`));
        }
        const chunks = [];
        res.on("data", (c) => chunks.push(c));
        res.on("end", () => resolve(Buffer.concat(chunks)));
        res.on("error", reject);
      })
      .on("error", reject);
  });
}

// ── main ───────────────────────────────────────────────────────────────────

async function main() {
  // If Go is available, the CLI wrapper (bin/cli.js) can use `go run` directly.
  // Skip binary download to keep install fast.
  if (hasGo()) {
    console.log("✔ Go detected — surge will use `go run` as fallback. Skipping binary download.");
    return;
  }

  const suffix = getPlatformSuffix();
  if (!suffix) {
    console.warn(
      `⚠ Unsupported platform: ${os.platform()}-${os.arch()}\n` +
        `  Use Docker instead: docker run diffsurge/cli surge --help`
    );
    return; // non-fatal — cli.js will show a helpful error at runtime
  }

  const version = getVersion();
  const binaryName = `surge-${suffix}`;
  const url = `https://github.com/${REPO}/releases/download/v${version}/${binaryName}.gz`;

  console.log(`⬇ Downloading surge v${version} for ${os.platform()}-${os.arch()}…`);

  try {
    const compressed = await download(url);
    const decompressed = zlib.gunzipSync(compressed);

    fs.mkdirSync(BIN_DIR, { recursive: true });

    const isWindows = os.platform() === "win32";
    const dest = path.join(BIN_DIR, isWindows ? "surge-engine.exe" : "surge-engine");
    fs.writeFileSync(dest, decompressed);

    if (!isWindows) {
      fs.chmodSync(dest, 0o755);
    }

    console.log(`✔ Installed surge-engine (${(decompressed.length / 1024 / 1024).toFixed(1)} MB)`);
  } catch (err) {
    console.warn(`⚠ Binary download failed: ${err.message}`);
    console.warn(`  surge will still work if Go 1.24+ is installed.`);
    console.warn(`  Or use Docker: docker run diffsurge/cli surge --help`);
    // non-fatal
  }
}

main().catch((err) => {
  console.warn(`⚠ preinstall warning: ${err.message}`);
});
