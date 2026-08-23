#!/usr/bin/env node
'use strict';

const { existsSync } = require('node:fs');
const { resolve, join } = require('node:path');
const { spawnSync } = require('node:child_process');

const binary = [
  process.env.GITU_BINARY,
  join(__dirname, '..', 'native', process.platform === 'win32' ? 'gitu.exe' : 'gitu')
].filter(Boolean).map((candidate) => resolve(candidate)).find(existsSync);

if (!binary) {
  console.error([
    'GitTrackUntracked native binary was not found.',
    'Run npm install again, or install from source with:',
    'go install github.com/MeetChaudhari/GitTrackUntracked/cmd/gitu@latest'
  ].join('\n'));
  process.exitCode = 1;
} else {
  const result = spawnSync(binary, process.argv.slice(2), { stdio: 'inherit' });
  if (result.error) throw result.error;
  process.exitCode = result.status ?? 1;
}
