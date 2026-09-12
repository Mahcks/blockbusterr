import { mkdtemp, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { spawn, spawnSync } from 'node:child_process';

const dataDir = await mkdtemp(join(tmpdir(), 'blockbusterr-browser-'));
await writeFile(join(dataDir, 'config.dev.yaml'), `version: dev
tmdb:
  api_key: browser-test
jobs:
  sync_interval: 24h
  mode: direct
  list: []
`);

const binary = join(dataDir, 'blockbusterr');
const build = spawnSync('go', ['build', '-o', binary, './cmd/app'], { stdio: 'inherit' });
if (build.status !== 0) process.exit(build.status ?? 1);

const child = spawn(binary, [], {
  stdio: 'inherit',
  env: {
    ...process.env,
    CONFIG_PATH: dataDir,
    DATA_DIR: dataDir,
    BLOCKBUSTERR_DRY_RUN: 'true',
  },
});

async function stop(signal) {
  try {
    child.kill('SIGKILL');
  } catch (error) {
    if (error.code !== 'ESRCH') throw error;
  }
  await rm(dataDir, { recursive: true, force: true });
  process.exit(signal === 'SIGINT' ? 130 : 0);
}

process.on('SIGTERM', () => stop('SIGTERM'));
process.on('SIGINT', () => stop('SIGINT'));
child.on('exit', async (code) => {
  await rm(dataDir, { recursive: true, force: true });
  process.exit(code ?? 0);
});
