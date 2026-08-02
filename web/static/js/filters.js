// Owns the Rules workspace: movie/show rule sets and universal title
// exceptions. Rule sets are reusable, media-specific discovery policies
// (config.RuleSet on the backend); a job references exactly one by ID.
// Title exceptions are flat ID lists that bypass rule-set evaluation.

let rulesView = 'movie';
let exceptionsMedia = 'movie';
let ruleSets = [];
let jobsList = [];
let exceptions = { allowed_movie_tmdb_ids: [], blocked_movie_tmdb_ids: [], allowed_show_tvdb_ids: [], blocked_show_tvdb_ids: [] };
let selectedRule = null;
let ruleSnapshot = null; // last-saved serialization of the open rule set, for dirty checks
let exceptionsSnapshot = null;
let ruleSearchQuery = '';
let requestedRuleID = new URLSearchParams(window.location.search).get('ruleSet');
const chips = new Map(); // field key -> { getValues, setValues }

document.addEventListener('DOMContentLoaded', init);

function init() {
  buildChipFields();
  loadAll();

  document.querySelectorAll('[data-rules-view]').forEach((button) => button.addEventListener('click', () => selectView(button.dataset.rulesView)));
  document.querySelectorAll('[data-exceptions-media]').forEach((button) => button.addEventListener('click', () => selectExceptionsMedia(button.dataset.exceptionsMedia)));

  document.getElementById('new-rule-set')?.addEventListener('click', openCreateRuleSet);
  document.getElementById('cancel-rule-set')?.addEventListener('click', closeCreateRuleSet);
  document.getElementById('create-rule-set-form')?.addEventListener('submit', createRuleSet);
  document.getElementById('rule-set-form')?.addEventListener('submit', saveRuleSet);
  document.getElementById('duplicate-rule-set')?.addEventListener('click', duplicateRuleSet);
  document.getElementById('delete-rule-set')?.addEventListener('click', deleteRuleSet);
  document.getElementById('toggle-rule-usage')?.addEventListener('click', toggleRuleUsage);
  document.getElementById('save-exceptions')?.addEventListener('click', saveExceptions);

  document.getElementById('rule-set-search')?.addEventListener('input', (event) => {
    ruleSearchQuery = String(event.target.value || '').trim().toLowerCase();
    renderRuleList();
  });

  const ruleForm = document.getElementById('rule-set-form');
  ruleForm?.addEventListener('input', () => updateRuleDirtyState());
  ruleForm?.addEventListener('change', () => updateRuleDirtyState());

  document.addEventListener('keydown', handleGlobalKeydown);
}

function handleGlobalKeydown(event) {
  if (!(event.key === 's' || event.key === 'S') || !(event.metaKey || event.ctrlKey)) return;
  event.preventDefault();
  if (rulesView === 'exceptions') {
    if (!document.getElementById('save-exceptions').disabled) saveExceptions();
  } else if (!document.getElementById('save-rule-set').disabled) {
    document.getElementById('rule-set-form').requestSubmit();
  }
}

async function loadAll() {
  try {
    const [rulesResponse, exceptionsResponse, jobsResponse] = await Promise.all([
      fetch('/v1/rule-sets'),
      fetch('/v1/title-exceptions'),
      fetch('/v1/jobs/list'),
    ]);
    if (!rulesResponse.ok || !exceptionsResponse.ok) throw new Error('Could not load rules.');
    ruleSets = (await rulesResponse.json()).map((item) => ({ ...item.rule_set, usage_count: item.usage_count }));
    jobsList = jobsResponse.ok ? await jobsResponse.json() : [];
    exceptions = await exceptionsResponse.json();
    populateCopyFromOptions();
    renderRuleList(requestedRuleID);
    requestedRuleID = null;
    exceptionsSnapshot = JSON.stringify(normalizeExceptions(exceptions));
    writeExceptions();
  } catch (err) {
    window.showNotification?.(err.message || 'Could not load rules.', 'error');
  }
}

function jobsUsing(ruleSetId) {
  return jobsList.filter((job) => (job.rule_set_id || `default-${job.media === 'show' ? 'shows' : 'movies'}`) === ruleSetId);
}

// ---- View switching -------------------------------------------------

function selectView(view) {
  if (view === rulesView) return;
  if (rulesView !== 'exceptions' && isRuleDirty() && !confirm('Discard unsaved changes to this rule set?')) return;
  rulesView = view;
  document.querySelectorAll('[data-rules-view]').forEach((button) => button.setAttribute('aria-selected', String(button.dataset.rulesView === view)));
  document.getElementById('rule-set-workspace').classList.toggle('hidden', view === 'exceptions');
  document.getElementById('exceptions-workspace').classList.toggle('hidden', view !== 'exceptions');
  document.getElementById('new-rule-set').classList.toggle('hidden', view === 'exceptions');
  closeCreateRuleSet();
  if (view !== 'exceptions') renderRuleList();
}

function selectExceptionsMedia(media) {
  exceptionsMedia = media;
  document.querySelectorAll('[data-exceptions-media]').forEach((button) => button.setAttribute('aria-selected', String(button.dataset.exceptionsMedia === media)));
  document.getElementById('exceptions-id-kind').textContent = media === 'show' ? 'TVDB' : 'TMDB';
  writeExceptions();
}

// ---- Rule set list ----------------------------------------------------

function renderRuleList(preferredID) {
  const compatible = ruleSets.filter((rule) => rule.media === rulesView);
  const filtered = ruleSearchQuery ? compatible.filter((rule) => rule.name.toLowerCase().includes(ruleSearchQuery)) : compatible;

  selectedRule = compatible.find((rule) => rule.id === preferredID)
    || compatible.find((rule) => rule.id === selectedRule?.id)
    || filtered[0]
    || compatible[0];

  const list = document.getElementById('rule-set-list');
  list.replaceChildren(...filtered.map((rule) => {
    const button = document.createElement('button');
    button.type = 'button';
    button.className = 'rule-set-item';
    button.setAttribute('role', 'listitem');
    button.setAttribute('aria-current', String(rule.id === selectedRule?.id));

    const nameRow = document.createElement('span');
    nameRow.className = 'rule-set-item-name';
    nameRow.appendChild(document.createTextNode(rule.name));
    if (isProtectedRuleSet(rule.id)) {
      const badge = document.createElement('span');
      badge.className = 'rule-set-item-default';
      badge.textContent = 'Default';
      nameRow.appendChild(badge);
    }

    const meta = document.createElement('small');
    meta.textContent = `${rule.usage_count} job${rule.usage_count === 1 ? '' : 's'}`;

    button.append(nameRow, meta);
    button.addEventListener('click', () => {
      if (rule.id === selectedRule?.id) return;
      if (isRuleDirty() && !confirm(`Discard unsaved changes to "${selectedRule.name}"?`)) return;
      selectedRule = rule;
      renderRuleList(rule.id);
    });
    return button;
  }));

  document.getElementById('rule-set-empty').classList.toggle('hidden', filtered.length > 0 || !ruleSearchQuery);
  document.getElementById('rule-set-form').classList.toggle('hidden', !selectedRule);
  if (selectedRule) writeRule(selectedRule);
  window.lucide?.createIcons();
}

function isProtectedRuleSet(id) {
  return id === 'default-movies' || id === 'default-shows';
}

// ---- Rule set editor ----------------------------------------------------

function writeRule(rule) {
  const values = rule[rule.media === 'show' ? 'shows' : 'movies'] || {};
  document.getElementById('rule-name').value = rule.name;
  document.getElementById('rule-media-badge').textContent = rule.media === 'show' ? 'Show' : 'Movie';
  document.getElementById('rule-media-badge').className = `job-type-badge ${rule.media === 'show' ? 'job-type-badge-show' : 'job-type-badge-movie'}`;
  document.getElementById('rule-revision').textContent = `Revision ${rule.revision}`;

  chips.get('countries').setValues(values.allowed_countries || []);
  chips.get('languages').setValues(values.allowed_languages || []);
  chips.get('genres').setValues(values.blacklisted_genres || []);
  chips.get('keywords').setValues(values.blacklisted_keywords || []);
  chips.get('networks').setValues(values.blacklisted_networks || []);
  chips.get('blockedIds').setValues((rule.media === 'show' ? values.blacklisted_tvdb_ids : values.blacklisted_tmdb_ids) || []);

  document.getElementById('rule-networks-field').classList.toggle('hidden', rule.media !== 'show');
  document.getElementById('rule-blocked-ids-label').textContent = rule.media === 'show' ? 'Blocked TVDB IDs' : 'Blocked TMDB IDs';

  set('rule-min-year', values.blacklisted_min_year);
  set('rule-max-year', values.blacklisted_max_year);
  set('rule-min-runtime', values.blacklisted_min_runtime);
  set('rule-max-runtime', values.blacklisted_max_runtime);
  set('rule-min-rating', values.min_rating);
  set('rule-min-votes', values.min_votes);

  const usage = jobsUsing(rule.id);
  document.getElementById('rule-usage-text').textContent = describeUsage(rule, usage.length);
  const toggle = document.getElementById('toggle-rule-usage');
  toggle.classList.toggle('hidden', usage.length === 0);
  toggle.setAttribute('aria-expanded', 'false');
  toggle.textContent = 'Show jobs';
  renderRuleUsageJobs(usage, false);

  document.getElementById('delete-rule-set').classList.toggle('hidden', isProtectedRuleSet(rule.id));
  clearFieldError('rule-name-error');
  clearFieldError('rule-boundary-error');

  ruleSnapshot = JSON.stringify(readRuleForm());
  updateRuleDirtyState();
}

function describeUsage(rule, count) {
  if (isProtectedRuleSet(rule.id)) return count === 1 ? 'Built-in default, used by 1 job.' : `Built-in default, used by ${count} jobs.`;
  if (count === 0) return 'Not assigned to any job yet.';
  if (count === 1) return 'Job-specific rules for 1 job.';
  return `Shared by ${count} jobs. Saving changes affects all of them.`;
}

function renderRuleUsageJobs(usage, expanded) {
  const list = document.getElementById('rule-usage-jobs');
  list.classList.toggle('hidden', !expanded);
  if (!expanded) return;
  list.replaceChildren(...usage.map((job) => {
    const item = document.createElement('li');
    const link = document.createElement('a');
    link.href = `/jobs?job=${encodeURIComponent(job.id)}`;
    link.textContent = job.name || job.id;
    item.appendChild(link);
    return item;
  }));
}

function toggleRuleUsage() {
  const toggle = document.getElementById('toggle-rule-usage');
  const expanded = toggle.getAttribute('aria-expanded') === 'true';
  toggle.setAttribute('aria-expanded', String(!expanded));
  toggle.textContent = expanded ? 'Show jobs' : 'Hide jobs';
  renderRuleUsageJobs(jobsUsing(selectedRule.id), !expanded);
}

function set(id, value) {
  const el = document.getElementById(id);
  el.value = value ? String(value) : '';
}

function num(id) {
  const value = Number(document.getElementById(id).value);
  return Number.isFinite(value) && value > 0 ? value : 0;
}

function readRuleForm() {
  const media = selectedRule?.media || rulesView;
  const values = {
    allowed_countries: chips.get('countries').getValues(),
    allowed_languages: chips.get('languages').getValues(),
    blacklisted_genres: chips.get('genres').getValues(),
    blacklisted_keywords: chips.get('keywords').getValues(),
    blacklisted_min_year: num('rule-min-year'),
    blacklisted_max_year: num('rule-max-year'),
    blacklisted_min_runtime: num('rule-min-runtime'),
    blacklisted_max_runtime: num('rule-max-runtime'),
    min_rating: Number(document.getElementById('rule-min-rating').value) || 0,
    min_votes: num('rule-min-votes'),
  };
  const blockedIds = chips.get('blockedIds').getValues();
  if (media === 'show') {
    values.blacklisted_networks = chips.get('networks').getValues();
    values.blacklisted_tvdb_ids = blockedIds;
  } else {
    values.blacklisted_tmdb_ids = blockedIds;
  }
  return { name: document.getElementById('rule-name').value.trim(), values };
}

function isRuleDirty() {
  if (!selectedRule || rulesView === 'exceptions' || ruleSnapshot === null) return false;
  return JSON.stringify(readRuleForm()) !== ruleSnapshot;
}

function updateRuleDirtyState() {
  const dirty = isRuleDirty();
  document.getElementById('save-rule-set').disabled = !dirty;
  document.getElementById('rule-set-dirty-hint').classList.toggle('hidden', !dirty);
  const status = document.getElementById('rule-set-save-status');
  if (dirty && status.dataset.state !== 'dirty') {
    status.textContent = '';
    status.dataset.state = 'dirty';
  }

  // Patch the sidebar dot in place rather than re-rendering the whole list
  // on every keystroke.
  const nameEl = document.querySelector('.rule-set-item[aria-current="true"] .rule-set-item-name');
  if (!nameEl) return;
  let dot = nameEl.querySelector('.rule-set-dirty-dot');
  if (dirty && !dot) {
    dot = document.createElement('i');
    dot.setAttribute('data-lucide', 'circle');
    dot.className = 'rule-set-dirty-dot';
    dot.setAttribute('aria-hidden', 'true');
    nameEl.prepend(dot);
    window.lucide?.createIcons();
  } else if (!dirty && dot) {
    dot.remove();
  }
}

async function saveRuleSet(event) {
  event.preventDefault();
  if (!selectedRule) return;
  const form = readRuleForm();
  const payload = { id: selectedRule.id, name: form.name, media: selectedRule.media, revision: selectedRule.revision };
  payload[selectedRule.media === 'show' ? 'shows' : 'movies'] = form.values;

  const button = document.getElementById('save-rule-set');
  button.disabled = true;
  const status = document.getElementById('rule-set-save-status');
  status.textContent = 'Saving…';
  status.dataset.state = 'saving';

  const response = await fetch(`/v1/rule-sets/${encodeURIComponent(selectedRule.id)}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(payload),
  });

  if (!response.ok) {
    const message = await responseError(response);
    applyFieldError(message);
    status.textContent = '';
    status.dataset.state = '';
    updateRuleDirtyState();
    window.showNotification?.(message, 'error');
    return;
  }

  clearFieldError('rule-name-error');
  clearFieldError('rule-boundary-error');
  status.textContent = 'Saved';
  status.dataset.state = 'saved';
  window.showNotification?.(`"${form.name}" saved.`, 'success');
  await loadAll();
  renderRuleList(selectedRule?.id || null);
  setTimeout(() => { if (document.getElementById('rule-set-save-status').dataset.state === 'saved') document.getElementById('rule-set-save-status').textContent = ''; }, 2500);
}

function applyFieldError(message) {
  const lower = message.toLowerCase();
  if (lower.includes('name') || lower.includes('named')) {
    showFieldError('rule-name-error', message);
  } else if (lower.includes('year') || lower.includes('runtime') || lower.includes('rating') || lower.includes('negative')) {
    showFieldError('rule-boundary-error', message);
  }
}

function showFieldError(id, message) {
  const el = document.getElementById(id);
  el.textContent = message;
  el.classList.remove('hidden');
}

function clearFieldError(id) {
  const el = document.getElementById(id);
  el.textContent = '';
  el.classList.add('hidden');
}

// ---- Create / duplicate / delete ----------------------------------------

function populateCopyFromOptions() {
  const select = document.getElementById('new-rule-copy-from');
  const compatible = ruleSets.filter((rule) => rule.media === rulesView);
  select.replaceChildren(...[
    new Option('Empty rule set', ''),
    ...compatible.map((rule) => new Option(rule.name, rule.id)),
  ]);
}

function openCreateRuleSet() {
  populateCopyFromOptions();
  const form = document.getElementById('create-rule-set-form');
  form.classList.remove('hidden');
  document.getElementById('new-rule-set').disabled = true;
  document.getElementById('new-rule-name').focus();
}

function closeCreateRuleSet() {
  const form = document.getElementById('create-rule-set-form');
  form.reset();
  form.classList.add('hidden');
  document.getElementById('new-rule-set').disabled = false;
}

async function createRuleSet(event) {
  event.preventDefault();
  const name = document.getElementById('new-rule-name').value.trim();
  if (!name) return;
  const copyFromID = document.getElementById('new-rule-copy-from').value;
  const copyFrom = copyFromID ? ruleSets.find((rule) => rule.id === copyFromID) : null;

  const payload = { name, media: rulesView };
  const key = rulesView === 'show' ? 'shows' : 'movies';
  payload[key] = copyFrom ? { ...copyFrom[key] } : {};

  const response = await fetch('/v1/rule-sets', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) });
  if (!response.ok) return window.showNotification?.(await responseError(response), 'error');
  const created = await response.json();
  closeCreateRuleSet();
  window.showNotification?.(`"${created.name}" created.`, 'success');
  await loadAll();
  renderRuleList(created.id);
}

async function duplicateRuleSet() {
  if (!selectedRule) return;
  const payload = structuredClone(selectedRule);
  delete payload.id;
  delete payload.usage_count;
  payload.name = `${payload.name} Copy`;
  const response = await fetch('/v1/rule-sets', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) });
  if (!response.ok) return window.showNotification?.(await responseError(response), 'error');
  const created = await response.json();
  window.showNotification?.(`Duplicated as "${created.name}".`, 'success');
  await loadAll();
  renderRuleList(created.id);
}

async function deleteRuleSet() {
  if (!selectedRule) return;
  const usage = jobsUsing(selectedRule.id);
  if (usage.length > 0) {
    window.showNotification?.(`"${selectedRule.name}" is used by ${usage.length} job${usage.length === 1 ? '' : 's'}. Reassign ${usage.length === 1 ? 'it' : 'them'} first.`, 'error');
    return;
  }
  if (!confirm(`Delete "${selectedRule.name}"? This cannot be undone.`)) return;
  const response = await fetch(`/v1/rule-sets/${encodeURIComponent(selectedRule.id)}`, { method: 'DELETE' });
  if (!response.ok) return window.showNotification?.(await responseError(response), 'error');
  window.showNotification?.(`"${selectedRule.name}" deleted.`, 'success');
  selectedRule = null;
  await loadAll();
}

// ---- Title exceptions -----------------------------------------------------

function normalizeExceptions(values) {
  return {
    allowed_movie_tmdb_ids: [...(values.allowed_movie_tmdb_ids || [])].sort((a, b) => a - b),
    blocked_movie_tmdb_ids: [...(values.blocked_movie_tmdb_ids || [])].sort((a, b) => a - b),
    allowed_show_tvdb_ids: [...(values.allowed_show_tvdb_ids || [])].sort((a, b) => a - b),
    blocked_show_tvdb_ids: [...(values.blocked_show_tvdb_ids || [])].sort((a, b) => a - b),
  };
}

function writeExceptions() {
  const allowKey = exceptionsMedia === 'show' ? 'allowed_show_tvdb_ids' : 'allowed_movie_tmdb_ids';
  const blockKey = exceptionsMedia === 'show' ? 'blocked_show_tvdb_ids' : 'blocked_movie_tmdb_ids';
  chips.get('exceptionsAllow').setValues(exceptions[allowKey] || []);
  chips.get('exceptionsBlock').setValues(exceptions[blockKey] || []);
  updateExceptionsDirtyState();
}

function readExceptionsForm() {
  const next = normalizeExceptions(exceptions);
  const allowKey = exceptionsMedia === 'show' ? 'allowed_show_tvdb_ids' : 'allowed_movie_tmdb_ids';
  const blockKey = exceptionsMedia === 'show' ? 'blocked_show_tvdb_ids' : 'blocked_movie_tmdb_ids';
  next[allowKey] = [...chips.get('exceptionsAllow').getValues()].sort((a, b) => a - b);
  next[blockKey] = [...chips.get('exceptionsBlock').getValues()].sort((a, b) => a - b);
  return next;
}

function isExceptionsDirty() {
  return JSON.stringify(readExceptionsForm()) !== exceptionsSnapshot;
}

function updateExceptionsDirtyState() {
  const dirty = isExceptionsDirty();
  document.getElementById('save-exceptions').disabled = !dirty;
  document.getElementById('exceptions-dirty-hint').classList.toggle('hidden', !dirty);
}

async function saveExceptions() {
  const payload = readExceptionsForm();
  const response = await fetch('/v1/title-exceptions', { method: 'PUT', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(payload) });
  if (!response.ok) return window.showNotification?.(await responseError(response), 'error');
  exceptions = await response.json();
  exceptionsSnapshot = JSON.stringify(normalizeExceptions(exceptions));
  updateExceptionsDirtyState();
  window.showNotification?.('Title exceptions saved.', 'success');
}

// ---- Chip / tag input component --------------------------------------

// Renders removable "chips" for a list-valued field, backed by a plain
// array. Shared by rule-set list fields and title-exception ID fields.
function buildChipFields() {
  registerChipField('countries', { normalize: (v) => v.trim().toUpperCase(), emptyText: 'Any country' });
  registerChipField('languages', { normalize: (v) => v.trim().toLowerCase(), emptyText: 'Any language' });
  registerChipField('genres', { normalize: (v) => v.trim(), emptyText: 'None blocked' });
  registerChipField('keywords', { normalize: (v) => v.trim(), emptyText: 'None blocked' });
  registerChipField('networks', { normalize: (v) => v.trim(), emptyText: 'None blocked' });
  registerChipField('blockedIds', { normalize: normalizeID, validate: validateID, emptyText: 'None blocked', numeric: true });
  registerChipField('exceptionsAllow', { normalize: normalizeID, validate: validateID, emptyText: 'No allowed titles', numeric: true });
  registerChipField('exceptionsBlock', { normalize: normalizeID, validate: validateID, emptyText: 'No blocked titles', numeric: true });

  ['countries', 'languages', 'genres', 'keywords', 'networks', 'blockedIds'].forEach((key) => chips.get(key)?.onChange(() => updateRuleDirtyState()));
  ['exceptionsAllow', 'exceptionsBlock'].forEach((key) => chips.get(key)?.onChange(() => updateExceptionsDirtyState()));

  document.getElementById('rule-name')?.addEventListener('input', () => clearFieldError('rule-name-error'));
}

function normalizeID(raw) {
  return Number(String(raw).trim());
}

function validateID(raw) {
  const value = Number(String(raw).trim());
  if (!Number.isInteger(value) || value <= 0) return 'must be a positive whole number';
  return null;
}

function registerChipField(key, { normalize, validate, emptyText, numeric }) {
  const entry = document.querySelector(`[data-chip-field="${key}"]`);
  const list = document.querySelector(`[data-chip-list="${key}"]`);
  if (!entry || !list) return;
  const input = entry.querySelector('input');
  const addButton = entry.querySelector('button');
  let items = [];
  let onChangeCb = null;

  function render() {
    list.replaceChildren();
    if (items.length === 0) {
      const empty = document.createElement('span');
      empty.className = 'filter-token-empty';
      empty.textContent = emptyText;
      list.appendChild(empty);
    } else {
      items.forEach((value, index) => {
        const chip = document.createElement('span');
        chip.className = 'filter-token';
        const label = document.createElement('span');
        label.textContent = value;
        const remove = document.createElement('button');
        remove.type = 'button';
        remove.setAttribute('aria-label', `Remove ${value}`);
        remove.textContent = '×';
        remove.addEventListener('click', () => {
          items.splice(index, 1);
          render();
          onChangeCb?.();
        });
        chip.append(label, remove);
        list.appendChild(chip);
      });
    }
  }

  function addFromInput() {
    const tokens = input.value.split(',').map((token) => token.trim()).filter(Boolean);
    if (tokens.length === 0) return;
    const rejected = [];
    tokens.forEach((token) => {
      const value = normalize(token);
      const error = validate?.(token);
      if (error) { rejected.push(`"${token}" ${error}`); return; }
      const exists = items.some((item) => (numeric ? item === value : String(item).toLowerCase() === String(value).toLowerCase()));
      if (!exists) items.push(value);
    });
    input.value = '';
    render();
    onChangeCb?.();
    if (rejected.length) window.showNotification?.(`Skipped ${rejected.join(', ')}.`, 'error');
  }

  input.addEventListener('keydown', (event) => {
    if (event.key === 'Enter') { event.preventDefault(); addFromInput(); }
    else if (event.key === 'Backspace' && !input.value && items.length) { items.pop(); render(); onChangeCb?.(); }
  });
  input.addEventListener('blur', () => { if (input.value.trim()) addFromInput(); });
  addButton.addEventListener('click', addFromInput);

  render();
  chips.set(key, {
    getValues: () => [...items],
    setValues: (next) => { items = [...(next || [])]; render(); },
    onChange: (cb) => { onChangeCb = cb; },
  });
}

// ---- Shared helpers ----------------------------------------------------

async function responseError(response) {
  try { return (await response.json()).error || 'Request failed.'; } catch { return 'Request failed.'; }
}
