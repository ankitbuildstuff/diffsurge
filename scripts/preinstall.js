#!/usr/bin/env node

/**
 * preinstall.js — Runs during `npm install diffsurge`
 *
 * Downloads the pre-built binary for the current platform from GitHub
 * Releases into bin/surge-engine. Go remains a source-checkout fallback at
 * runtime, but npm installs should not depend on it.
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
  const suffix = getPlatformSuffix();
  if (!suffix) {
    console.warn(
      `⚠ Unsupported platform: ${os.platform()}-${os.arch()}\n` +
        `  Use Docker instead: docker run diffsurge/cli surge --help`
    );
    if (hasGo()) {
      console.warn("  Go was detected, so source-checkout usage can still fall back to `go run`.");
    }
    return; // non-fatal — cli.js will show a helpful error at runtime
  }

  const version = getVersion();
  const binaryName = `surge-${suffix}`;
  const url = `https://github.com/${REPO}/releases/download/v${version}/${binaryName}.gz`;
  const isWindows = os.platform() === "win32";
  const dest = path.join(BIN_DIR, isWindows ? "surge-engine.exe" : "surge-engine");

  if (fs.existsSync(dest)) {
    const stat = fs.statSync(dest);
    if (stat.size > 1024 * 1024) {
      console.log("✔ Found existing surge-engine binary. Skipping download.");
      return;
    }
  }

  console.log(`⬇ Downloading surge v${version} for ${os.platform()}-${os.arch()}…`);

  try {
    const compressed = await download(url);
    const decompressed = zlib.gunzipSync(compressed);

    fs.mkdirSync(BIN_DIR, { recursive: true });
    fs.writeFileSync(dest, decompressed);

    if (!isWindows) {
      fs.chmodSync(dest, 0o755);
    }

    console.log(`✔ Installed surge-engine (${(decompressed.length / 1024 / 1024).toFixed(1)} MB)`);
  } catch (err) {
    console.warn(`⚠ Binary download failed: ${err.message}`);
    if (hasGo()) {
      console.warn(`  Go 1.24+ was detected, so local source checkouts can still use \
\`go run\`.`);
    }
    console.warn(`  Or use Docker: docker run diffsurge/cli surge --help`);
    // non-fatal
  }
}

main().catch((err) => {
  console.warn(`⚠ preinstall warning: ${err.message}`);
});
