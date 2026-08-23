#!/usr/bin/env node
'use strict';

const crypto = require('node:crypto');
const fs = require('node:fs');
const https = require('node:https');
const path = require('node:path');

if (process.env.GITU_SKIP_DOWNLOAD === '1' || process.env.GITU_BINARY) process.exit(0);

const packageRoot = path.resolve(__dirname, '..');
const pkg = require(path.join(packageRoot, 'package.json'));
const targets = {
  'darwin-x64': 'darwin_amd64', 'darwin-arm64': 'darwin_arm64',
  'linux-x64': 'linux_amd64', 'linux-arm64': 'linux_arm64',
  'win32-x64': 'windows_amd64', 'win32-arm64': 'windows_arm64'
};
const target = targets[`${process.platform}-${process.arch}`];
if (!target) {
  console.error(`GitTrackUntracked does not yet distribute a binary for ${process.platform}/${process.arch}.`);
  process.exit(1);
}

const extension = process.platform === 'win32' ? '.exe' : '';
const asset = `gitu_${pkg.version}_${target}${extension}`;
const base = `https://github.com/MeetChaudhari/GitTrackUntracked/releases/download/v${pkg.version}`;
const nativeDir = path.join(packageRoot, 'native');
const destination = path.join(nativeDir, `gitu${extension}`);
const temporary = `${destination}.download`;

function request(url) {
  return new Promise((resolve, reject) => {
    https.get(url, { headers: { 'User-Agent': 'gittrackuntracked-npm' } }, (response) => {
      if (response.statusCode >= 300 && response.statusCode < 400 && response.headers.location) {
        response.resume();
        resolve(request(new URL(response.headers.location, url).toString()));
        return;
      }
      if (response.statusCode !== 200) {
        response.resume();
        reject(new Error(`download failed (${response.statusCode}) for ${url}`));
        return;
      }
      resolve(response);
    }).on('error', reject);
  });
}

async function download(url, file) {
  const response = await request(url);
  await new Promise((resolve, reject) => {
    const output = fs.createWriteStream(file, { mode: 0o755 });
    response.pipe(output);
    output.on('finish', () => output.close(resolve));
    output.on('error', reject);
    response.on('error', reject);
  });
}

async function main() {
  fs.mkdirSync(nativeDir, { recursive: true });
  const checksums = await request(`${base}/checksums.txt`);
  const text = await new Promise((resolve, reject) => {
    let value = '';
    checksums.setEncoding('utf8');
    checksums.on('data', (chunk) => { value += chunk; });
    checksums.on('end', () => resolve(value));
    checksums.on('error', reject);
  });
  const expected = text.split('\n').map((line) => line.trim().split(/\s+/)).find((entry) => entry[1] === asset)?.[0];
  if (!expected) throw new Error(`checksum for ${asset} is missing from this release`);
  await download(`${base}/${asset}`, temporary);
  const actual = crypto.createHash('sha256').update(fs.readFileSync(temporary)).digest('hex');
  if (actual !== expected) throw new Error(`checksum mismatch for ${asset}`);
  fs.renameSync(temporary, destination);
  if (process.platform !== 'win32') fs.chmodSync(destination, 0o755);
  console.log(`Installed GitTrackUntracked native binary for ${process.platform}/${process.arch}.`);
}

main().catch((error) => {
  fs.rmSync(temporary, { force: true });
  console.error(`GitTrackUntracked installation failed: ${error.message}`);
  process.exit(1);
});
