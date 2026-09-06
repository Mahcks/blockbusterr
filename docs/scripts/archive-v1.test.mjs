import assert from 'node:assert/strict';
import fs from 'node:fs';
import os from 'node:os';
import path from 'node:path';
import { execFileSync } from 'node:child_process';

const directory = fs.mkdtempSync(path.join(os.tmpdir(), 'blockbusterr-docs-'));
try {
 const file = path.join(directory, 'index.html');
 fs.writeFileSync(file, '<a href="/concepts/jobs/#schedule">Jobs</a><a href="/v1/api/">API</a><a href="https://blockbusterr.dev/">Stable</a> ghcr.io/mahcks/blockbusterr:latest');
 execFileSync(process.execPath, [new URL('./archive-v1.mjs', import.meta.url).pathname, directory]);
 assert.equal(fs.readFileSync(file, 'utf8'), '<a href="/v1/concepts/jobs/#schedule">Jobs</a><a href="/v1/api/">API</a><a href="https://blockbusterr.dev/">Stable</a> ghcr.io/mahcks/blockbusterr:v1.5.2');
} finally {
 fs.rmSync(directory, { recursive: true, force: true });
}
