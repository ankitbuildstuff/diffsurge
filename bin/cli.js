#!/usr/bin/env node

/**
 * DiffSurge CLI entry-point.
 *
 * Resolution order:
 *   1. Pre-built binary downloaded during preinstall  (bin/surge-engine)
 *   2. Locally-built Go binary via `go run`           (source checkout only)
 *   3. Helpful error with install instructions
 */

const { execFileSync, execSync } = require("child_process");
const path = require("path");
const os = require("os");
const fs = require("fs");

// ── helpers ────────────────────────────────────────────────────────────────

function loadDotEnv() {
  const envPath = path.join(process.cwd(), ".env");
  try {
    const content = fs.readFileSync(envPath, "utf8");
    for (const line of content.split("\n")) {
      const trimmed = line.trim();
      if (!trimmed || trimmed.startsWith("#")) continue;
      const eqIdx = trimmed.indexOf("=");
      if (eqIdx === -1) continue;
      const key = trimmed.slice(0, eqIdx).trim();
      let val = trimmed.slice(eqIdx + 1).trim();
      if (
        (val.startsWith('"') && val.endsWith('"')) ||
        (val.startsWith("'") && val.endsWith("'"))
      ) {
        val = val.slice(1, -1);
      }
      if (!(key in process.env)) {
        process.env[key] = val;
      }
    }
  } catch {
    // .env not found — that's fine
  }
}

function hasGo() {
  try {
    execSync("go version", { stdio: "ignore" });
    return true;
  } catch {
    return false;
  }
}

// ── main ───────────────────────────────────────────────────────────────────

loadDotEnv();

const isWindows = os.platform() === "win32";
const engineName = isWindows ? "surge-engine.exe" : "surge-engine";
const binaryPath = path.join(__dirname, engineName);

// 1) Try pre-built binary
if (fs.existsSync(binaryPath)) {
  try {
    execFileSync(binaryPath, process.argv.slice(2), {
      stdio: "inherit",
      env: process.env,
    });
    process.exit(0);
  } catch (err) {
    if (err.status !== undefined) process.exit(err.status);
    // fall through to Go
  }
}

// 2) Try local Go build from a source checkout
const goMain = path.resolve(__dirname, "..", "diffsurge-go", "cmd", "cli", "main.go");
if (hasGo() && fs.existsSync(goMain)) {
  try {
    execFileSync("go", ["run", goMain, ...process.argv.slice(2)], {
      stdio: "inherit",
      env: process.env,
      cwd: path.resolve(__dirname, "..", "diffsurge-go"),
    });
    process.exit(0);
  } catch (err) {
    if (err.status !== undefined) process.exit(err.status);
  }
}

// 3) Nothing worked
console.error("╔══════════════════════════════════════════════════════════════╗");
console.error("║  surge binary not found.                                    ║");
console.error("║                                                            ║");
console.error("║  The npm install step should have downloaded a prebuilt    ║");
console.error("║  engine for this platform.                                  ║");
console.error("║                                                            ║");
console.error("║  Fix options:                                              ║");
console.error("║    npm rebuild diffsurge                                   ║");
console.error("║    npm install -g diffsurge                                ║");
console.error("║    docker run diffsurge/cli surge --help                   ║");
if (hasGo()) {
  console.error("║                                                            ║");
  console.error("║  Go was detected, but source files were not found next to  ║");
  console.error("║  this package, so `go run` fallback is unavailable here.   ║");
}
console.error("╚══════════════════════════════════════════════════════════════╝");
process.exit(1);
