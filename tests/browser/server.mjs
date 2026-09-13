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

const containerImage = process.env.BLOCKBUSTERR_BROWSER_IMAGE;
const containerName = `blockbusterr-browser-${process.pid}`;
const binary = join(dataDir, 'blockbusterr');
if (!containerImage) {
  const build = spawnSync('go', ['build', '-o', binary, './cmd/app'], { stdio: 'inherit' });
  if (build.status !== 0) process.exit(build.status ?? 1);
}

const child = spawn(containerImage ? 'docker' : binary, containerImage ? [
  'run', '--rm', '--name', containerName, '--network', 'host',
  '--tmpfs', '/app/data', '-e', 'BLOCKBUSTERR_DRY_RUN=true', containerImage,
] : [], {
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
    if (containerImage) spawnSync('docker', ['stop', containerName], { stdio: 'ignore' });
    else child.kill('SIGKILL');
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
