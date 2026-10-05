#!/usr/bin/env node
// The whole of the npm wrapper. It finds the binary npm already installed
// for this platform and runs it, unchanged.
//
// There is no download here and no postinstall script. The binary arrives
// as an ordinary dependency, which is why `npm install --ignore-scripts`
// works and why installing needs no network beyond the registry itself
// (ADR-0024).

const { spawnSync } = require("child_process");

const pkg = `forgelore-${process.platform}-${process.arch}`;
const exe = process.platform === "win32" ? "forgelore.exe" : "forgelore";

let binary;
try {
  binary = require.resolve(`${pkg}/${exe}`);
} catch {
  process.stderr.write(
    `forgelore: no binary for ${process.platform}/${process.arch}.\n` +
      `This wrapper ships ${pkg}, which npm did not install. Either the\n` +
      `platform is not one Forgelore builds for, or the install ran with\n` +
      `--no-optional. Installing the binary directly works everywhere:\n` +
      `  https://github.com/forgeprint/forgelore#installing\n`,
  );
  process.exit(1);
}

// stdio is inherited rather than piped: `forgelore mcp` speaks a protocol
// over stdin and stdout, and anything in between would have to be correct
// about framing forever.
const result = spawnSync(binary, process.argv.slice(2), { stdio: "inherit" });
if (result.error) {
  process.stderr.write(`forgelore: ${result.error.message}\n`);
  process.exit(1);
}
// A signal is not an exit code. 128 + signal is what a shell reports.
if (result.signal) {
  process.exit(1);
}
process.exit(result.status === null ? 1 : result.status);
