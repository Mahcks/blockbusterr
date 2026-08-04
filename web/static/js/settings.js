// Owns the Settings workspace: connection status, dirty-state tracking,
// save/discard, scoring validation, profile/root-folder loading, import,
// restore, and section navigation. Depends on togglePassword/showNotification
// from app.js.

let baselineSnapshot = null;
let sessionTestResults = {}; // service -> 'connected' | 'failed', session-local only
let restoreFile = null;

const REQUIRED_PAIR = {
  radarr: ['radarr.url', 'radarr.api_key'],
  sonarr: ['sonarr.url', 'sonarr.api_key'],
  jellyseerr: ['jellyseerr.url', 'jellyseerr.api_key'],
  trakt: ['trakt.client_id', 'trakt.client_secret'],
};

const PRESETS = {
  balanced: { rating: 0.6, popularity: 0.3, recency: 0.1 },
  quality: { rating: 0.8, popularity: 0.1, recency: 0.1 },
  trending: { rating: 0.3, popularity: 0.6, recency: 0.1 },
  new: { rating: 0.4, popularity: 0.2, recency: 0.4 },
};

document.addEventListener('DOMContentLoaded', init);

function init() {
  const form = document.getElementById('settings-form');
  if (!form) return;

  baselineSnapshot = serializeForm(form);

  renderServiceStatuses();
  renderModeReadiness();
  updateWeightTotal();
  renderScoringSummary();
  renderLimitsSummary();
  updateSaveBar();

  form.addEventListener('input', onFormChange);
  form.addEventListener('change', onFormChange);

  document.addEventListener('click', handleSettingsClick);
  document.getElementById('scoring-preset')?.addEventListener('change', applyPreset);
  document.getElementById('import-form')?.addEventListener('submit', handleImport);
  document.getElementById('restore-form')?.addEventListener('submit', handleRestoreSubmit);
  document.addEventListener('keydown', handleKeydown);
  window.addEventListener('beforeunload', handleBeforeUnload);

  loadLatestVersion();
  setupNavActiveTracking();
  // Radarr/Sonarr may be slow or unreachable, so the dirty baseline is never
  // gated on their response — each load patches only its own select's value
  // into the existing baseline once it settles, instead of re-snapshotting
  // the whole form (which would race with anything the user typed meanwhile).
  autoLoadRadarrSonarrOptions();
}

function onFormChange() {
  renderServiceStatuses();
  renderModeReadiness();
  updateWeightTotal();
  renderScoringSummary();
  renderLimitsSummary();
  updateSaveBar();
}

function serializeForm(form) {
  const data = new FormData(form);
  // Unchecked checkboxes are absent from FormData entirely, so represent
  // scoring.enabled explicitly rather than relying on its presence/absence.
  const entries = [...data.entries()].filter(([key]) => key !== 'scoring.enabled');
  entries.push(['scoring.enabled', String(document.getElementById('scoring-enabled')?.checked)]);
  entries.sort((a, b) => (a[0] < b[0] ? -1 : a[0] > b[0] ? 1 : 0));
  return JSON.stringify(entries);
}

function isDirty() {
  const form = document.getElementById('settings-form');
  return form ? serializeForm(form) !== baselineSnapshot : false;
}

function fieldValue(form, name) {
  return String(form.elements[name]?.value || '').trim();
}

// ---- Service status -----------------------------------------------------

function serviceState(service) {
  const form = document.getElementById('settings-form');
  const tested = sessionTestResults[service];

  if (service === 'tmdb' || service === 'simkl') {
    const value = fieldValue(form, service === 'tmdb' ? 'tmdb.api_key' : 'simkl.client_id');
    return value ? { state: 'configured', label: 'Configured' } : { state: 'not-configured', label: 'Not configured' };
  }

  const [firstKey, secondKey] = REQUIRED_PAIR[service];
  const first = fieldValue(form, firstKey);
  const second = fieldValue(form, secondKey);
  const optional = service === 'jellyseerr';

  if (!first && !second) {
    return optional
      ? { state: 'not-configured', label: 'Not configured' }
      : { state: 'not-configured', label: 'Not configured' };
  }
  if (!first || !second) {
    return { state: 'needs-attention', label: 'Needs attention' };
  }
  if (tested === 'connected') return { state: 'connected', label: 'Connected' };
  if (tested === 'failed') return { state: 'failed', label: 'Failed' };
  return { state: 'configured', label: 'Configured' };
}

function renderServiceStatuses() {
  document.querySelectorAll('[data-service-status]').forEach((el) => {
    const service = el.dataset.serviceStatus;
    const { state, label } = serviceState(service);
    el.dataset.state = state;
    el.replaceChildren(document.createTextNode(label));
  });
}

// ---- Delivery mode readiness ---------------------------------------------

function renderModeReadiness() {
  const direct = document.querySelector('[data-mode-readiness="direct"]');
  const jelly = document.querySelector('[data-mode-readiness="jellyseerr"]');
  if (direct) {
    const radarr = serviceState('radarr').state === 'configured' || serviceState('radarr').state === 'connected';
    const sonarr = serviceState('sonarr').state === 'configured' || serviceState('sonarr').state === 'connected';
    if (radarr && sonarr) writeReadiness(direct, 'ready', 'Ready — Radarr and Sonarr configured');
    else if (radarr) writeReadiness(direct, 'partial', 'Movies only — Sonarr missing', '#svc-sonarr');
    else if (sonarr) writeReadiness(direct, 'partial', 'Shows only — Radarr missing', '#svc-radarr');
    else writeReadiness(direct, 'blocked', 'Needs Radarr or Sonarr', '#svc-radarr');
  }
  if (jelly) {
    const ready = serviceState('jellyseerr').state === 'configured' || serviceState('jellyseerr').state === 'connected';
    if (ready) writeReadiness(jelly, 'ready', 'Ready — request server configured');
    else writeReadiness(jelly, 'blocked', 'Needs Jellyseerr / Seerr connection', '#svc-jellyseerr');
  }
}

function writeReadiness(el, state, text, anchor) {
  el.dataset.state = state;
  el.replaceChildren();
  const icon = document.createElement('i');
  icon.setAttribute('data-lucide', state === 'ready' ? 'check' : state === 'partial' ? 'alert-triangle' : 'circle-dashed');
  icon.setAttribute('aria-hidden', 'true');
  el.appendChild(icon);
  if (anchor) {
    const link = document.createElement('a');
    link.href = anchor;
    link.textContent = text;
    link.addEventListener('click', (event) => { event.stopPropagation(); openServiceRow(anchor); });
    el.appendChild(link);
  } else {
    el.appendChild(document.createTextNode(text));
  }
  window.renderLucideIcons?.(el);
}

function openServiceRow(anchor) {
  document.querySelector(anchor)?.setAttribute('open', '');
}

// ---- Scoring ---------------------------------------------------------

function weightValues() {
  return {
    rating: Number(document.getElementById('rating-weight')?.value) || 0,
    popularity: Number(document.getElementById('popularity-weight')?.value) || 0,
    recency: Number(document.getElementById('recency-weight')?.value) || 0,
  };
}

function updateWeightTotal() {
  const { rating, popularity, recency } = weightValues();
  const total = rating + popularity + recency;
  const el = document.querySelector('[data-weight-total]');
  const errorEl = document.getElementById('scoring-weight-error');
  const valid = total > 0.999 && total < 1.001;
  if (el) {
    el.dataset.state = valid ? 'valid' : 'invalid';
    el.textContent = `sum ${total.toFixed(2)}`;
  }
  if (errorEl) {
    if (valid) {
      errorEl.classList.add('hidden');
      errorEl.textContent = '';
    } else {
      errorEl.classList.remove('hidden');
      errorEl.textContent = `Weights must total 1.0 — currently ${total.toFixed(2)}.`;
    }
  }
}

function scoringWeightsValid() {
  const { rating, popularity, recency } = weightValues();
  const total = rating + popularity + recency;
  return total > 0.999 && total < 1.001;
}

function renderScoringSummary() {
  const el = document.querySelector('[data-scoring-summary]');
  if (!el) return;
  const enabled = document.getElementById('scoring-enabled')?.checked;
  if (!enabled) {
    el.textContent = 'Disabled';
    return;
  }
  const { rating, popularity, recency } = weightValues();
  el.textContent = `Enabled — ${rating.toFixed(1)}/${popularity.toFixed(1)}/${recency.toFixed(1)}`;
}

function renderLimitsSummary() {
  const el = document.querySelector('[data-limits-summary]');
  if (!el) return;
  const movies = Number(document.getElementById('global-limit-movies')?.value) || 0;
  const shows = Number(document.getElementById('global-limit-shows')?.value) || 0;
  if (!movies && !shows) {
    el.textContent = 'Unlimited';
    return;
  }
  const period = document.getElementById('global-period')?.selectedOptions[0]?.textContent || 'period';
  el.textContent = `${movies || 'Unlimited'} movies, ${shows || 'Unlimited'} shows per ${period.toLowerCase()}`;
}

function applyPreset(event) {
  const key = event.target.value;
  const preset = PRESETS[key];
  if (!preset) return;
  document.getElementById('rating-weight').value = preset.rating;
  document.getElementById('popularity-weight').value = preset.popularity;
  document.getElementById('recency-weight').value = preset.recency;
  event.target.value = '';
  onFormChange();
}

// ---- Save bar ----------------------------------------------------------

function updateSaveBar() {
  const bar = document.getElementById('save-bar');
  if (!bar) return;
  bar.dataset.visible = String(isDirty());
}

function setSaveStatus(text, state) {
  const el = document.querySelector('[data-save-status]');
  if (!el) return;
  el.textContent = text;
  el.dataset.state = state || '';
}

async function saveSettings() {
  const form = document.getElementById('settings-form');
  if (!form) return;

  if (document.getElementById('scoring-enabled')?.checked && !scoringWeightsValid()) {
    setSaveStatus('Fix scoring weights before saving', 'error');
    document.querySelector('details.settings-disclosure')?.setAttribute('open', '');
    document.getElementById('rating-weight')?.focus();
    return;
  }

  const saveButton = document.getElementById('save-button');
  saveButton.disabled = true;
  setSaveStatus('Saving…');

  const formData = new FormData(form);
  formData.set('scoring.enabled', document.getElementById('scoring-enabled')?.checked ? 'true' : 'false');

  try {
    const response = await fetch('/config/save', { method: 'POST', body: formData });
    const result = await response.json().catch(() => ({}));
    if (!response.ok) {
      setSaveStatus(result.error || 'Save failed', 'error');
      window.showNotification?.(result.error || 'Failed to save configuration.', 'error');
      saveButton.disabled = false;
      return;
    }
    baselineSnapshot = serializeForm(form);
    updateSaveBar();
    setSaveStatus('Saved', 'success');
    window.showNotification?.('Configuration saved.', 'success');
    setTimeout(() => setSaveStatus(''), 2500);
  } catch (err) {
    setSaveStatus('Network error', 'error');
    window.showNotification?.(`Failed to save configuration: ${err.message}`, 'error');
    saveButton.disabled = false;
  }
}

function discardChanges() {
  const form = document.getElementById('settings-form');
  if (!form || !baselineSnapshot) return;
  const original = new Map(JSON.parse(baselineSnapshot));
  [...form.elements].forEach((el) => {
    if (!el.name) return;
    if (el.type === 'checkbox') {
      el.checked = original.get(el.name) === 'true';
    } else if (el.type === 'radio') {
      el.checked = original.get(el.name) === el.value;
    } else if (original.has(el.name)) {
      el.value = original.get(el.name);
    }
  });
  onFormChange();
  setSaveStatus('');
}

// ---- Connection testing ---------------------------------------------------

const TEST_ENDPOINTS = {
  trakt: { path: '/v1/trakt/validate', first: 'trakt.client_id', second: 'trakt.client_secret', params: (a, b) => ({ client_id: a, client_secret: b, test: 'true' }) },
  radarr: { path: '/v1/radarr/validate', first: 'radarr.url', second: 'radarr.api_key', params: (a, b) => ({ url: a, api_key: b, test: 'true' }) },
  sonarr: { path: '/v1/sonarr/validate', first: 'sonarr.url', second: 'sonarr.api_key', params: (a, b) => ({ url: a, api_key: b, test: 'true' }) },
  jellyseerr: { path: '/v1/jellyseerr/validate', first: 'jellyseerr.url', second: 'jellyseerr.api_key', params: (a, b) => ({ url: a, api_key: b, test: 'true' }) },
};

async function testConnection(service, button) {
  const config = TEST_ENDPOINTS[service];
  if (!config) return;
  const form = document.getElementById('settings-form');
  const first = fieldValue(form, config.first);
  const second = fieldValue(form, config.second);
  const resultEl = document.querySelector(`[data-test-result="${service}"]`);

  if (!first || (service !== 'trakt' && !second)) {
    writeTestResult(resultEl, 'error', 'Fill in the required fields first.');
    return;
  }

  const originalText = button.textContent;
  button.disabled = true;
  button.textContent = 'Testing…';
  writeTestResult(resultEl, 'pending', 'Testing…');

  try {
    const params = config.params(first, second);
    const response = await fetch(`${config.path}?${new URLSearchParams(params)}`);
    const data = await response.json().catch(() => ({}));
    if (data.connected || (response.ok && data.message)) {
      sessionTestResults[service] = 'connected';
      writeTestResult(resultEl, 'success', data.message || 'Connected.');
    } else {
      sessionTestResults[service] = 'failed';
      writeTestResult(resultEl, 'error', data.error || 'Connection failed.');
    }
  } catch (err) {
    sessionTestResults[service] = 'failed';
    writeTestResult(resultEl, 'error', `Network error: ${err.message}`);
  } finally {
    button.disabled = false;
    button.textContent = originalText;
    renderServiceStatuses();
    renderModeReadiness();
  }
}

function writeTestResult(el, state, message) {
  if (!el) return;
  el.dataset.state = state;
  const icon = document.createElement('i');
  icon.setAttribute('data-lucide', state === 'success' ? 'check' : state === 'error' ? 'x' : 'loader-2');
  icon.setAttribute('aria-hidden', 'true');
  el.replaceChildren(icon, document.createTextNode(message));
  window.renderLucideIcons?.(el);
}

// ---- Radarr / Sonarr profile + root folder loading ------------------------

function formatBytes(bytes) {
  if (!bytes) return '0 Bytes';
  const k = 1024;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${Math.round((bytes / Math.pow(k, i)) * 100) / 100} ${sizes[i]}`;
}

async function loadOptions(service, kind, silent) {
  const selectId = `${service}-${kind === 'profiles' ? 'quality-profile' : 'root-folder'}`;
  const select = document.getElementById(selectId);
  if (!select) return;
  const savedValue = select.dataset.selected;
  const path = kind === 'profiles' ? `/v1/${service}/quality-profiles` : `/v1/${service}/root-folders`;

  try {
    const response = await fetch(path);
    const data = await response.json();
    const items = data.data || [];
    select.replaceChildren(new Option(kind === 'profiles' ? 'Select a quality profile…' : 'Select a root folder…', ''));
    if (items.length === 0) {
      if (!silent) window.showNotification?.(`No ${kind} found. Test the connection first.`, 'error');
      return;
    }
    items.forEach((item) => {
      const option = kind === 'profiles'
        ? new Option(item.name, String(item.id))
        : new Option(`${item.path} (${item.freeSpace ? formatBytes(item.freeSpace) : 'unknown'} free)`, item.path);
      if (savedValue && String(kind === 'profiles' ? item.id : item.path) === String(savedValue)) option.selected = true;
      select.appendChild(option);
    });
    if (silent) refreshBaselineField(select.name, select.value);
    else window.showNotification?.(`${kind === 'profiles' ? 'Quality profiles' : 'Root folders'} loaded.`, 'success');
  } catch (err) {
    if (!silent) window.showNotification?.(`Failed to load ${kind}: ${err.message}`, 'error');
  }
}

// Patches a single field's value into the stored baseline without
// re-serializing the whole form, so an auto-selected option (the saved
// profile/folder becoming available) doesn't register as an unsaved edit —
// while anything the user typed in the meantime is left alone.
function refreshBaselineField(name, value) {
  if (!baselineSnapshot || !name) return;
  const entries = JSON.parse(baselineSnapshot);
  const match = entries.find(([key]) => key === name);
  if (match) match[1] = value;
  else entries.push([name, value]);
  baselineSnapshot = JSON.stringify(entries);
  updateSaveBar();
}

function autoLoadRadarrSonarrOptions() {
  const form = document.getElementById('settings-form');
  const pending = [];
  if (fieldValue(form, 'radarr.url')) {
    pending.push(loadOptions('radarr', 'profiles', true), loadOptions('radarr', 'folders', true));
  }
  if (fieldValue(form, 'sonarr.url')) {
    pending.push(loadOptions('sonarr', 'profiles', true), loadOptions('sonarr', 'folders', true));
  }
  return Promise.all(pending);
}

// ---- Import / restore ---------------------------------------------------

async function handleImport(event) {
  event.preventDefault();
  const fileInput = document.getElementById('config-file');
  const messageEl = document.getElementById('import-message');
  const button = event.target.querySelector('button[type="submit"]');
  if (!fileInput.files?.length) return showFieldMessage(messageEl, 'Select a configuration file first.');

  const formData = new FormData();
  formData.append('config', fileInput.files[0]);
  button.disabled = true;
  const originalText = button.textContent;
  button.textContent = 'Importing…';

  try {
    const response = await fetch('/config/import', { method: 'POST', body: formData });
    const result = await response.json().catch(() => ({}));
    if (response.ok) {
      window.showNotification?.(result.message || 'Configuration imported.', 'success');
      setTimeout(() => window.location.reload(), 1200);
    } else {
      showFieldMessage(messageEl, result.error || 'Failed to import configuration.');
      button.disabled = false;
      button.textContent = originalText;
    }
  } catch (err) {
    showFieldMessage(messageEl, `Network error: ${err.message}`);
    button.disabled = false;
    button.textContent = originalText;
  }
}

function showFieldMessage(el, message) {
  if (!el) return;
  el.textContent = message;
  el.classList.remove('hidden');
}

function handleRestoreSubmit(event) {
  event.preventDefault();
  const fileInput = document.getElementById('restore-file');
  const messageEl = document.getElementById('restore-message');
  if (!fileInput.files?.length) return showFieldMessage(messageEl, 'Select a backup file first.');
  restoreFile = fileInput.files[0];
  document.getElementById('restore-confirm-filename').textContent = restoreFile.name;
  document.getElementById('restore-confirm').classList.remove('hidden');
}

async function confirmRestore() {
  const dialog = document.getElementById('restore-confirm');
  const messageEl = document.getElementById('restore-message');
  const button = dialog.querySelector('[data-action="confirm-restore"]');
  if (!restoreFile) return;

  button.disabled = true;
  const originalText = button.textContent;
  button.textContent = 'Restoring…';

  const formData = new FormData();
  formData.append('config', restoreFile);

  try {
    const response = await fetch('/config/restore', { method: 'POST', body: formData });
    const result = await response.json().catch(() => ({}));
    if (response.ok) {
      window.showNotification?.(result.message || 'Configuration restored.', 'success');
      setTimeout(() => window.location.reload(), 1200);
    } else {
      dialog.classList.add('hidden');
      showFieldMessage(messageEl, result.error || 'Failed to restore configuration.');
    }
  } catch (err) {
    dialog.classList.add('hidden');
    showFieldMessage(messageEl, `Network error: ${err.message}`);
  } finally {
    button.disabled = false;
    button.textContent = originalText;
    restoreFile = null;
  }
}

// ---- Navigation, keyboard, and unload guard --------------------------------

function setupNavActiveTracking() {
  const links = [...document.querySelectorAll('.settings-nav a')];
  const sections = links
    .map((link) => document.querySelector(link.getAttribute('href')))
    .filter(Boolean);
  if (!sections.length) return;

  const setActive = (id) => links.forEach((link) => link.setAttribute('aria-current', String(link.getAttribute('href') === `#${id}`)));
  setActive(sections[0].id);

  const observer = new IntersectionObserver((entries) => {
    const visible = entries.filter((entry) => entry.isIntersecting).sort((a, b) => a.boundingClientRect.top - b.boundingClientRect.top);
    if (visible[0]) setActive(visible[0].target.id);
  }, { rootMargin: '-15% 0px -70% 0px' });

  sections.forEach((section) => observer.observe(section));
}

function handleKeydown(event) {
  if ((event.key === 's' || event.key === 'S') && (event.metaKey || event.ctrlKey)) {
    event.preventDefault();
    if (isDirty()) saveSettings();
    return;
  }
  if (event.key === 'Escape') {
    document.getElementById('restore-confirm')?.classList.add('hidden');
  }
}

function handleBeforeUnload(event) {
  if (!isDirty()) return;
  event.preventDefault();
  event.returnValue = '';
}

// ---- Event delegation ----------------------------------------------------

function handleSettingsClick(event) {
  const target = event.target.closest('[data-action]');
  if (!target) return;
  const action = target.dataset.action;

  if (action === 'toggle-password') {
    const input = document.getElementById(target.dataset.target);
    if (input) window.togglePassword?.(target.dataset.target, target);
    return;
  }
  if (action === 'test-connection') {
    testConnection(target.dataset.service, target);
    return;
  }
  if (action === 'reload-options') {
    loadOptions(target.dataset.service, target.dataset.kind, false);
    return;
  }
  if (action === 'save') {
    saveSettings();
    return;
  }
  if (action === 'discard') {
    discardChanges();
    return;
  }
  if (action === 'cancel-restore') {
    document.getElementById('restore-confirm').classList.add('hidden');
    restoreFile = null;
    return;
  }
  if (action === 'confirm-restore') {
    confirmRestore();
    return;
  }
}

// ---- About: latest version ------------------------------------------------

function loadLatestVersion() {
  const el = document.getElementById('latest-version');
  if (!el) return;
  fetch('https://api.github.com/repos/mahcks/blockbusterr/releases/latest')
    .then((response) => response.json())
    .then((release) => {
      el.textContent = release.tag_name || 'unavailable';
      if (release.html_url) el.href = release.html_url;
    })
    .catch(() => { el.textContent = 'unavailable'; });
}
