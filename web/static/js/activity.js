let currentSort = { field: 'timestamp', direction: 'desc' };
let currentPage = 1;
let autoRefreshInterval = null;
let searchTimeout = null;
let deliveryChart = null;
let decisionChart = null;
let recentRuns = [];
let jobsByID = new Map();
let activityViewMode = 'items';
let hasActivityEntries = null;
let hasJobRuns = null;

function updateActivityEmptyState() {
  if (hasActivityEntries === null || hasJobRuns === null) return;
  document.getElementById('activityEmptyState')?.classList.toggle('hidden', hasActivityEntries || hasJobRuns);
}

// Initialize
document.addEventListener('DOMContentLoaded', function() {
  setupAutoRefresh();
  loadActivityChart();
  loadJobRuns();
  loadActivityJobs();
  loadActivityLanguages();
  setActivityView('items');
  const timelineSearch = document.getElementById('timelineRunSearchInput');
  const timelineStatus = document.getElementById('timelineRunStatusFilter');
  if (timelineSearch) timelineSearch.addEventListener('input', renderActivityTimeline);
  if (timelineStatus) timelineStatus.addEventListener('change', renderActivityTimeline);
  document.getElementById('searchInput')?.addEventListener('input', debounceSearch);
  document.getElementById('autoRefresh')?.addEventListener('change', toggleAutoRefresh);
  document.getElementById('clearLogsDays')?.addEventListener('change', updateClearLogsConfirmation);
  document.getElementById('clearLogsConfirmation')?.addEventListener('input', updateClearLogsConfirmation);
  ['statusFilter', 'mediaFilter', 'jobFilter', 'languageFilter', 'dateRangeFilter'].forEach((id) => {
    document.getElementById(id)?.addEventListener('change', applyFilters);
  });
  document.getElementById('pageSizeSelect')?.addEventListener('change', changePageSize);
  document.addEventListener('keydown', (event) => {
    if (event.key === 'Escape') closeClearLogsModal();
  });
});

document.addEventListener('click', function(event) {
  const trigger = event.target.closest('[data-action]');
  if (!trigger) return;

  const actions = {
    'export-csv': exportToCSV,
    'show-clear-logs': showClearLogsModal,
    'close-clear-logs': closeClearLogsModal,
    'clear-logs': clearOldLogs,
    'refresh': applyFilters,
    'toggle-auto-refresh': toggleAutoRefresh
  };
  const action = trigger.dataset.action;
  if (actions[action]) return actions[action]();
  if (action === 'set-activity-view') return setActivityView(trigger.dataset.view);
  if (action === 'filter-status') return quickFilterStatus(trigger.dataset.status || '');
  if (action === 'filter-job') return quickFilterJob(trigger.dataset.jobId);
  if (action === 'sort') return sortBy(trigger.dataset.sort);
  if (action === 'toggle-timeline-run') return toggleTimelineRun(Number(trigger.dataset.runId));
  if (action === 'filter-run-entries') return filterRunEntries(trigger);
  if (action === 'go-to-page') return goToPage(Number(trigger.dataset.page));
  if (action === 'toggle-details') return toggleDetails(trigger, Number(trigger.dataset.index));
  if (action === 'toggle-history') return toggleHistory(trigger, Number(trigger.dataset.index));
});

async function loadActivityLanguages() {
  const select = document.getElementById('languageFilter');
  if (!select) return;
  try {
    const response = await fetch('/v1/activity/languages');
    if (!response.ok) return;
    const data = await response.json();
    (data.languages || []).forEach((language) => select.add(new Option(language.toUpperCase(), language)));
  } catch (_) {
    // Language filtering is optional; the remaining Activity view stays usable.
  }
}

async function loadActivityJobs() {
  const select = document.getElementById('jobFilter');
  if (!select) return;
  try {
    const response = await fetch('/v1/jobs/enabled');
    if (!response.ok) return;
    const jobs = await response.json();
    jobs.sort((a, b) => a.name.localeCompare(b.name));
    jobs.forEach((job) => select.add(new Option(job.name, job.id)));
  } catch (_) {
    // Activity remains usable when job metadata cannot be loaded.
  }
}

function setupAutoRefresh() {
  const enabled = document.getElementById('autoRefresh').checked;
  if (enabled && !autoRefreshInterval) {
    autoRefreshInterval = setInterval(() => {
      applyFilters();
      htmx.trigger('#stats', 'statsUpdate');
      loadJobRuns();
    }, 30000); // 30 seconds
  } else if (!enabled && autoRefreshInterval) {
    clearInterval(autoRefreshInterval);
    autoRefreshInterval = null;
  }
}

function toggleAutoRefresh() {
  setupAutoRefresh();
}

function goToPage(page) {
  currentPage = page;
  applyFilters();
  // Scroll to top of activity table
  document.getElementById('activityTable').scrollIntoView({ behavior: 'smooth', block: 'start' });
}

function changePageSize() {
  currentPage = 1; // Reset to first page when changing page size
  applyFilters();
}

function setActivityView(view) {
  activityViewMode = view === 'timeline' ? 'timeline' : 'items';
  const itemsBtn = document.getElementById('activityViewItemsBtn');
  const timelineBtn = document.getElementById('activityViewTimelineBtn');
  const itemsPanel = document.getElementById('activityItemsPanel');
  const timelinePanel = document.getElementById('activityTimelinePanel');

  if (!itemsBtn || !timelineBtn || !itemsPanel || !timelinePanel) return;

  const timelineSelected = activityViewMode === 'timeline';
  itemsBtn.setAttribute('aria-selected', String(!timelineSelected));
  timelineBtn.setAttribute('aria-selected', String(timelineSelected));
  if (timelineSelected) {
    itemsPanel.classList.add('hidden');
    timelinePanel.classList.remove('hidden');
    renderActivityTimeline();
  } else {
    timelinePanel.classList.add('hidden');
    itemsPanel.classList.remove('hidden');
  }
}

function formatDuration(ms) {
  const n = Number(ms || 0);
  if (n < 1000) return `${n}ms`;
  const totalSec = Math.floor(n / 1000);
  const h = Math.floor(totalSec / 3600);
  const m = Math.floor((totalSec % 3600) / 60);
  const s = totalSec % 60;
  const parts = [];
  if (h > 0) parts.push(`${h}h`);
  if (m > 0) parts.push(`${m}m`);
  if (s > 0 || parts.length === 0) parts.push(`${s}s`);
  return parts.join(' ');
}

function sortRunsNewestFirst(runs) {
  return [...(runs || [])].sort((a, b) => {
    const aTime = a && a.started_at ? new Date(a.started_at).getTime() : 0;
    const bTime = b && b.started_at ? new Date(b.started_at).getTime() : 0;
    if (aTime !== bTime) return bTime - aTime;
    return Number((b && b.id) || 0) - Number((a && a.id) || 0);
  });
}

function statusClassForRun(status) {
  return status === 'failed' ? 'run-status-failed' : status === 'running' ? 'run-status-running' : 'run-status-completed';
}

function sourceClass(source) {
  return source === 'simkl' ? 'run-source-simkl' : source === 'tmdb' ? 'run-source-tmdb' : source === 'trakt' ? 'run-source-trakt' : '';
}

function escapeHTML(value) {
  return String(value ?? '').replace(/[&<>"']/g, (character) => ({
    '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;'
  })[character]);
}

function renderRunOutcome(run) {
  const found = Number(run.total_found || 0);
  const passed = Number(run.passed_filters || 0);
  const rejected = Number(run.rejected || 0);
  const outcomes = [['Added', run.added, 'added'], ['Requested', run.requested, 'requested'], ['Rejected', rejected, 'rejected'], ['Skipped', run.skipped, 'skipped'], ['Failed', run.failed, 'failed']];

  if (found === 0) {
    return `<div class="run-flow run-flow-empty" aria-label="Run decision flow">
      <i data-lucide="inbox" class="h-4 w-4" aria-hidden="true"></i>
      Nothing was discovered in this run.
    </div>`;
  }

  const percent = (value) => found ? Math.round((value / found) * 100) : 0;
  const node = (label, value, tone, ratio = percent(Number(value || 0))) => {
    const count = Number(value || 0);
    return `<span class="run-flow-node run-flow-${tone}${count === 0 ? ' run-flow-node-empty' : ''}"><small>${label}</small><b>${count}</b><em>${ratio}%</em></span>`;
  };
  const passedOutcomes = outcomes.filter(([, value, tone]) => tone !== 'rejected' && Number(value || 0) > 0);

  // The tree above is already the single source of truth for every count and
  // percentage; this strip is purely a proportion-at-a-glance visual, so it
  // carries no text of its own (only a title tooltip per segment) and no
  // separate legend, which would just restate the same five numbers again.
  const distributionLabel = outcomes
    .filter(([, value]) => Number(value || 0) > 0)
    .map(([label, value]) => `${label} ${Number(value)} (${percent(Number(value))}%)`)
    .join(', ');
  const distributionTrack = outcomes.map(([label, value, tone]) => {
    const count = Number(value || 0);
    if (count <= 0) return '';
    const pct = percent(count);
    return `<div class="run-distribution-seg run-distribution-${tone}" style="width:${pct}%" title="${label}: ${count} (${pct}%)"></div>`;
  }).join('');

  return `<div class="run-flow" aria-label="Run decision flow">
    <div class="run-flow-tree">
      ${node('Found', found, 'found', 100)}
      <span class="run-flow-line" aria-hidden="true"></span>
      <span class="run-flow-branches">
        <span class="run-flow-path">${node('Rejected', rejected, 'rejected')}</span>
        <span class="run-flow-path run-flow-passed-path">
          ${node('Passed filters', passed, 'passed')}
          <span class="run-flow-line" aria-hidden="true"></span>
          <span class="run-flow-outcomes">${passedOutcomes.map(([label, value, tone]) => node(label, value, tone)).join('')}</span>
        </span>
      </span>
    </div>
    <div class="run-distribution">
      <div class="run-distribution-track" role="img" aria-label="Outcome distribution: ${distributionLabel}">${distributionTrack}</div>
    </div>
  </div>`;
}

function filterTimelineRuns() {
  const search = ((document.getElementById('timelineRunSearchInput') || {}).value || '').trim().toLowerCase();
  const status = ((document.getElementById('timelineRunStatusFilter') || {}).value || '').trim();
  return recentRuns.filter((run) => {
    const jobName = String(run.job_name || run.job_id || '').toLowerCase();
    const statusMatch = !status || run.status === status;
    const searchMatch = !search || jobName.includes(search) || String(run.id).includes(search);
    return statusMatch && searchMatch;
  });
}

function renderTimelineRunRow(run) {
  const status = run.status || 'unknown';
  const jobName = run.job_name || run.job_id || 'Unnamed Job';
  const runMode = run.mode || '-';
  const media = run.media_type || '-';
  const job = jobsByID.get(run.job_id) || {};
  const source = String(job.source || '');
  const jobType = String(job.type || '');
  const outcome = (label, value, tone) => {
    const count = Number(value || 0);
    if (tone !== 'all' && count === 0) return '';
    return `<button type="button" data-action="filter-run-entries" data-run-id="${run.id}" data-status="${tone}" class="run-entry-filter run-entry-filter-${tone}" aria-pressed="${tone === 'all'}" ${count === 0 ? 'data-empty="true"' : ''}><span>${label}</span><b>${count}</b></button>`;
  };
  const resultChip = (label, value, tone) => {
    const count = Number(value || 0);
    return count > 0 ? `<span class="run-row-${tone}">${count} ${label}</span>` : '';
  };
  return `
    <article class="run-timeline-row" data-status="${escapeHTML(status)}">
      <button data-action="toggle-timeline-run" data-run-id="${run.id}" class="run-timeline-trigger" aria-expanded="false">
        <time class="run-time">${run.started_at ? new Date(run.started_at).toLocaleString([], { month: 'short', day: 'numeric', hour: 'numeric', minute: '2-digit' }) : '-'}</time>
        <span class="run-timeline-main">
          <span class="run-title"><b>${escapeHTML(jobName)}</b>${source ? `<span class="run-source ${sourceClass(source)}">${escapeHTML(source)}</span>` : ''}<span>${escapeHTML(jobType || 'Job')} · ${escapeHTML(media)} · ${escapeHTML(runMode)} · Run #${Number(run.id || 0)}</span></span>
          <span class="run-row-results">
            <span>${Number(run.total_found || 0)} found</span>
            <span>${Number(run.passed_filters || 0)} passed</span>
            ${resultChip('added', run.added, 'added')}
            ${resultChip('requested', run.requested, 'requested')}
            ${resultChip('skipped', run.skipped, 'skipped')}
            ${resultChip('rejected', run.rejected, 'rejected')}
            ${resultChip('failed', run.failed, 'failed')}
          </span>
        </span>
        <span class="run-summary"><span class="run-status ${statusClassForRun(status)}">${escapeHTML(status)}</span><span>${formatDuration(run.duration_ms || 0)}</span></span>
        <i id="timeline-run-chevron-${run.id}" data-lucide="chevron-down" class="h-4 w-4" aria-hidden="true"></i>
      </button>
      <div id="timeline-run-logs-${run.id}" class="run-timeline-detail hidden">
        <div class="run-timeline-detail-title">Decision flow</div>
        ${renderRunOutcome(run)}
        <div class="run-timeline-detail-title run-entries-title">Activity Entries <span>Filter by outcome</span></div>
        <div class="run-entry-filters" data-run-filters="${run.id}">
          ${outcome('All', run.total_found, 'all')}
          ${outcome('Added', run.added, 'added')}
          ${outcome('Requested', run.requested, 'requested')}
          ${outcome('Skipped', run.skipped, 'skipped')}
          ${outcome('Rejected', run.rejected, 'rejected')}
          ${outcome('Failed', run.failed, 'failed')}
        </div>
        <div id="timeline-run-logs-body-${run.id}">
          <div class="p-3 text-xs text-slate-500">Loading entries…</div>
        </div>
      </div>
    </article>
  `;
}

function filterRunEntries(trigger) {
  const runID = trigger.dataset.runId;
  const status = trigger.dataset.status;
  const body = document.getElementById(`timeline-run-logs-body-${runID}`);
  if (!body) return;
  document.querySelectorAll(`[data-run-filters="${runID}"] .run-entry-filter`).forEach((button) => button.setAttribute('aria-pressed', String(button === trigger)));

  const rows = body.querySelectorAll('.activity-ledger-row');
  let visible = 0;
  rows.forEach((row) => {
    const matches = status === 'all' || row.dataset.status === status;
    row.classList.toggle('hidden', !matches);
    if (matches) visible++;
  });

  // Any explicit outcome filter (including "All") now drives visibility by
  // status match directly, superseding the initial progressive-disclosure cap.
  body.querySelector('.activity-show-more')?.remove();

  let emptyEl = body.querySelector('.run-entries-empty');
  if (rows.length > 0 && visible === 0) {
    if (!emptyEl) {
      emptyEl = document.createElement('div');
      emptyEl.className = 'run-entries-empty';
      body.querySelector('.activity-ledger')?.after(emptyEl);
    }
    emptyEl.textContent = `No ${status} entries in this run.`;
  } else if (emptyEl) {
    emptyEl.remove();
  }
}

function collapseLongEntryList(container) {
  const THRESHOLD = 20;
  const rows = Array.from(container.querySelectorAll('.activity-ledger-row'));
  if (rows.length <= THRESHOLD) return;

  const ledger = container.querySelector('.activity-ledger');
  if (!ledger) return;

  rows.slice(THRESHOLD).forEach((row) => {
    row.classList.add('activity-row-more', 'hidden');
  });

  const remaining = rows.length - THRESHOLD;
  const btn = document.createElement('button');
  btn.type = 'button';
  btn.className = 'activity-show-more';
  btn.textContent = `Show ${remaining} more ${remaining === 1 ? 'entry' : 'entries'}`;
  btn.addEventListener('click', () => {
    container.querySelectorAll('.activity-row-more').forEach((row) => row.classList.remove('hidden'));
    btn.remove();
  });
  ledger.after(btn);
}

function renderActivityTimeline() {
  const container = document.getElementById('activityTimelineList');
  if (!container) return;
  const runs = filterTimelineRuns();

  // Auto-refresh re-renders this list from scratch; without preserving
  // expanded state and scroll position, every refresh silently collapses
  // whatever run the user was inspecting and yanks them back to the top.
  const expandedRunIds = Array.from(container.querySelectorAll('.run-timeline-detail:not(.hidden)'))
    .map((el) => Number(el.id.replace('timeline-run-logs-', '')))
    .filter((id) => Number.isFinite(id));
  const scrollY = window.scrollY;

  if (!runs.length) {
    container.innerHTML = '<div class="px-4 py-10 text-center text-sm text-slate-500">No Job Runs match these filters.</div>';
    return;
  }
  container.innerHTML = runs.map(renderTimelineRunRow).join('');
  window.lucide?.createIcons();

  expandedRunIds.forEach((runId) => {
    if (document.getElementById(`timeline-run-logs-${runId}`)) {
      toggleTimelineRun(runId);
    }
  });

  window.scrollTo(0, scrollY);
}

async function toggleTimelineRun(runID) {
  const panel = document.getElementById(`timeline-run-logs-${runID}`);
  const body = document.getElementById(`timeline-run-logs-body-${runID}`);
  const chevron = document.getElementById(`timeline-run-chevron-${runID}`);
  if (!panel || !body) return;
  const trigger = panel.previousElementSibling;

  const hidden = panel.classList.contains('hidden');
  if (!hidden) {
    panel.classList.add('hidden');
    trigger?.setAttribute('aria-expanded', 'false');
    if (chevron) chevron.setAttribute('data-lucide', 'chevron-down');
    window.lucide?.createIcons();
    return;
  }

  panel.classList.remove('hidden');
  trigger?.setAttribute('aria-expanded', 'true');
  if (chevron) chevron.setAttribute('data-lucide', 'chevron-up');
  window.lucide?.createIcons();
  if (panel.dataset.loaded === 'true') return;

  const url = `/v1/activity/logs?page=1&pageSize=200&run_id=${runID}&dedupe=false`;
  try {
    if (window.htmx) {
      await htmx.ajax('GET', url, {
        target: `#timeline-run-logs-body-${runID}`,
        swap: 'innerHTML',
        headers: { 'HX-Request': 'true' }
      });
    } else {
      const resp = await fetch(url, { headers: { 'HX-Request': 'true' } });
      body.innerHTML = await resp.text();
    }
    parseFilterDetailsIn(body);
    collapseLongEntryList(body);
    panel.dataset.loaded = 'true';
  } catch (err) {
    body.textContent = `Failed to load logs: ${err}`;
    body.className = 'p-3 text-xs text-red-400';
  }
}

function parseFilterDetailsIn(container) {
  container.querySelectorAll('.filter-details-content').forEach(function(el) {
    const filterDetails = el.getAttribute('data-filter-details');
    if (!filterDetails) return;
    try {
      const checks = JSON.parse(filterDetails);
      if (!Array.isArray(checks)) return;
      let html = '';
      checks.forEach(check => {
        const icon = check.passed
          ? '<svg class="w-3.5 h-3.5 text-green-400" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-7.25 7.25a1 1 0 01-1.414 0l-3.25-3.25a1 1 0 011.414-1.414l2.543 2.543 6.543-6.543a1 1 0 011.414 0z" clip-rule="evenodd"/></svg>'
          : '<svg class="w-3.5 h-3.5 text-red-400" fill="currentColor" viewBox="0 0 20 20"><path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM7.293 7.293a1 1 0 011.414 0L10 8.586l1.293-1.293a1 1 0 111.414 1.414L11.414 10l1.293 1.293a1 1 0 01-1.414 1.414L10 11.414l-1.293 1.293a1 1 0 01-1.414-1.414L8.586 10 7.293 8.707a1 1 0 010-1.414z" clip-rule="evenodd"/></svg>';
        const state = check.passed ? 'filter-check-passed' : 'filter-check-failed';
        html += `<div class="filter-check ${state}"><span class="filter-check-icon">${icon}</span><span><strong>${escapeHTML(check.name)}</strong> ${escapeHTML(check.message)}</span></div>`;
      });
      el.innerHTML = html;
    } catch (_err) {
      el.innerHTML = `<div class="text-xs text-slate-400 whitespace-pre-wrap break-all">${escapeHTML(filterDetails)}</div>`;
    }
  });
}

function applyFilters() {
  const pageSize = document.getElementById('pageSizeSelect').value;
  const status = document.getElementById('statusFilter').value;
  const media = document.getElementById('mediaFilter').value;
  const job = document.getElementById('jobFilter').value;
	  const language = document.getElementById('languageFilter').value;
  const dateRange = document.getElementById('dateRangeFilter').value;
  const search = document.getElementById('searchInput').value;
  
  let url = `/v1/activity/logs?page=${currentPage}&pageSize=${pageSize}`;
  if (status) url += `&status=${status}`;
  if (media) url += `&media=${media}`;
  if (job) url += `&job=${encodeURIComponent(job)}`;
  if (language) url += `&language=${encodeURIComponent(language)}`;
  if (dateRange) url += `&date_range=${dateRange}`;
  if (search) url += `&search=${encodeURIComponent(search)}`;
  if (currentSort.field) url += `&sort=${currentSort.field}&order=${currentSort.direction}`;
  
  htmx.ajax('GET', url, {target: '#activityTable'});
}

function debounceSearch() {
  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => {
    applyFilters();
  }, 500);
}

function sortBy(field) {
  if (currentSort.field === field) {
    currentSort.direction = currentSort.direction === 'desc' ? 'asc' : 'desc';
  } else {
    currentSort.field = field;
    currentSort.direction = 'desc';
  }
  
  // Update UI
  document.querySelectorAll('[id^="sort-"]').forEach(btn => {
    btn.classList.remove('bg-slate-700', 'text-blue-400');
  });
  const btn = document.getElementById(`sort-${field}`);
  btn.classList.add('bg-slate-700', 'text-blue-400');
  btn.textContent = btn.textContent.split(' ')[0] + (currentSort.direction === 'desc' ? ' ↓' : ' ↑');
  
  applyFilters();
}


function quickFilterStatus(status) {
  // Update the status filter dropdown
  document.getElementById('statusFilter').value = status;
  // Update button styles
  document.querySelectorAll('[id^="filter-"]').forEach(btn => btn.setAttribute('aria-pressed', 'false'));
  const btnId = status ? `filter-${status}` : 'filter-all';
  document.getElementById(btnId).setAttribute('aria-pressed', 'true');
  applyFilters();
}

// Allow filtering by job from activity_table.html
function quickFilterJob(jobID) {
  const jobFilter = document.getElementById('jobFilter');
  if (jobFilter) {
    jobFilter.value = jobID;
    applyFilters();
  }
}

function toggleDetails(triggerOrIndex, maybeIndex) {
  let element = null;
  let trigger = null;
  if (typeof triggerOrIndex === 'object' && triggerOrIndex !== null) {
    trigger = triggerOrIndex;
    const index = maybeIndex;
    const card = trigger.closest('[id^="activity-card-"]');
    if (card) element = card.querySelector('#details-' + index);
  } else {
    const index = triggerOrIndex;
    element = document.getElementById('details-' + index);
  }
  if (!element) return;
  const nowHidden = element.classList.toggle('hidden');
  if (trigger) trigger.setAttribute('aria-expanded', String(!nowHidden));
}

function toggleHistory(triggerOrIndex, maybeIndex) {
  let element = null;
  if (typeof triggerOrIndex === 'object' && triggerOrIndex !== null) {
    const trigger = triggerOrIndex;
    const index = maybeIndex;
    const card = trigger.closest('[id^="activity-card-"]');
    if (card) element = card.querySelector('#history-' + index);
  } else {
    const index = triggerOrIndex;
    element = document.getElementById('history-' + index);
  }
  if (element) element.classList.toggle('hidden');
}

function showClearLogsModal() {
  const modal = document.getElementById('clearLogsModal');
  modal.classList.remove('hidden');
  updateClearLogsConfirmation();
  modal.querySelector('select')?.focus();
}

function closeClearLogsModal() {
  document.getElementById('clearLogsModal').classList.add('hidden');
  document.getElementById('clearLogsConfirmation').value = '';
  document.getElementById('clearDeliveryMemory').checked = false;
  updateClearLogsConfirmation();
}

function refreshActivityData() {
  if (typeof applyFilters === 'function') {
    applyFilters();
  }
  if (window.htmx) {
    htmx.trigger('#stats', 'statsUpdate');
  }
  if (typeof loadJobRuns === 'function') {
    loadJobRuns();
  }
}

async function clearOldLogs() {
  const days = document.getElementById('clearLogsDays').value;
  const clearAll = days === 'all';
  const confirmation = document.getElementById('clearLogsConfirmation').value;
  if (clearAll && confirmation !== 'CLEAR') return;
  try {
    const clearDeliveryMemory = document.getElementById('clearDeliveryMemory')?.checked === true;
    const query = clearAll ? `scope=all&confirm=CLEAR&clear_delivery_memory=${clearDeliveryMemory}` : `days=${encodeURIComponent(days)}`;
    const response = await fetch(`/v1/activity/logs?${query}`, {
      method: 'DELETE'
    });
    if (!response.ok) throw new Error(`Request failed (${response.status})`);
    const data = await response.json();
    closeClearLogsModal();
    window.showNotification(clearAll ? 'Cleared completed activity history' : `Cleared ${data.count} old activity entries and completed runs`, 'success');
    applyFilters();
    htmx.trigger('#stats', 'statsUpdate');
    loadJobRuns();
  } catch (err) {
    window.showNotification(`Failed to clear logs: ${err.message}`, 'error');
  }
}

function updateClearLogsConfirmation() {
  const clearAll = document.getElementById('clearLogsDays')?.value === 'all';
  const confirmation = document.getElementById('clearLogsConfirmation');
  document.getElementById('clearAllConfirmation')?.classList.toggle('hidden', !clearAll);
  document.getElementById('clearLogsButton').disabled = clearAll && confirmation?.value !== 'CLEAR';
}

async function exportToCSV() {
  try {
		const language = document.getElementById('languageFilter').value;
		const response = await fetch(`/v1/activity/logs?limit=10000&dedupe=false&format=json${language ? `&language=${encodeURIComponent(language)}` : ''}`);
    if (!response.ok) throw new Error(`Request failed (${response.status})`);
    const payload = await response.json();
    let logs = [];

    if (Array.isArray(payload)) {
      logs = payload;
    } else if (payload && Array.isArray(payload.logs)) {
      logs = payload.logs.map(entry => entry && (entry.Log || entry.log) ? (entry.Log || entry.log) : entry);
    }

    if (!logs || logs.length === 0) {
      window.showNotification('No activity entries to export', 'info');
      return;
    }

		const headers = ['Timestamp', 'Job Type', 'Source', 'Media Type', 'Title', 'Language', 'Year', 'Score', 'Rank', 'Status', 'Message', 'TMDB ID', 'IMDB ID', 'TVDB ID'];
    const csvRows = [headers.join(',')];
    
    logs.forEach(log => {
      const row = [
        log.timestamp,
        log.job_type,
		log.source || '',
        log.media_type,
        `"${(log.title || '').replace(/"/g, '""')}"`,
			log.language || '',
        log.year || '',
        log.score || '',
        log.rank || '',
        log.status,
        `"${(log.message || '').replace(/"/g, '""')}"`,
        log.tmdb_id || '',
        log.imdb_id || '',
        log.tvdb_id || ''
      ];
      csvRows.push(row.join(','));
    });

    const csvContent = csvRows.join('\n');
    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `blockbusterr-activity-${new Date().toISOString().split('T')[0]}.csv`;
    a.click();
    window.URL.revokeObjectURL(url);
  } catch (err) {
    window.showNotification(`Failed to export activity: ${err.message}`, 'error');
  }
}

async function loadActivityChart() {
  try {
    const response = await fetch('/v1/activity/chart');
    const data = await response.json();
    
    deliveryChart?.destroy();
    decisionChart?.destroy();

    const chartOptions = {
      responsive: true,
      maintainAspectRatio: false,
      interaction: { intersect: false, mode: 'index' },
      plugins: { legend: { labels: { color: 'rgb(148, 163, 184)', boxWidth: 10, boxHeight: 10 } } },
      scales: {
        y: { beginAtZero: true, stacked: true, ticks: { color: 'rgb(148, 163, 184)', precision: 0 }, grid: { color: 'rgba(148, 163, 184, 0.1)' } },
        x: { stacked: true, ticks: { color: 'rgb(148, 163, 184)' }, grid: { display: false } }
      }
    };

    const deliveryDatasets = [
      { label: 'Added', data: data.added || [], backgroundColor: 'rgb(34, 197, 94)', borderRadius: 2 },
      { label: 'Requested', data: data.requested || [], backgroundColor: 'rgb(99, 102, 241)', borderRadius: 2 }
    ];
    if (data.dry_run) {
      deliveryDatasets.push(
        { label: 'Would add', data: data.would_add || [], backgroundColor: 'rgba(34, 197, 94, 0.4)', borderRadius: 2 },
        { label: 'Would request', data: data.would_request || [], backgroundColor: 'rgba(99, 102, 241, 0.45)', borderRadius: 2 }
      );
      document.getElementById('deliveryChartDescription').textContent = 'Real outcomes and dry-run simulations are shown separately.';
    } else {
      document.getElementById('deliveryChartDescription').textContent = 'Added directly or requested through Jellyseerr/Seerr.';
    }

    deliveryChart = new Chart(document.getElementById('deliveryChart'), {
      type: 'bar',
      data: {
        labels: data.labels || [],
        datasets: deliveryDatasets
      },
      options: chartOptions
    });

    decisionChart = new Chart(document.getElementById('decisionChart'), {
      type: 'bar',
      data: {
        labels: data.labels || [],
        datasets: [
          { label: 'Rejected', data: data.rejected || [], backgroundColor: 'rgb(239, 68, 68)', borderRadius: 2 },
          { label: 'Skipped', data: data.skipped || [], backgroundColor: 'rgb(234, 179, 8)', borderRadius: 2 },
          { label: 'Failed', data: data.failed || [], backgroundColor: 'rgb(168, 85, 247)', borderRadius: 2 }
        ]
      },
      options: chartOptions
    });

    const total = (keys) => keys.reduce((sum, key) => sum + (data[key] || []).reduce((a, value) => a + value, 0), 0);
    const delivered = total(['added', 'requested']);
    const simulated = total(['would_add', 'would_request']);
    document.getElementById('deliveryChartTotal').textContent = `${delivered.toLocaleString()} delivered${data.dry_run && simulated ? ` · ${simulated.toLocaleString()} simulated` : ''}`;
    document.getElementById('decisionChartTotal').textContent = `${total(['rejected', 'skipped', 'failed']).toLocaleString()} decisions`;
  } catch (err) {
    console.error('Failed to load chart:', err);
  }
}

async function loadJobRuns() {
  try {
    const [response, jobsResponse] = await Promise.all([fetch('/v1/activity/runs?limit=50'), fetch('/v1/jobs/list')]);
    if (!response.ok) throw new Error(`Request failed (${response.status})`);
    const data = await response.json();
    if (jobsResponse.ok) {
      const jobs = await jobsResponse.json();
      jobsByID = new Map((Array.isArray(jobs) ? jobs : []).map((job) => [job.id, job]));
    }
    recentRuns = sortRunsNewestFirst((data && data.runs) ? data.runs : []);
	  const cycles = (data && data.cycles) ? data.cycles : [];
	  hasJobRuns = recentRuns.length > 0 || cycles.length > 0;
	  updateActivityEmptyState();
	  renderSelectionCycles(cycles);

    renderActivityTimeline();
  } catch (err) {
    const timeline = document.getElementById('activityTimelineList');
    if (timeline) timeline.innerHTML = '<div class="text-red-400">Failed to load timeline.</div>';
  }
}

function renderSelectionCycles(cycles) {
  const container = document.getElementById('selectionCycleTimeline');
  if (!container) return;
  if (!cycles.length) {
    container.classList.add('hidden');
    return;
  }
  const latest = cycles[0];
  const statusClass = latest.status === 'completed' ? 'text-green-400' : latest.status === 'failed' ? 'text-red-400' : 'text-blue-400';
  const accounting = latest.accounting_complete
    ? `${(latest.movie_winners || 0) + (latest.show_winners || 0)} planned · ${(latest.movie_delivered || 0) + (latest.show_delivered || 0)} delivered · ${latest.failed_items || 0} failed`
    : `${(latest.movie_winners || 0) + (latest.show_winners || 0)} planned · delivery outcome unavailable`;
  container.classList.remove('hidden');
  container.innerHTML = `
    <div class="job-nested-panel mb-3 flex flex-wrap items-center justify-between gap-3 p-3 text-xs">
      <span><strong class="text-slate-100">Latest ranked selection</strong> · ${escapeHTML(new Date(latest.started_at).toLocaleString())}</span>
      <span>${accounting} · <span class="${statusClass}">${escapeHTML(latest.status)}</span></span>
    </div>`;
}

// Handle stats response
htmx.on('htmx:afterSwap', function(evt) {
  if (evt.detail.target.id === 'stats') {
    const stats = JSON.parse(evt.detail.xhr.response);
    hasActivityEntries = ['total_added', 'total_rejected', 'total_skipped', 'total_failed']
      .some((key) => Number(stats[key]) > 0);
    updateActivityEmptyState();
    evt.detail.target.innerHTML = `
      <div class="activity-metric">
        <div class="flex justify-between items-start mb-2">
          <div class="text-sm text-green-100 font-medium">Total Delivered</div>
          <svg class="w-8 h-8 text-green-200 opacity-75" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="text-4xl font-bold text-white mb-1">${stats.total_added || 0}</div>
        <div class="text-xs text-green-100">Direct additions and requests</div>
      </div>
      <div class="activity-metric">
        <div class="flex justify-between items-start mb-2">
          <div class="text-sm text-yellow-100 font-medium">Total Rejected</div>
          <svg class="w-8 h-8 text-yellow-200 opacity-75" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M13.477 14.89A6 6 0 015.11 6.524l8.367 8.368zm1.414-1.414L6.524 5.11a6 6 0 018.367 8.367zM18 10a8 8 0 11-16 0 8 8 0 0116 0z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="text-4xl font-bold text-white mb-1">${stats.total_rejected || 0}</div>
        <div class="text-xs text-yellow-100">Filtered by rules</div>
      </div>
      <div class="activity-metric">
        <div class="flex justify-between items-start mb-2">
          <div class="text-sm text-amber-100 font-medium">Total Skipped</div>
          <svg class="w-8 h-8 text-amber-200 opacity-75" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm4-8a1 1 0 11-2 0 1 1 0 012 0zm-4 0a1 1 0 11-2 0 1 1 0 012 0zm-4 0a1 1 0 11-2 0 1 1 0 012 0z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="text-4xl font-bold text-white mb-1">${stats.total_skipped || 0}</div>
        <div class="text-xs text-amber-100">Already existed/requested</div>
      </div>
      <div class="activity-metric">
        <div class="flex justify-between items-start mb-2">
          <div class="text-sm text-red-100 font-medium">Total Failed</div>
          <svg class="w-8 h-8 text-red-200 opacity-75" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="text-4xl font-bold text-white mb-1">${stats.total_failed || 0}</div>
        <div class="text-xs text-red-100">Errors encountered</div>
      </div>
      <div class="activity-metric">
        <div class="flex justify-between items-start mb-2">
          <div class="text-sm text-blue-100 font-medium">Last 24 Hours</div>
          <svg class="w-8 h-8 text-blue-200 opacity-75" fill="currentColor" viewBox="0 0 20 20">
            <path fill-rule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-12a1 1 0 10-2 0v4a1 1 0 00.293.707l2.828 2.829a1 1 0 101.415-1.415L11 9.586V6z" clip-rule="evenodd" />
          </svg>
        </div>
        <div class="text-4xl font-bold text-white mb-1">${stats.added_last_24h || 0}</div>
        <div class="text-xs text-blue-100">Recent activity</div>
      </div>
    `;
    loadActivityChart();
    loadRejectionBreakdown();
  }
});

// Ensure action buttons update visible data immediately after success.
document.body.addEventListener('htmx:afterRequest', function(evt) {
  const detail = evt.detail || {};
  if (!detail.successful) return;
  const trigger = detail.elt || (detail.requestConfig && detail.requestConfig.elt) || evt.target;
  if (!trigger || !trigger.closest) return;

  const isAddAnyway = !!trigger.closest('.add-anyway-btn');
  const isBlock = !!trigger.closest('.block-btn');
  if (!isAddAnyway && !isBlock) return;

  refreshActivityData();
});

// Parse filter details after HTMX loads the content
htmx.on('htmx:afterSettle', function(evt) {
  if (evt.detail.target.id === 'activityTable') {
    parseFilterDetailsIn(evt.detail.target);
  }
});

// Load rejection breakdown
async function loadRejectionBreakdown() {
  try {
    const response = await fetch('/v1/activity/rejection-breakdown');
    const data = await response.json();
    
    if (data.total_rejected > 0) {
      document.getElementById('rejection-breakdown').classList.remove('hidden');
      document.getElementById('rejection-total').textContent = `(${data.total_rejected} total)`;
      
      const container = document.getElementById('rejection-reasons');
      container.innerHTML = '';
      
      // Sort by count descending
      const sorted = Object.entries(data.breakdown).sort((a, b) => b[1] - a[1]);
      
      sorted.forEach(([reason, count]) => {
        const percentage = ((count / data.total_rejected) * 100).toFixed(0);
        container.innerHTML += `
          <div class="rejection-row">
            <span>${escapeHTML(reason)}</span>
            <span class="rejection-bar"><i style="width:${percentage}%"></i></span>
            <b>${Number(count || 0)}</b>
            <small>${percentage}%</small>
          </div>
        `;
      });
    }
  } catch (err) {
    // Silently fail if no rejections
  }
}
