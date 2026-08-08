import { expect, test } from '@playwright/test';

const jobs = [
  { id: 'browser-one', name: 'Browser One', type: 'trending', source: 'tmdb', media: 'movie', limit: 20, enabled: true, rule_set_id: 'default-movies' },
  { id: 'browser-two', name: 'Browser Two', type: 'popular', source: 'tmdb', media: 'movie', limit: 20, enabled: true, rule_set_id: 'default-movies' },
];

test.beforeEach(async ({ request }) => {
  const current = await request.get('/v1/jobs/list');
  expect(current.ok()).toBeTruthy();
  const existing = new Set((await current.json()).map((job) => job.id));
  for (const job of jobs) {
    const response = existing.has(job.id)
      ? await request.put(`/v1/jobs/${job.id}`, { data: job })
      : await request.post('/v1/jobs', { data: job });
    expect(response.ok(), `Could not seed ${job.id}: ${response.status()} ${await response.text()}`).toBeTruthy();
  }
});

test('job edits require an explicit discard and remain isolated', async ({ page }) => {
  const errors = [];
  page.on('console', (message) => { if (message.type() === 'error') errors.push(message.text()); });
  page.on('pageerror', (error) => errors.push(error.message));
  await page.goto('/jobs');

  await page.getByText('Browser One', { exact: true }).click();
  const name = page.locator('#modal-name');
  await name.fill('Unsaved Browser One');
  await expect(page.locator('#job-editor-dirty')).toBeVisible();

  page.once('dialog', async (dialog) => dialog.dismiss());
  await page.getByRole('button', { name: 'Close job editor' }).click();
  await expect(page.locator('#job-modal')).toBeVisible();
  await expect(name).toHaveValue('Unsaved Browser One');

  page.once('dialog', async (dialog) => dialog.accept());
  await page.getByRole('button', { name: 'Close job editor' }).click();
  await expect(page.locator('#job-modal')).toBeHidden();

  await page.getByText('Browser Two', { exact: true }).click();
  await expect(page.locator('#modal-name')).toHaveValue('Browser Two');
  expect(errors).toEqual([]);
});

test('tabs and destructive dialogs are keyboard complete', async ({ page }) => {
  await page.goto('/filters');
  const movie = page.getByRole('tab', { name: 'Movie rules' });
  await movie.focus();
  await page.keyboard.press('ArrowRight');
  await expect(page.getByRole('tab', { name: 'Show rules' })).toBeFocused();
  await page.keyboard.press('End');
  await expect(page.getByRole('tab', { name: 'Title exceptions' })).toBeFocused();

  await page.goto('/activity');
  const open = page.getByRole('button', { name: 'Clear Old Logs' });
  await open.click();
  await expect(page.getByRole('dialog', { name: 'Clear activity history' })).toBeVisible();
  await page.keyboard.press('Escape');
  await expect(page.locator('#clearLogsModal')).toBeHidden();
  await expect(open).toBeFocused();
});

test('failed job saves preserve the user draft', async ({ page }) => {
  await page.route('**/v1/jobs/browser-one', async (route) => {
    if (route.request().method() === 'PUT') {
      await route.fulfill({ status: 500, contentType: 'application/json', body: '{"error":"Test save failure"}' });
    } else {
      await route.continue();
    }
  });
  await page.goto('/jobs');
  await page.getByText('Browser One', { exact: true }).click();
  await page.locator('#modal-name').fill('Draft survives');
  await page.getByRole('button', { name: 'Save Changes' }).click();
  await expect(page.locator('#job-modal')).toBeVisible();
  await expect(page.locator('#modal-name')).toHaveValue('Draft survives');
  await expect(page.getByText('Test save failure')).toBeVisible();
});

test('title-exception media changes require an explicit discard', async ({ page }) => {
  await page.goto('/filters');
  await page.getByRole('tab', { name: 'Title exceptions' }).click();
  await page.getByLabel('Add allowed title ID').fill('27205');
  await page.locator('[data-chip-field="exceptionsAllow"] button').click();

  page.once('dialog', async (dialog) => dialog.dismiss());
  await page.locator('#exceptions-tab-show').click();
  await expect(page.locator('#exceptions-tab-movie')).toHaveAttribute('aria-selected', 'true');
  await expect(page.locator('[data-chip-list="exceptionsAllow"]')).toContainText('27205');

  page.once('dialog', async (dialog) => dialog.accept());
  await page.locator('#exceptions-tab-show').click();
  await expect(page.locator('#exceptions-tab-show')).toHaveAttribute('aria-selected', 'true');
});

test('CSV cells neutralize spreadsheet formulas and follow RFC 4180 quoting', async ({ page }) => {
  await page.goto('/activity');
  const values = await page.evaluate(() => [
    window.encodeCSVCell('=1+1'),
    window.encodeCSVCell('plain'),
    window.encodeCSVCell('a,b'),
    window.encodeCSVCell('say "hi"'),
    window.encodeCSVCell('line\nbreak'),
  ]);
  expect(values).toEqual(["'=1+1", 'plain', '"a,b"', '"say ""hi"""', '"line\nbreak"']);
});
