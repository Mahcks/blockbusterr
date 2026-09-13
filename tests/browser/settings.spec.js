import { expect, test } from '@playwright/test';
import http from 'node:http';

async function restore(request, tmdb = '') {
  const response = await request.post('/config/restore', { multipart: { config: {
    name: 'fixture.yaml', mimeType: 'application/yaml',
    buffer: Buffer.from(`version: v2.0.1\ntmdb:\n  api_key: "${tmdb}"\njobs:\n  sync_interval: 24h\n  mode: direct\n  list: []\n`),
  } } });
  expect(response.ok(), await response.text()).toBeTruthy();
}

test('fresh settings connect, save twice, and retain hidden credentials and defaults', async ({ page, request }) => {
  let unavailable = false;
  const upstream = http.createServer((req, res) => {
    res.setHeader('Content-Type', 'application/json');
    if (req.headers['x-api-key'] !== 'browser-secret') {
      res.writeHead(401); res.end('{}'); return;
    }
    if (unavailable) { res.writeHead(503); res.end('{}'); return; }
    res.end(JSON.stringify(req.url.includes('qualityprofile')
      ? [{ id: 7, name: 'HD' }]
      : req.url.includes('rootfolder') ? [{ id: 1, path: '/media', freeSpace: 1000 }]
      : { version: '1.0.0' }));
  });
  await new Promise(resolve => upstream.listen(0, '127.0.0.1', resolve));
  const url = `http://127.0.0.1:${upstream.address().port}`;
  const errors = [];
  page.on('pageerror', error => errors.push(error.message));
  try {
    await restore(request);
    await page.goto('/config');
    await expect(page.locator('[data-service-status="tmdb"]')).toHaveText('Not configured');
    await page.locator('#svc-tmdb').evaluate(el => { el.open = true; });
    await page.locator('#tmdb-api-key').fill('browser-tmdb');
    for (const service of ['radarr', 'sonarr', 'jellyseerr']) {
      await page.locator(`#svc-${service}`).evaluate(el => { el.open = true; });
      await page.locator(`#${service}-url`).fill(url);
      await page.locator(`#${service}-api-key`).fill('browser-secret');
      await page.locator(`[data-action="test-connection"][data-service="${service}"]`).click();
      await expect(page.locator(`[data-test-result="${service}"]`)).toHaveAttribute('data-state', 'success');
      if (service !== 'jellyseerr') {
        await page.locator(`[data-action="reload-options"][data-service="${service}"][data-kind="profiles"]`).click();
        await expect(page.locator(`#${service}-quality-profile option[value="7"]`)).toHaveCount(1);
        await page.locator(`#${service}-quality-profile`).selectOption('7');
        await page.locator(`[data-action="reload-options"][data-service="${service}"][data-kind="folders"]`).click();
        await expect(page.locator(`#${service}-root-folder option[value="/media"]`)).toHaveCount(1);
        await page.locator(`#${service}-root-folder`).selectOption('/media');
      }
    }
    await page.locator('#save-button').click();
    await expect(page.locator('#save-bar')).toHaveAttribute('data-visible', 'false');
    await page.locator('#radarr-monitor').selectOption('none');
    await expect(page.locator('#save-button')).toBeEnabled();
    await page.locator('#save-button').click();
    await expect(page.locator('#save-bar')).toHaveAttribute('data-visible', 'false');
    // Keep edits made after submission dirty, and recover from a failed save.
    let releaseSave;
    let saveStarted;
    const heldSave = new Promise(resolve => { releaseSave = resolve; });
    const started = new Promise(resolve => { saveStarted = resolve; });
    await page.route('**/config/save', async route => {
      const response = await route.fetch();
      saveStarted();
      await heldSave;
      await route.fulfill({ response });
    }, { times: 1 });
    await page.locator('#radarr-monitor').selectOption('movieOnly');
    await page.locator('#save-button').click();
    await started;
    await page.locator('#radarr-monitor').selectOption('none');
    releaseSave();
    await expect(page.locator('#save-button')).toBeEnabled();
    await expect(page.locator('#save-bar')).toHaveAttribute('data-visible', 'true');
    await page.route('**/config/save', route => route.fulfill({
      status: 500, contentType: 'application/json', body: '{"error":"Fixture save failure"}',
    }), { times: 1 });
    await page.locator('#save-button').click();
    await expect(page.locator('[data-save-status]')).toHaveText('Fixture save failure');
    await expect(page.locator('#save-button')).toBeEnabled();
    await expect(page.locator('#radarr-monitor')).toHaveValue('none');
    await page.locator('#save-button').click();
    await expect(page.locator('#save-bar')).toHaveAttribute('data-visible', 'false');
    await page.goto('/jobs');
    await page.goto('/config');
    await expect(page.locator('[data-service-status="tmdb"]')).toHaveText('Configured');
    await expect(page.locator('#tmdb-api-key')).toHaveValue('');
    for (const service of ['radarr', 'sonarr', 'jellyseerr']) {
      await expect(page.locator(`[data-service-status="${service}"]`)).toHaveText('Configured');
      await expect(page.locator(`#${service}-api-key`)).toHaveValue('');
      await page.locator(`#svc-${service}`).evaluate(el => { el.open = true; });
      await page.locator(`[data-action="test-connection"][data-service="${service}"]`).click();
      await expect(page.locator(`[data-test-result="${service}"]`)).toHaveAttribute('data-state', 'success');
      if (service !== 'jellyseerr') {
        await expect(page.locator(`#${service}-quality-profile`)).toHaveValue('7');
        await expect(page.locator(`#${service}-root-folder`)).toHaveValue('/media');
      }
    }
    await expect(page.locator('#radarr-monitor')).toHaveValue('none');
    // An outage must not erase saved defaults during an unrelated save.
    unavailable = true;
    await page.reload();
    await page.locator('#svc-radarr').evaluate(el => { el.open = true; });
    await page.locator('[data-action="reload-options"][data-service="radarr"][data-kind="folders"]').click();
    await expect(page.getByText(/Failed to load folders:/)).toBeVisible();
    await expect(page.locator('#radarr-root-folder')).toHaveValue('/media');
    await page.locator('#radarr-monitor').selectOption('movieOnly');
    await page.locator('#save-button').click();
    await expect(page.locator('#save-bar')).toHaveAttribute('data-visible', 'false');
    await page.reload();
    await expect(page.locator('#radarr-root-folder')).toHaveValue('/media');
    await expect(page.locator('#sonarr-root-folder')).toHaveValue('/media');
    expect(errors).toEqual([]);
  } finally {
    await restore(request, 'browser-test');
    upstream.closeAllConnections();
    await new Promise(resolve => upstream.close(resolve));
  }
});
