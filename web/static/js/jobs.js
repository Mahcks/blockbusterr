  // Owns Jobs page state and the /v1/jobs/* interactions. The module stays
  // global-compatible while templates migrate away from inline handlers.
  let allJobs = [];
  let jobModalRequest = 0;
  let jobTypes = {};
  let jobTemplates = [];
  let ruleSets = [];
  let currentJob = null;
  let currentTemplateCategory = 'Movies';
  let templateSearchQuery = '';
  let previewData = null;
  let currentFilter = 'all';
  let jobFormSnapshot = null;
  const globalMode = document.getElementById("jobs-page")?.dataset.globalMode || "direct";

  function openDialog(id, trigger = document.activeElement, onRequestClose) {
    window.blockbusterrDialog.open(id, { trigger, onRequestClose });
  }

  function closeDialog(id) {
    window.blockbusterrDialog.close(id);
  }

  // Job type icons and colors
  function getTypeIcon(type, className) {
    const cls = className || 'w-6 h-6';
    const iconNames = {
      trending: 'trending-up',
      popular: 'flame',
      watched: 'eye',
      collected: 'archive',
      favorited: 'heart',
      played: 'play',
      anticipated: 'clock-3',
      box_office: 'ticket',
      smart_popular: 'brain',
      list: 'list',
      recommendations: 'sparkles'
    };
    const iconName = iconNames[type] || iconNames.popular;
    return `<i data-lucide="${iconName}" class="${cls}"></i>`;
  }

  const typeStyles = {
    'trending': { color: 'text-yellow-400' },
    'popular': { color: 'text-red-400' },
    'watched': { color: 'text-cyan-400' },
    'collected': { color: 'text-amber-400' },
    'favorited': { color: 'text-pink-400' },
    'played': { color: 'text-blue-400' },
    'anticipated': { color: 'text-purple-400' },
    'box_office': { color: 'text-green-400' },
    'smart_popular': { color: 'text-indigo-400' },
    'list': { color: 'text-emerald-400' },
    'recommendations': { color: 'text-violet-400' }
  };

  // Initialize
  document.addEventListener('DOMContentLoaded', async function() {
    await loadData();
    renderJobsList();
    renderTemplates();

    // Supports "navigate to the relevant job" links from the Rules page.
    const requestedJobId = new URLSearchParams(window.location.search).get('job');
    if (requestedJobId && allJobs.some(job => String(job.id) === requestedJobId)) {
      openJobModal(requestedJobId);
    }

    const templateSearch = document.getElementById('template-search');
    if (templateSearch) {
      templateSearch.addEventListener('input', function(evt) {
        templateSearchQuery = String(evt.target.value || '').trim().toLowerCase();
        renderTemplates();
      });
    }

    document.getElementById('custom-type')?.addEventListener('change', updateCustomFormFields);
    document.getElementById('custom-source')?.addEventListener('change', updateCustomFormFields);
	document.getElementById('custom-list-kind')?.addEventListener('change', () => updateListGuidance('custom'));
	document.getElementById('custom-selection-cycle')?.addEventListener('change', () => updateSelectionFields('custom'));
    document.getElementById('custom-media')?.addEventListener('change', () => {
      updateCustomRuleSets();
      updateCustomFormFields();
    });
    document.getElementById('modal-type')?.addEventListener('change', onJobTypeChange);
    document.getElementById('modal-source')?.addEventListener('change', updateModalTypeFields);
	document.getElementById('modal-list-kind')?.addEventListener('change', () => updateListGuidance('modal'));
	document.getElementById('modal-selection-cycle')?.addEventListener('change', () => updateSelectionFields('modal'));
    document.getElementById('modal-rule-set')?.addEventListener('change', updateRuleSetSummary);
    document.addEventListener('click', handleJobsAction);
    const jobForm = document.getElementById('job-config-form');
    jobForm?.addEventListener('input', updateJobDirtyState);
    jobForm?.addEventListener('change', updateJobDirtyState);
    window.addEventListener('beforeunload', (event) => {
      if (!isJobDirty()) return;
      event.preventDefault();
      event.returnValue = '';
    });
    // 'error' doesn't bubble for <img>, so this must be a capture-phase listener.
    document.getElementById('preview-content')?.addEventListener('error', handlePreviewPosterError, true);

    // Add event listener for mode dropdown to show/hide direct mode fields
    const modeSelect = document.getElementById('modal-mode');
    if (modeSelect) {
      modeSelect.addEventListener('change', function() {
        if (currentJob) {
          // Create temp job with current form values
          const tempJob = {
            ...currentJob,
            type: document.getElementById('modal-type').value,
            media: document.getElementById('modal-media').value
          };
          updateDirectModeFields(tempJob);
        }
      });
    }

    // Add event listener for media type dropdown to update direct mode fields
    const mediaSelect = document.getElementById('modal-media');
    if (mediaSelect) {
      mediaSelect.addEventListener('change', function() {
        switchJobFilterMedia(mediaSelect.value);
        updateModalTypeFields();
        if (currentJob) {
          // Create temp job with current form values
          const tempJob = {
            ...currentJob,
            type: document.getElementById('modal-type').value,
            media: document.getElementById('modal-media').value
          };
          updateDirectModeFields(tempJob);
        }
      });
    }

    document.getElementById('import-job-file')?.addEventListener('change', importJobBundle);
  });

  function handleJobsAction(event) {
    const target = event.target.closest('[data-action]');
    if (!target) return;
    const actions = {
      'migrate-jobs': () => migrateJobs(),
      'dismiss-migration': () => dismissMigrationBanner(),
      'open-add-job': () => openAddJobModal(target),
      'close-add-job': () => closeAddJobModal(),
      'filter-templates': () => filterTemplates(target.dataset.category),
      'show-custom-job': () => showCustomJobForm(),
      'open-job': () => openJobModal(target.dataset.jobId, target),
      'open-job-filters': () => openJobFilters(target.dataset.jobId, target),
      'customize-job-rules': () => customizeJobRules(target),
      'trigger-job': () => triggerJob(target.dataset.jobId),
      'delete-job': () => deleteJob(target.dataset.jobId),
      'create-from-template': () => {
        const template = jobTemplates[Number(target.dataset.templateIndex)];
        if (!template) return;
        if (!template.ready) {
          showNotification(template.missing || 'Complete setup before using this recipe.', 'error');
          return;
        }
        createJobFromTemplate(template);
      },
      'close-preview': () => closePreviewModal(),
      'filter-preview': () => filterPreview(target.dataset.filter),
      'close-job': () => closeJobModal(),
      'toggle-advanced': () => toggleModalAdvanced(),
      'preview-job': () => previewJob(),
      'run-job': () => runJobNow(),
      'toggle-job': () => toggleJobEnabled(),
      'export-job': () => exportCurrentJob(),
	  'inspect-list': () => inspectList(target.dataset.scope, target),
      'delete-current-job': () => deleteCurrentJob()
    };
    if (!actions[target.dataset.action]) return;
    event.preventDefault();
    actions[target.dataset.action]();
  }

  async function loadData() {
    try {
      // Load all data in parallel
      const [jobsRes, typesRes, templatesRes, ruleSetsRes] = await Promise.all([
        fetch('/v1/jobs/list'),
        fetch('/v1/jobs/types'),
        fetch('/v1/jobs/templates'),
        fetch('/v1/rule-sets')
      ]);

      allJobs = await jobsRes.json();
      jobTypes = await typesRes.json();
      jobTemplates = await templatesRes.json();
      ruleSets = (await ruleSetsRes.json()).map(item => ({ ...item.rule_set, usage_count: item.usage_count }));

    } catch (error) {
      console.error('Failed to load data:', error);
      showNotification('Failed to load jobs data', 'error');
    }
  }

  function renderJobsList() {
    const activeContainer = document.getElementById('active-jobs-list');
    const noJobsMsg = document.getElementById('no-jobs-message');
    const jobsCount = document.getElementById('jobs-count');
	const unavailableSection = document.getElementById('unavailable-jobs-section');
	const unavailableContainer = document.getElementById('unavailable-jobs-list');
	const unavailableCount = document.getElementById('unavailable-jobs-count');
	const activeJobs = allJobs.filter(isJobAvailable);
	const unavailableJobs = allJobs.filter(job => !isJobAvailable(job));

	jobsCount.textContent = activeJobs.length;

    // Check for legacy jobs and show migration banner
    checkLegacyJobs();

    if (activeJobs.length === 0) {
      activeContainer.innerHTML = '';
      noJobsMsg.classList.remove('hidden');
    } else {
      noJobsMsg.classList.add('hidden');
	  activeContainer.innerHTML = renderJobCards(activeJobs, false);
	}

	if (unavailableJobs.length) {
	  unavailableCount.textContent = unavailableJobs.length;
	  unavailableContainer.innerHTML = renderJobCards(unavailableJobs, true);
	  unavailableSection.classList.remove('hidden');
	} else {
	  unavailableContainer.innerHTML = '';
	  unavailableSection.classList.add('hidden');
	}

	if (window.renderLucideIcons) {
	  window.renderLucideIcons(activeContainer);
	  window.renderLucideIcons(unavailableContainer);
	}
  }

  function isJobAvailable(job) {
	const source = job.source || 'trakt';
	return jobTypes[job.type]?.sources?.includes(source) || false;
  }

  function renderJobCards(items, unavailable) {
	return items.map(job => {
        const style = typeStyles[job.type] || typeStyles['popular'];
        const isLegacy = job.id.startsWith('legacy_');
		const source = job.source || 'trakt';
		const sourceLabel = source === 'tmdb' ? 'TMDB' : capitalize(source);
		const assignedRules = ruleSets.find(rules => rules.id === job.rule_set_id) || ruleSets.find(rules => rules.id === `default-${job.media === 'show' ? 'shows' : 'movies'}`);

        return `
		  <article class="job-row">
            <button type="button" class="job-row-main" data-action="open-job" data-job-id="${escapeHTML(job.id)}">
              <span class="job-row-icon ${style.color}" aria-hidden="true">${getTypeIcon(job.type, 'w-4 h-4')}</span>
              <span class="min-w-0">
                <span class="flex min-w-0 items-center gap-2">
                  <span class="truncate text-sm font-medium text-slate-100">${escapeHTML(job.name)}</span>
                  ${isLegacy ? '<span class="text-xs text-slate-500">Legacy</span>' : ''}
                  ${job.enabled ? '' : '<span class="text-xs text-slate-500">Disabled</span>'}
                  ${unavailable ? '<span class="text-xs font-medium text-amber-300">Setup required</span>' : ''}
                </span>
                <span class="mt-0.5 block truncate text-xs text-slate-500">${escapeHTML(capitalize(job.type))} · ${job.limit} items${job.selection_cycle === true ? ' · Ranked selection' : ''}</span>
              </span>
            </button>
            <span class="job-row-value">${escapeHTML(sourceLabel)}</span>
            <span class="job-row-value">${job.media === 'movie' ? 'Movie' : 'TV show'}</span>
            <button type="button" class="job-filter-policy" data-action="open-job-filters" data-job-id="${escapeHTML(job.id)}" ${isLegacy ? 'disabled' : ''}>
              <span>${escapeHTML(assignedRules?.name || 'Rules unavailable')}</span>
              <small>${assignedRules ? `Revision ${assignedRules.revision}` : 'Check assignment'}</small>
            </button>
            <div class="job-row-actions">
				  ${unavailable || !job.enabled ? '' : `<button
                    data-action="trigger-job"
                    data-job-id="${escapeHTML(job.id)}"
                    class="icon-button hover:text-green-300"
                    title="Run now"
                    aria-label="Run job now"
                  >
                    <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M14.752 11.168l-3.197-2.132A1 1 0 0010 9.87v4.263a1 1 0 001.555.832l3.197-2.132a1 1 0 000-1.664z"></path>
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
                    </svg>
				  </button>`}
                  ${!isLegacy ? `
                  <button
                    data-action="delete-job"
                    data-job-id="${escapeHTML(job.id)}"
                    class="icon-button hover:text-red-300"
                    title="Delete job"
                    aria-label="Delete job"
                  >
                    <svg class="h-5 w-5" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true">
                      <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"></path>
                    </svg>
                  </button>
                  ` : ''}
            </div>
          </article>
        `;
	}).join('');
  }

  function capitalize(str) {
    if (!str) return '';
    return str.charAt(0).toUpperCase() + str.slice(1).replace(/_/g, ' ');
  }

  // Job names and discovery metadata cross into HTML template strings.
  function escapeHTML(value) {
    return String(value ?? '').replace(/[&<>"']/g, character => ({
      '&': '&amp;',
      '<': '&lt;',
      '>': '&gt;',
      '"': '&quot;',
      "'": '&#39;'
    })[character]);
  }

  // Migration functions
  let migrationDismissed = false;

  function checkLegacyJobs() {
    if (migrationDismissed) return;

    const legacyJobs = allJobs.filter(job => job.id.startsWith('legacy_'));
    const banner = document.getElementById('migration-banner');
    const countEl = document.getElementById('legacy-count');
    const action = banner.querySelector('[data-action="migrate-jobs"]');

    if (legacyJobs.length > 0) {
      countEl.textContent = legacyJobs.length;
      action.textContent = `Upgrade ${legacyJobs.length} ${legacyJobs.length === 1 ? 'job' : 'jobs'}`;
      banner.classList.remove('hidden');
    } else {
      banner.classList.add('hidden');
    }
  }

  function dismissMigrationBanner() {
    migrationDismissed = true;
    document.getElementById('migration-banner').classList.add('hidden');
  }

  async function migrateJobs() {
    const banner = document.getElementById('migration-banner');
    const action = banner.querySelector('[data-action="migrate-jobs"]');
    const status = document.getElementById('migration-status');
    const count = Number(document.getElementById('legacy-count').textContent) || 0;
    if (!confirm(`Upgrade ${count} legacy ${count === 1 ? 'job' : 'jobs'}? Existing settings are preserved and the v1 entries are disabled after the new configuration is saved.`)) {
      return;
    }

    action.disabled = true;
    action.setAttribute('aria-busy', 'true');
    action.textContent = 'Upgrading…';
    status.classList.add('hidden');

    try {
      const response = await fetch('/v1/jobs/migrate', {
        method: 'POST'
      });

      const data = await response.json().catch(() => ({}));

      if (!response.ok) {
        throw new Error(data.error || 'Migration failed');
      }

      if (data.count === 0) {
        showNotification('No legacy jobs to migrate', 'info');
        banner.classList.add('hidden');
      } else {
        showNotification(`${data.count} ${data.count === 1 ? 'job' : 'jobs'} upgraded`, 'success');
        await loadData();
        renderJobsList();
      }
    } catch (error) {
      status.textContent = error.message || 'Migration failed. Your existing jobs were not changed.';
      status.classList.remove('hidden');
      showNotification(error.message, 'error');
    } finally {
      action.disabled = false;
      action.removeAttribute('aria-busy');
      if (!banner.classList.contains('hidden')) {
        action.textContent = `Upgrade ${count} ${count === 1 ? 'job' : 'jobs'}`;
      }
    }
  }

  // Add Job Modal
  function openAddJobModal(trigger) {
    // Reset to templates view
    templateSearchQuery = '';
    const templateSearch = document.getElementById('template-search');
    if (templateSearch) templateSearch.value = '';
    showTemplatesView();
    filterTemplates('Movies');
    openDialog('add-job-modal', trigger, closeAddJobModal);
  }

  function closeAddJobModal() {
    closeDialog('add-job-modal');
    // Reset custom form
    document.getElementById('custom-job-form').reset();
  }

  function exportCurrentJob() {
    if (!currentJob || currentJob.id.startsWith('legacy_')) return;
    window.location.assign(`/config/jobs/${encodeURIComponent(currentJob.id)}/export`);
  }

  async function importJobBundle(event) {
    const input = event.currentTarget;
    const file = input.files?.[0];
    if (!file) return;
    const body = new FormData();
    body.append('config', file);
    try {
      const response = await fetch('/config/jobs/import', { method: 'POST', body });
      const result = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(result.error || 'Could not import job');
      await loadData();
      renderJobsList();
      closeAddJobModal();
	  showNotification(`${result.name} imported disabled with ${result.rule_set_name}; IDs regenerated.`, 'success');
	  currentJob = result;
	  previewJob();
    } catch (error) {
      showNotification(error.message, 'error');
    } finally {
      input.value = '';
    }
  }

  function updateTabStyles(activeTab) {
    const tabs = ['tab-movies', 'tab-shows', 'tab-custom'];
    tabs.forEach(tabId => {
      const tab = document.getElementById(tabId);
      tab.setAttribute('aria-selected', String(tabId === activeTab));
    });
  }

  function showTemplatesView() {
    document.getElementById('templates-content').classList.remove('hidden');
    document.getElementById('custom-job-content').classList.add('hidden');
  }

  function filterTemplates(category) {
    currentTemplateCategory = category;
    showTemplatesView();
    updateTabStyles(category === 'Movies' ? 'tab-movies' : 'tab-shows');
    renderTemplates();
  }

  function showCustomJobForm() {
    document.getElementById('templates-content').classList.add('hidden');
    document.getElementById('custom-job-content').classList.remove('hidden');
    updateTabStyles('tab-custom');

    // Populate job type dropdown
    const typeSelect = document.getElementById('custom-type');
    typeSelect.innerHTML = Object.entries(jobTypes).map(([key, def]) => {
      return `<option value="${key}">${def.name}${def.sources?.length ? '' : ' (unavailable)'}</option>`;
    }).join('');

    // Trigger initial field update
    updateCustomFormFields();
  }

  function updateCustomFormFields() {
    const typeSelect = document.getElementById('custom-type');
    const mediaSelect = document.getElementById('custom-media');
    const periodContainer = document.getElementById('custom-period-container');
    const smartContainer = document.getElementById('custom-smart-container');
    const descriptionEl = document.getElementById('custom-type-description');
    const limitInput = document.getElementById('custom-limit');
	const sourceSelect = document.getElementById('custom-source');
	const listContainer = document.getElementById('custom-list-container');
	const recommendationsContainer = document.getElementById('custom-recommendations-container');

    const selectedType = typeSelect.value;
    const typeDef = jobTypes[selectedType];

    if (!typeDef) return;
	const currentSource = sourceSelect.value;
	const sources = typeDef.sources || [];
	const knownSources = typeDef.known_sources || sources;
	sourceSelect.innerHTML = knownSources.map(source => `<option value="${source}" ${sources.includes(source) ? '' : 'disabled'}>${source === 'tmdb' ? 'TMDB' : capitalize(source)}${sources.includes(source) ? '' : ' — unavailable'}</option>`).join('');
	if (!sources.length) sourceSelect.insertAdjacentHTML('afterbegin', '<option value="" selected>No configured source available</option>');
	if (sources.includes(currentSource)) sourceSelect.value = currentSource;
	document.getElementById('custom-source-help').textContent = !sources.length
	  ? (selectedType === 'list' ? 'List providers become selectable when Blockbusterr supports them and their settings are configured.' : 'Configure a supported discovery source in Settings to create this job.')
	  : sourceSelect.value === 'simkl'
	  ? 'Simkl attribution will be shown with sourced results.'
	  : 'Only configured discovery sources are shown.';
	listContainer.classList.toggle('hidden', selectedType !== 'list');
	recommendationsContainer.classList.toggle('hidden', selectedType !== 'recommendations');
	populateRecommendationListSources('custom');
	if (selectedType === 'list') updateListGuidance('custom');
	document.getElementById('custom-create-button').disabled = !sources.length;
	const customPeriod = document.getElementById('custom-period');
	if (sourceSelect.value === 'simkl' && selectedType === 'watched') {
	  customPeriod.innerHTML = '<option value="weekly">Weekly</option><option value="monthly">Monthly</option>';
	} else {
	  customPeriod.innerHTML = '<option value="weekly">Weekly</option><option value="monthly">Monthly</option><option value="yearly">Yearly</option><option value="all">All Time</option>';
	}

    // Update description
    const style = typeStyles[selectedType] || {};
    descriptionEl.innerHTML = `
      <div class="flex items-center gap-2">
        <span class="${style.color || 'text-slate-300'}">${getTypeIcon(selectedType, 'w-5 h-5')}</span>
        <div>
          <p class="text-sm font-medium text-slate-200">${typeDef.name}</p>
          <p class="text-sm text-slate-400">${typeDef.description}</p>
        </div>
      </div>
    `;
    if (window.renderLucideIcons) {
      window.renderLucideIcons(descriptionEl);
    }

    // Update media type options based on supported media
    const supportedMedia = typeDef.supported_media || ['movie', 'show'];
    const currentMedia = mediaSelect.value;
    mediaSelect.innerHTML = '';
    if (supportedMedia.includes('movie')) {
      mediaSelect.innerHTML += '<option value="movie">Movie</option>';
    }
    if (supportedMedia.includes('show')) {
      mediaSelect.innerHTML += '<option value="show">TV Show</option>';
    }
    if (supportedMedia.includes(currentMedia)) mediaSelect.value = currentMedia;
    document.getElementById('custom-series-type-container').classList.toggle('hidden', mediaSelect.value !== 'show');

    // Show/hide period field
    if (typeDef.requires_period) {
      periodContainer.classList.remove('hidden');
    } else {
      periodContainer.classList.add('hidden');
    }

    // Show/hide smart job fields
    if (typeDef.is_smart_job) {
      smartContainer.classList.remove('hidden');
    } else {
      smartContainer.classList.add('hidden');
    }

    // Update limit max and description
    const maxLimit = sourceSelect.value === 'simkl' ? Math.min(typeDef.max_limit || 1000, 500) : (typeDef.max_limit || 1000);
    limitInput.max = maxLimit;
    const limitDescription = document.getElementById('custom-limit-description');
    if (limitDescription) {
      limitDescription.textContent = `Number of items to fetch (max ${maxLimit})`;
    }
    if (parseInt(limitInput.value) > maxLimit) {
      limitInput.value = maxLimit;
    }
    updateCustomRuleSets();
  }

  function updateCustomRuleSets() {
    const media = document.getElementById('custom-media')?.value || 'movie';
    const select = document.getElementById('custom-rule-set');
    if (!select) return;
    const current = select.value;
    const compatible = ruleSets.filter(rules => rules.media === media);
    select.replaceChildren(...compatible.map(rules => new Option(`${rules.name} · ${rules.usage_count} job${rules.usage_count === 1 ? '' : 's'}`, rules.id)));
    select.value = compatible.some(rules => rules.id === current) ? current : `default-${media === 'show' ? 'shows' : 'movies'}`;
  }

  function updateSelectionFields(scope) {
	const enabled = document.getElementById(`${scope}-selection-cycle`)?.checked;
	const minimum = document.getElementById(`${scope}-minimum-picks`);
	if (!minimum) return;
	minimum.disabled = !enabled;
	if (!enabled) minimum.value = '0';
	if (scope === 'modal') {
	  const interval = document.getElementById('modal-interval');
	  const help = document.getElementById('modal-interval-help');
	  interval.disabled = enabled;
	  help.textContent = enabled
	    ? 'Managed by the shared ranked-selection schedule in Settings.'
	    : 'Override the default interval (e.g., 1h, 30m, 0 */2 * * *).';
	  const runButton = document.getElementById('modal-run-button');
	  const managedByCycle = enabled || currentJob?.selection_cycle === true;
	  runButton.disabled = managedByCycle;
	  runButton.title = managedByCycle
	    ? 'Participating jobs run only through the shared ranked-selection cycle.'
	    : 'Runs this job immediately, outside its schedule.';
	}
  }

  // Custom job form submission
  document.addEventListener('DOMContentLoaded', function() {
    const customForm = document.getElementById('custom-job-form');
    if (customForm) {
      customForm.addEventListener('submit', async function(e) {
        e.preventDefault();

        const formData = new FormData(e.target);
        const selectedType = formData.get('type');
        const typeDef = jobTypes[selectedType];

        const job = {
          name: formData.get('name'),
          type: selectedType,
          media: formData.get('media'),
          enabled: true,
          limit: parseInt(formData.get('limit')),
		  delivery_limit: parseInt(formData.get('delivery_limit')) || 0,
		  selection_cycle: formData.get('selection_cycle') === 'on',
		  minimum_picks: parseInt(formData.get('minimum_picks')) || 0,
		  repeat_policy: formData.get('repeat_policy') || '',
		  source: formData.get('source') || 'trakt',
          rule_set_id: formData.get('rule_set_id')
        };
        if (selectedType === 'list') {
		  job.list = {
			kind: formData.get('list_kind'),
			owner: formData.get('list_owner'),
			list_id: formData.get('list_id'),
			ordering: formData.get('list_ordering')
		  };
		}
		if (selectedType === 'recommendations') {
		  job.recommendation_seeds = parseTMDBSeeds(formData.get('recommendation_seeds'));
		  job.recommendation_list = recommendationListFromForm('custom');
		}
		if (job.media === 'show') job.series_type = formData.get('series_type') || 'standard';

        // Add period if required
        if (typeDef?.requires_period) {
          job.period = formData.get('period');
        }

        // Add smart job fields if applicable
        if (typeDef?.is_smart_job) {
          job.base_min_rating = parseFloat(formData.get('base_min_rating'));
          job.adjustment_factor = parseFloat(formData.get('adjustment_factor'));
        }

        try {
          const response = await fetch('/v1/jobs', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(job)
          });

          if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to create job');
          }

          const newJob = await response.json();
          allJobs.push(newJob);

          closeAddJobModal();
          renderJobsList();
          showNotification(`${job.name} created successfully!`, 'success');

          // Open the edit modal for the new job
          openJobModal(newJob.id);
        } catch (error) {
          showNotification(error.message, 'error');
        }
      });
    }
  });

  function renderTemplates() {
    const grid = document.getElementById('templates-grid');
    const count = document.getElementById('templates-count');
    const filtered = jobTemplates
      .filter(t => t.category === currentTemplateCategory)
      .filter(t => {
        if (!templateSearchQuery) return true;
        const name = String(t.name || '').toLowerCase();
        const type = String(t.type || '').toLowerCase();
        const description = String(t.description || '').toLowerCase();
		const source = String(t.source || '').toLowerCase();
		const rules = String(t.rule_set_name || '').toLowerCase();
        return name.includes(templateSearchQuery) || type.includes(templateSearchQuery) || description.includes(templateSearchQuery) || source.includes(templateSearchQuery) || rules.includes(templateSearchQuery);
      })
      .sort((a, b) => String(a.name || '').localeCompare(String(b.name || '')));

    if (count) {
      count.textContent = `${filtered.filter(template => template.ready).length} of ${filtered.length} ready`;
    }

    if (!filtered.length) {
      grid.innerHTML = `
        <div class="p-6 text-center">
          <p class="text-sm font-medium text-slate-200">No recipes found</p>
          <p class="mt-1 text-xs text-slate-400">Try a different search, or use Custom Job.</p>
        </div>
      `;
      if (window.renderLucideIcons) {
        window.renderLucideIcons(grid);
      }
      return;
    }

    grid.innerHTML = filtered.map(template => {
      const style = typeStyles[template.type] || typeStyles['popular'];
	  const cadence = template.sync_interval === '24h' ? 'Daily' : template.sync_interval === '168h' ? 'Weekly' : template.sync_interval;
	  const availability = template.ready ? 'Ready to preview' : template.missing;

      return `
        <button
          type="button"
          data-action="create-from-template"
          data-template-index="${jobTemplates.indexOf(template)}"
		  aria-disabled="${String(!template.ready)}"
		  class="template-row ${template.ready ? '' : 'template-row-unavailable'}"
        >
          <div class="flex items-center gap-3">
            <span class="job-row-icon ${style.color}">${getTypeIcon(template.type, 'w-4 h-4')}</span>
            <div class="flex-1">
              <div class="flex items-center justify-between gap-3">
                <h3 class="text-sm font-medium text-slate-100">${escapeHTML(template.name)}</h3>
				<span class="shrink-0 text-xs ${template.ready ? 'text-green-400' : 'text-yellow-400'}">${escapeHTML(availability)}</span>
              </div>
              <p class="mt-0.5 text-xs text-slate-500">${escapeHTML(template.description)}</p>
			  <p class="template-meta">${escapeHTML(capitalize(template.source))} · ${escapeHTML(cadence)} · ${template.limit} candidates · ${template.delivery_limit || 'Unlimited'} deliveries · ${escapeHTML(template.rule_set_name)}</p>
            </div>
          </div>
        </button>
      `;
    }).join('');

    if (window.renderLucideIcons) {
      window.renderLucideIcons(grid);
    }
  }

  async function createJobFromTemplate(template) {
    try {
	  const response = await fetch(`/v1/jobs/recipes/${encodeURIComponent(template.id)}`, { method: 'POST' });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Failed to create job');
      }

      const newJob = await response.json();
	  await loadData();

      closeAddJobModal();
      renderJobsList();
	  showNotification(`${template.name} created disabled ${template.default_rules ? 'using the shared default rules' : 'with dedicated rules'}.`, 'success');
	  currentJob = newJob;
	  previewJob();
    } catch (error) {
      showNotification(error.message, 'error');
    }
  }

  // Job Configuration Modal
  async function openJobModal(jobId, trigger) {
    const request = ++jobModalRequest;
    let job;
    try {
      const response = await fetch(`/v1/jobs/dynamic/${encodeURIComponent(jobId)}`);
      if (!response.ok) throw new Error('Job not found');
      job = await response.json();
    } catch (error) {
      showNotification(error.message || 'Could not load job', 'error');
      return;
    }
    if (request !== jobModalRequest) return;
    if (!job) {
      showNotification('Job not found', 'error');
      return;
    }

    currentJob = job;
    const isLegacy = job.id.startsWith('legacy_');
    const style = typeStyles[job.type] || typeStyles['popular'];
    const typeDef = jobTypes[job.type];

    // Set modal values
    document.getElementById('modal-title').textContent = job.name;
    document.getElementById('modal-description').textContent = typeDef?.description || '';
    document.getElementById('modal-job-id').value = job.id;
    document.getElementById('modal-name').value = job.name;
    document.getElementById('modal-limit').value = job.limit;
    document.getElementById('modal-delivery-limit').value = job.delivery_limit || 0;
	document.getElementById('modal-selection-cycle').checked = job.selection_cycle === true;
	document.getElementById('modal-minimum-picks').value = job.minimum_picks || 0;
	updateSelectionFields('modal');
	document.getElementById('modal-repeat-policy').value = job.repeat_policy || '';
    document.getElementById('modal-interval').value = job.sync_interval || '';
    document.getElementById('modal-mode').value = job.mode || '';
    document.getElementById('modal-source').value = job.source || 'trakt';
	document.getElementById('modal-list-kind').value = job.list?.kind || 'public_list';
	document.getElementById('modal-list-ordering').value = job.list?.ordering || 'source';
	document.getElementById('modal-list-owner').value = job.list?.owner || '';
	document.getElementById('modal-list-id').value = job.list?.list_id || job.list?.slug || '';
	document.getElementById('modal-recommendation-seeds').value = (job.recommendation_seeds || []).join(', ');
	populateRecommendationListSources('modal', job.recommendation_list?.source || '');
	document.getElementById('modal-recommendation-list-owner').value = job.recommendation_list?.list?.owner || '';
	document.getElementById('modal-recommendation-list-id').value = job.recommendation_list?.list?.list_id || '';
	document.getElementById('modal-series-type').value = job.series_type || 'standard';
    document.getElementById('modal-base-rating').value = job.base_min_rating ?? 6.0;
    document.getElementById('modal-adjustment-factor').value = job.adjustment_factor ?? 0.5;

    // Populate job type dropdown
    const typeSelect = document.getElementById('modal-type');
	const selectableTypes = Object.entries(jobTypes).filter(([key, def]) => def.sources?.length || key === job.type);
    typeSelect.innerHTML = selectableTypes.map(([key, def]) => {
      return `<option value="${key}">${def.name}</option>`;
    }).join('');
    typeSelect.value = job.type;

    // Populate media dropdown and set value
    const mediaSelect = document.getElementById('modal-media');
    mediaSelect.value = job.media;

    // Set badges
    document.getElementById('modal-badges').innerHTML = `
      <span class="job-type-badge">${capitalize(job.type)}</span>
      <span class="job-type-badge ${job.media === 'movie' ? 'job-type-badge-movie' : 'job-type-badge-show'}">${job.media === 'movie' ? 'Movie' : 'TV Show'}</span>
	  <span class="job-type-badge">${job.source === 'tmdb' ? 'TMDB' : capitalize(job.source || 'trakt')}</span>
      ${isLegacy ? '<span class="job-type-badge job-type-badge-legacy">Legacy (Read-Only)</span>' : ''}
    `;

    // Set max limit based on job type
    const limitInput = document.getElementById('modal-limit');
    const limitDescription = limitInput.parentElement.querySelector('.text-xs');
    const maxLimit = typeDef?.max_limit || 100;
    limitInput.max = maxLimit;
    limitDescription.textContent = `Number of items to fetch (1-${maxLimit})`;

    // Disable editing for legacy jobs
    const nameContainer = document.getElementById('modal-name-container');
    const typeContainer = document.getElementById('modal-type-container');
    if (isLegacy) {
      nameContainer.classList.add('opacity-50');
      typeContainer.classList.add('opacity-50');
      document.getElementById('modal-name').disabled = true;
      document.getElementById('modal-type').disabled = true;
      document.getElementById('modal-media').disabled = true;
    } else {
      nameContainer.classList.remove('opacity-50');
      typeContainer.classList.remove('opacity-50');
      document.getElementById('modal-name').disabled = false;
      document.getElementById('modal-type').disabled = false;
      document.getElementById('modal-media').disabled = false;
    }

    // Update type-dependent fields (period, smart, media options)
    updateModalTypeFields();
    populateJobFilters(job, isLegacy);

    // Update direct mode fields
    updateDirectModeFields(job);

    // Update enable/disable button
    const toggleButton = document.getElementById('toggle-job-button');
    const toggleDescription = document.getElementById('toggle-job-description');
    if (job.enabled) {
      toggleButton.textContent = 'Disable Job';
      toggleButton.className = 'job-disable-button w-full';
      toggleDescription.textContent = 'Pause this job from running';
    } else {
      toggleButton.textContent = 'Enable Job';
      toggleButton.className = 'job-enable-button w-full';
      toggleDescription.textContent = 'Start this job running on schedule';
    }

    // Show/hide delete button
    const deleteZone = document.getElementById('delete-job-zone');
    if (isLegacy) {
      deleteZone.classList.add('hidden');
      document.getElementById('export-job-button').classList.add('hidden');
    } else {
      deleteZone.classList.remove('hidden');
      document.getElementById('export-job-button').classList.remove('hidden');
    }

    jobFormSnapshot = serializeJobForm();
    updateJobDirtyState();
    openDialog('job-modal', trigger, closeJobModal);
  }

  async function openJobFilters(jobId, trigger) {
    await openJobModal(jobId, trigger);
    if (!currentJob || currentJob.id !== jobId) return;
    requestAnimationFrame(() => {
      document.getElementById('modal-filter-settings').scrollIntoView({ block: 'start' });
      document.getElementById('modal-rule-set')?.focus();
    });
  }

  // Handle job type change in edit modal
  function onJobTypeChange() {
    updateModalTypeFields();
    switchJobFilterMedia(document.getElementById('modal-media').value);
    // Update direct mode fields as media type may have changed
    if (currentJob) {
      // Create a temporary job object with updated values for direct mode check
      const tempJob = {
        ...currentJob,
        type: document.getElementById('modal-type').value,
        media: document.getElementById('modal-media').value
      };
      updateDirectModeFields(tempJob);
    }
  }

  // Update modal fields based on selected job type
  function updateModalTypeFields() {
    const typeSelect = document.getElementById('modal-type');
    const mediaSelect = document.getElementById('modal-media');
    const periodContainer = document.getElementById('modal-period-container');
    const smartContainer = document.getElementById('modal-smart-container');
	const sourceSelect = document.getElementById('modal-source');
	const listContainer = document.getElementById('modal-list-container');
	const recommendationsContainer = document.getElementById('modal-recommendations-container');

    const selectedType = typeSelect.value;
    const typeDef = jobTypes[selectedType];

    if (!typeDef) return;
	const currentSource = sourceSelect.value || currentJob?.source || 'trakt';
	const sources = typeDef.sources || [];
	const knownSources = typeDef.known_sources || sources;
	sourceSelect.innerHTML = knownSources.map(source => `<option value="${source}" ${sources.includes(source) || source === currentSource ? '' : 'disabled'}>${source === 'tmdb' ? 'TMDB' : capitalize(source)}${sources.includes(source) ? '' : ' — unavailable'}</option>`).join('');
	if (sources.includes(currentSource)) sourceSelect.value = currentSource;
	else if (knownSources.includes(currentSource)) sourceSelect.value = currentSource;
	document.getElementById('modal-source-help').textContent = sources.length
	  ? (sourceSelect.value === 'simkl' ? 'Simkl attribution will be shown with sourced results.' : 'Only configured discovery sources are shown.')
	  : 'This job requires a discovery source that is not configured. Change its type or delete it.';
	listContainer.classList.toggle('hidden', selectedType !== 'list');
	recommendationsContainer.classList.toggle('hidden', selectedType !== 'recommendations');
	populateRecommendationListSources('modal');
	if (selectedType === 'list') updateListGuidance('modal');
	const modalPeriod = document.getElementById('modal-period');
	if (sourceSelect.value === 'simkl' && selectedType === 'watched') {
	  modalPeriod.innerHTML = '<option value="weekly">Weekly</option><option value="monthly">Monthly</option>';
	} else {
	  modalPeriod.innerHTML = '<option value="weekly">Weekly</option><option value="monthly">Monthly</option><option value="yearly">Yearly</option><option value="all">All Time</option>';
	}
	if ([...modalPeriod.options].some(option => option.value === currentJob?.period)) {
	  modalPeriod.value = currentJob.period;
	}

    // Update media type options based on supported media
    const supportedMedia = typeDef.supported_media || ['movie', 'show'];
    const currentMedia = mediaSelect.value;
    mediaSelect.innerHTML = '';
    if (supportedMedia.includes('movie')) {
      mediaSelect.innerHTML += '<option value="movie">Movie</option>';
    }
    if (supportedMedia.includes('show')) {
      mediaSelect.innerHTML += '<option value="show">TV Show</option>';
    }
    // Try to keep current selection if still supported
    if (supportedMedia.includes(currentMedia)) {
      mediaSelect.value = currentMedia;
    }
    document.getElementById('modal-series-type-container').classList.toggle('hidden', mediaSelect.value !== 'show');

    // Update max limit based on job type
    const limitInput = document.getElementById('modal-limit');
    const limitDescription = limitInput.parentElement.querySelector('.text-xs');
    const maxLimit = sourceSelect.value === 'simkl' ? Math.min(typeDef.max_limit || 100, 500) : (typeDef.max_limit || 100);
    limitInput.max = maxLimit;
    limitDescription.textContent = `Number of items to fetch (1-${maxLimit})`;
    if (parseInt(limitInput.value) > maxLimit) {
      limitInput.value = maxLimit;
    }

    // Show/hide period field
    if (typeDef.requires_period) {
      periodContainer.classList.remove('hidden');
      // Set default period if not already set
      const periodSelect = document.getElementById('modal-period');
      if (!periodSelect.value) {
        periodSelect.value = currentJob?.period || 'weekly';
      }
    } else {
      periodContainer.classList.add('hidden');
    }

    // Show/hide smart job fields
    if (typeDef.is_smart_job) {
      smartContainer.classList.remove('hidden');
      // Set default values if not already set
      const baseRating = document.getElementById('modal-base-rating');
      const adjustmentFactor = document.getElementById('modal-adjustment-factor');
      if (!baseRating.value) {
        baseRating.value = currentJob?.base_min_rating ?? 6.0;
      }
      if (!adjustmentFactor.value) {
        adjustmentFactor.value = currentJob?.adjustment_factor ?? 0.5;
      }
    } else {
      smartContainer.classList.add('hidden');
    }

    // Update modal description
    document.getElementById('modal-description').textContent = typeDef.description || '';
  }

  function populateJobFilters(job, isLegacy) {
    const select = document.getElementById('modal-rule-set');
    const compatible = ruleSets.filter(rules => rules.media === job.media);
    select.replaceChildren(...compatible.map(rules => new Option(`${rules.name} · ${rules.usage_count} job${rules.usage_count === 1 ? '' : 's'}`, rules.id)));
    select.value = job.rule_set_id || `default-${job.media === 'show' ? 'shows' : 'movies'}`;
    select.disabled = isLegacy;
    document.getElementById('modal-filter-settings').classList.toggle('opacity-50', isLegacy);
    updateRuleSetSummary();
  }

  function switchJobFilterMedia(media) {
    if (!currentJob) return;
    populateJobFilters({ ...currentJob, media, rule_set_id: '' }, currentJob.id.startsWith('legacy_'));
  }

  function updateRuleSetSummary() {
    const rules = ruleSets.find(item => item.id === document.getElementById('modal-rule-set')?.value);
    document.getElementById('modal-rule-set-summary').textContent = rules ? `${describeRuleSet(rules)} Changes apply to every assigned job.` : 'Select a compatible rule set.';
    const customize = document.getElementById('customize-job-rules');
    const isJobSpecific = rules && !rules.id.startsWith('default-') && rules.usage_count === 1 && currentJob?.rule_set_id === rules.id;
    customize.classList.toggle('hidden', Boolean(isJobSpecific) || currentJob?.id.startsWith('legacy_'));
    const edit = document.getElementById('edit-job-rules');
    edit.textContent = isJobSpecific ? 'Edit job rules' : 'View selected rules';
    edit.href = rules ? `/filters?ruleSet=${encodeURIComponent(rules.id)}` : '/filters';
  }

  function updateListGuidance(scope) {
	const source = document.getElementById(`${scope}-source`)?.value;
	const kind = document.getElementById(`${scope}-list-kind`)?.value;
	const owner = document.getElementById(`${scope}-list-owner`);
	const listID = document.getElementById(`${scope}-list-id`);
	if (!owner || !listID) return;
	const watchlist = kind === 'watchlist';
	const ownerNeeded = source === 'trakt' || source === 'letterboxd' || (source === 'mdblist' && !watchlist);
	owner.placeholder = watchlist
	  ? (source === 'trakt' ? 'Blank for connected account, or public username' : source === 'letterboxd' ? 'Required public Letterboxd member' : `Connected ${source === 'mdblist' ? 'MDBList' : 'TMDB'} account`)
	  : (source === 'trakt' ? 'Optional Trakt username' : source === 'letterboxd' ? 'Required Letterboxd member' : source === 'mdblist' ? 'Optional MDBList username' : 'Not needed for TMDB');
	owner.disabled = !ownerNeeded;
	if (!ownerNeeded) owner.value = '';
	listID.placeholder = watchlist ? 'Not needed for watchlists' : 'Provider list ID or slug';
	listID.required = !watchlist;
	listID.disabled = watchlist;
	if (watchlist) listID.value = '';
	document.getElementById(`${scope}-letterboxd-warning`)?.classList.toggle('hidden', source !== 'letterboxd');
  }

  function parseTMDBSeeds(value) {
	return [...new Set(String(value || '').split(/[\s,]+/).filter(Boolean).map(Number).filter(Number.isInteger))];
  }

  function populateRecommendationListSources(scope, selectedSource) {
	const select = document.getElementById(`${scope}-recommendation-list-source`);
	if (!select) return;
	const current = selectedSource ?? select.value;
	const sources = jobTypes.list?.sources || [];
	select.replaceChildren(new Option('No seed list', ''), ...sources.map(source => new Option(source === 'tmdb' ? 'TMDB list' : `${capitalize(source)} list`, source)));
	if (current && !sources.includes(current)) select.add(new Option(`${capitalize(current)} list (unavailable)`, current));
	select.value = current;
  }

  function recommendationListFromForm(scope) {
	const source = document.getElementById(`${scope}-recommendation-list-source`)?.value;
	if (!source) return null;
	return { source, list: { kind: 'public_list', owner: document.getElementById(`${scope}-recommendation-list-owner`).value, list_id: document.getElementById(`${scope}-recommendation-list-id`).value, ordering: 'source' } };
  }

  async function inspectList(scope, button) {
	const source = document.getElementById(`${scope}-source`)?.value;
	const result = document.getElementById(`${scope}-list-result`);
	const locator = {
	  kind: document.getElementById(`${scope}-list-kind`)?.value,
	  owner: document.getElementById(`${scope}-list-owner`)?.value || '',
	  list_id: document.getElementById(`${scope}-list-id`)?.value || '',
	  ordering: document.getElementById(`${scope}-list-ordering`)?.value || 'source'
	};
	button.disabled = true;
	result.textContent = 'Checking source…';
	try {
	  const response = await fetch('/v1/jobs/lists/inspect', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ source, list: locator }) });
	  const data = await response.json();
	  if (!response.ok) throw new Error(data.error || 'Could not read source');
	  result.textContent = `${data.name || capitalize(source)} connected · sample: ${data.movies} movies, ${data.shows} shows`;
	} catch (error) {
	  result.textContent = error.message;
	} finally {
	  button.disabled = false;
	}
  }

  function describeRuleSet(rules) {
    const values = rules[rules.media === 'show' ? 'shows' : 'movies'] || {};
    const parts = [];
    if (values.blacklisted_min_year || values.blacklisted_max_year) parts.push(`${values.blacklisted_min_year || 'Any'}-${values.blacklisted_max_year || 'now'}`);
    if (values.min_rating) parts.push(`rating ${values.min_rating}+`);
    if (values.allowed_languages?.length) parts.push(values.allowed_languages.join(', '));
    if (values.blacklisted_genres?.length) parts.push(`blocks ${values.blacklisted_genres.slice(0, 2).join(', ')}${values.blacklisted_genres.length > 2 ? ` +${values.blacklisted_genres.length - 2}` : ''}`);
    const required = ['allowed_countries', 'allowed_languages', 'required_genres', 'required_keywords', 'required_networks'].reduce((count, key) => count + (values[key]?.length || 0), 0);
    const overrides = ['allow_countries', 'allow_languages', 'allow_genres', 'allow_keywords', 'allow_networks'].reduce((count, key) => count + (values[key]?.length || 0), 0) + (values.allow_min_rating ? 1 : 0);
    if (required) parts.push(`${required} required`);
    if (overrides) parts.push(`${overrides} override${overrides === 1 ? '' : 's'}`);
    return parts.length ? parts.join(' · ') + '.' : 'No filtering criteria.';
  }

  async function customizeJobRules(button) {
    if (!currentJob || currentJob.id.startsWith('legacy_')) return;
    button.disabled = true;
    button.textContent = 'Creating copy...';
    try {
      const response = await fetch(`/v1/jobs/${encodeURIComponent(currentJob.id)}/customize-rules`, { method: 'POST' });
      const result = await response.json();
      if (!response.ok) throw new Error(result.error || 'Could not create job rules');
      const previousRuleSetID = currentJob.rule_set_id || `default-${currentJob.media === 'show' ? 'shows' : 'movies'}`;
      const previousRules = ruleSets.find(rules => rules.id === previousRuleSetID);
      if (previousRules) previousRules.usage_count = Math.max(0, previousRules.usage_count - 1);
      ruleSets.push({ ...result.rule_set, usage_count: 1 });
      currentJob = { ...currentJob, rule_set_id: result.rule_set.id };
      const index = allJobs.findIndex(job => job.id === currentJob.id);
      if (index !== -1) allJobs[index] = currentJob;
      populateJobFilters(currentJob, false);
      renderJobsList();
      showNotification('Created rules for this job.', 'success');
    } catch (error) {
      showNotification(error.message, 'error');
    } finally {
      button.disabled = false;
      button.textContent = 'Customize for this job';
    }
  }

  function serializeJobForm() {
    const form = document.getElementById('job-config-form');
    if (!form) return '';
    return JSON.stringify([...form.elements]
      .filter((field) => field.id && !['button', 'submit', 'file'].includes(field.type))
      .map((field) => [field.id, field.type === 'checkbox' ? field.checked : field.value])
      .sort(([left], [right]) => left.localeCompare(right)));
  }

  function isJobDirty() {
    return currentJob && jobFormSnapshot !== null && serializeJobForm() !== jobFormSnapshot;
  }

  function updateJobDirtyState() {
    document.getElementById('job-editor-dirty')?.classList.toggle('hidden', !isJobDirty());
  }

  function closeJobModal(force = false) {
    if (!force && isJobDirty() && !confirm('Discard unsaved changes to this job?')) return false;
    jobModalRequest++;
    closeDialog('job-modal');
    document.getElementById('modal-advanced').classList.add('hidden');
    document.getElementById('advanced-chevron').style.transform = '';
    currentJob = null;
    jobFormSnapshot = null;
    updateJobDirtyState();
    return true;
  }

  function toggleModalAdvanced() {
    const advanced = document.getElementById('modal-advanced');
    const chevron = document.getElementById('advanced-chevron');

    if (advanced.classList.contains('hidden')) {
      advanced.classList.remove('hidden');
      chevron.style.transform = 'rotate(180deg)';
    } else {
      advanced.classList.add('hidden');
      chevron.style.transform = 'rotate(0deg)';
    }
  }

  function isDirectMode(job) {
    const modeSelect = document.getElementById('modal-mode');
    const selectedMode = modeSelect ? modeSelect.value : '';
    if (selectedMode !== '') {
      return selectedMode === 'direct';
    }
    return (job.mode || globalMode) === 'direct';
  }

  function updateDirectModeFields(job) {
    const minAvailContainer = document.getElementById('modal-min-availability-container');
    const radarrMonitorContainer = document.getElementById('modal-radarr-monitor-container');
    const sonarrMonitorContainer = document.getElementById('modal-sonarr-monitor-container');

    const directMode = isDirectMode(job);
    const isMovie = job.media === 'movie';

    // Minimum availability - only for movies in direct mode
    if (isMovie && directMode) {
      minAvailContainer.classList.remove('hidden');
      document.getElementById('modal-min-availability').value = job.minimum_availability || '';
    } else {
      minAvailContainer.classList.add('hidden');
    }

    // Monitor - Radarr for movies, Sonarr for shows, only in direct mode
    if (isMovie && directMode) {
      radarrMonitorContainer.classList.remove('hidden');
      sonarrMonitorContainer.classList.add('hidden');
      document.getElementById('modal-radarr-monitor').value = job.monitor || '';
    } else if (!isMovie && directMode) {
      sonarrMonitorContainer.classList.remove('hidden');
      radarrMonitorContainer.classList.add('hidden');
      document.getElementById('modal-sonarr-monitor').value = job.monitor || '';
    } else {
      radarrMonitorContainer.classList.add('hidden');
      sonarrMonitorContainer.classList.add('hidden');
    }
  }

  // Form submission
  document.getElementById('job-config-form').addEventListener('submit', async function(e) {
    e.preventDefault();

    if (!currentJob) return;

    const isLegacy = currentJob.id.startsWith('legacy_');
    if (isLegacy) {
      showNotification('Legacy jobs cannot be edited. Please migrate to dynamic jobs.', 'error');
      return;
    }

    // Get selected type and its definition
    const selectedType = document.getElementById('modal-type').value;
    const selectedMedia = document.getElementById('modal-media').value;
    const typeDef = jobTypes[selectedType];

    // Build updated job object with type, media, and source
    const updatedJob = {
      ...currentJob,
      name: document.getElementById('modal-name').value,
      type: selectedType,
      media: selectedMedia,
      source: document.getElementById('modal-source').value || 'trakt',
      limit: parseInt(document.getElementById('modal-limit').value),
	  delivery_limit: parseInt(document.getElementById('modal-delivery-limit').value) || 0,
	  selection_cycle: document.getElementById('modal-selection-cycle').checked,
	  minimum_picks: parseInt(document.getElementById('modal-minimum-picks').value) || 0,
	  repeat_policy: document.getElementById('modal-repeat-policy').value || '',
      sync_interval: document.getElementById('modal-interval').value || '',
      mode: document.getElementById('modal-mode').value || ''
    };
    updatedJob.rule_set_id = document.getElementById('modal-rule-set').value;
    updatedJob.use_custom_filters = false;
	updatedJob.list = selectedType === 'list' ? {
	  kind: document.getElementById('modal-list-kind').value,
	  owner: document.getElementById('modal-list-owner').value,
	  list_id: document.getElementById('modal-list-id').value,
	  ordering: document.getElementById('modal-list-ordering').value
	} : null;
	updatedJob.recommendation_seeds = selectedType === 'recommendations'
	  ? parseTMDBSeeds(document.getElementById('modal-recommendation-seeds').value)
	  : [];
	updatedJob.recommendation_list = selectedType === 'recommendations' ? recommendationListFromForm('modal') : null;
	updatedJob.series_type = selectedMedia === 'show' ? document.getElementById('modal-series-type').value : '';

    // Add period if applicable
    if (typeDef?.requires_period) {
      updatedJob.period = document.getElementById('modal-period').value;
    } else {
      // Clear period if job type doesn't require it
      updatedJob.period = '';
    }

    // Add smart job fields if applicable
    if (typeDef?.is_smart_job) {
      updatedJob.base_min_rating = parseFloat(document.getElementById('modal-base-rating').value);
      updatedJob.adjustment_factor = parseFloat(document.getElementById('modal-adjustment-factor').value);
    } else {
      // Clear smart job fields if not a smart job
      updatedJob.base_min_rating = 0;
      updatedJob.adjustment_factor = 0;
    }

    // Add Radarr/Sonarr specific fields based on the UPDATED media type
    const tempJob = { ...updatedJob };
    if (isDirectMode(tempJob)) {
      if (selectedMedia === 'movie') {
        updatedJob.minimum_availability = document.getElementById('modal-min-availability').value || '';
        updatedJob.monitor = document.getElementById('modal-radarr-monitor').value || '';
      } else {
        updatedJob.minimum_availability = ''; // Clear movie-specific field
        updatedJob.monitor = document.getElementById('modal-sonarr-monitor').value || '';
      }
    }

    try {
      const response = await fetch(`/v1/jobs/${currentJob.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updatedJob)
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Failed to save job');
      }

      const savedJob = await response.json();

      // Keep the list and modal aligned with the canonical persisted job.
      const index = allJobs.findIndex(j => j.id === currentJob.id);
      if (index !== -1) {
        allJobs[index] = savedJob;
      }
      currentJob = savedJob;
	  updateSelectionFields('modal');
      jobFormSnapshot = serializeJobForm();
      updateJobDirtyState();

      renderJobsList();
      showNotification('Job saved successfully!', 'success');
    } catch (error) {
      showNotification(error.message, 'error');
    }
  });

  async function toggleJobEnabled() {
    if (!currentJob) return;

    const isLegacy = currentJob.id.startsWith('legacy_');
    if (isLegacy) {
      showNotification('Legacy jobs cannot be modified here. Use the configuration file.', 'error');
      return;
    }

    const newEnabled = !currentJob.enabled;
    const action = newEnabled ? 'enable' : 'disable';

    if (!confirm(`${capitalize(action)} "${currentJob.name}"?`)) return;

    try {
      const updatedJob = { ...currentJob, enabled: newEnabled };
      const response = await fetch(`/v1/jobs/${currentJob.id}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(updatedJob)
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Failed to update job');
      }

      // Update local state
      const index = allJobs.findIndex(j => j.id === currentJob.id);
      if (index !== -1) {
        allJobs[index].enabled = newEnabled;
      }

      closeJobModal(true);
      renderJobsList();
      showNotification(`Job ${action}d successfully!`, 'success');
    } catch (error) {
      showNotification(error.message, 'error');
    }
  }

  async function deleteJob(jobId) {
    const job = allJobs.find(j => j.id === jobId);
    if (!job) return;

    if (job.id.startsWith('legacy_')) {
      showNotification('Legacy jobs cannot be deleted. Disable them in the configuration.', 'error');
      return;
    }

    if (!confirm(`Delete "${job.name}"? This cannot be undone.`)) return;

    try {
      const response = await fetch(`/v1/jobs/${jobId}`, {
        method: 'DELETE'
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.error || 'Failed to delete job');
      }

      // Remove from local state
      allJobs = allJobs.filter(j => j.id !== jobId);

      renderJobsList();
      showNotification('Job deleted successfully!', 'success');
    } catch (error) {
      showNotification(error.message, 'error');
    }
  }

  function deleteCurrentJob() {
    if (currentJob) {
      const jobID = currentJob.id;
      if (!closeJobModal()) return;
      deleteJob(jobID);
    }
  }

  async function triggerJob(jobId) {
    const job = allJobs.find(j => j.id === jobId);
    if (!job) return;

    try {
      const response = await fetch(`/v1/jobs/${jobId}/trigger`, {
        method: 'POST'
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Failed to trigger job');
      }

      showNotification(data.message || `${job.name} started!`, 'success');
    } catch (error) {
      showNotification(error.message, 'error');
    }
  }

  function runJobNow() {
    if (!currentJob) return;
    if (currentJob.selection_cycle === true) {
      showNotification('This job runs through the shared ranked-selection cycle.', 'error');
      return;
    }
    triggerJob(currentJob.id);
  }

  // Preview functionality
  function previewJob() {
    if (!currentJob) return;

    const job = currentJob;
    if (!closeJobModal()) return;

    setTimeout(async () => {
      currentFilter = 'all';

      document.getElementById('preview-title').textContent = job.name;
      openDialog('preview-modal', document.activeElement, closePreviewModal);

      const previewContent = document.getElementById('preview-content');
      previewData = null;
      document.getElementById('preview-stats').innerHTML = '';
      document.getElementById('preview-tabs').classList.add('hidden');
      const startedAt = performance.now();
      const previousDuration = Number(localStorage.getItem(`preview-duration:${job.id}`)) || 0;
      let estimateInterval;
      const slowTimer = setTimeout(() => {
        const updateEstimate = () => {
          const elapsed = Math.ceil((performance.now() - startedAt) / 1000);
          const estimate = document.getElementById('preview-loading-estimate');
          if (!estimate) return;
          estimate.textContent = previousDuration > elapsed
            ? `About ${Math.ceil(previousDuration - elapsed)} seconds remaining, based on the last preview.`
            : previousDuration
              ? `Taking longer than the last preview (${Math.ceil(previousDuration)} seconds).`
              : 'Checking the provider and your library can take up to a minute.';
        };
        updateEstimate();
        estimateInterval = setInterval(updateEstimate, 1000);
      }, 3000);

      previewContent.innerHTML = `
        <div class="preview-loading" role="status" aria-live="polite">
          <div class="preview-loading-spinner" aria-hidden="true"></div>
          <strong>Building decision preview</strong>
          <p id="preview-loading-estimate">Fetching candidates and checking your library…</p>
        </div>
      `;

      try {
        const response = await fetch(`/v1/jobs/${job.id}/preview`, {
          method: 'POST'
        });

        previewData = await response.json();
		if (!response.ok) throw new Error(previewData.error || 'Preview failed');
        localStorage.setItem(`preview-duration:${job.id}`, String((performance.now() - startedAt) / 1000));
        renderPreview(previewData);
      } catch (error) {
        previewContent.innerHTML = `
          <div class="text-center py-12">
            <svg class="w-16 h-16 mx-auto mb-4 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <p class="text-slate-300 text-lg font-medium">Failed to load preview</p>
            <p class="text-slate-400 text-sm mt-2">${escapeHTML(error.message)}</p>
          </div>
        `;
      } finally {
        clearTimeout(slowTimer);
        clearInterval(estimateInterval);
      }
    }, 150);
  }

  function handlePreviewPosterError(event) {
    const img = event.target;
    if (!(img instanceof HTMLImageElement) || !img.closest('.preview-row-poster')) return;
    const icon = document.createElement('i');
    icon.setAttribute('data-lucide', 'film');
    icon.setAttribute('aria-hidden', 'true');
    img.replaceWith(icon);
    window.renderLucideIcons?.(icon.parentElement);
  }

  function filterPreview(filter) {
    currentFilter = filter;

    ['all', 'will_add', 'already_exists', 'filtered_out'].forEach(f => {
      document.getElementById(`filter-${f}`)?.setAttribute('aria-selected', String(f === filter));
    });

    if (previewData) {
      renderPreview(previewData, currentFilter);
    }
  }

  function closePreviewModal() {
    closeDialog('preview-modal');
  }

  function renderPreview(data, filter) {
	const source = data.source || 'trakt';
	const sourceLabel = source === 'tmdb' ? 'TMDB' : capitalize(source);
	const simklMostWatchedURL = data.media_type === 'show' ? 'https://simkl.com/tv/best-shows/most-watched' : 'https://simkl.com/movies/best-movies/most-watched';
	document.getElementById('preview-source').innerHTML = source === 'simkl'
	  ? `Most Watched on <a class="text-blue-400 hover:underline" href="${simklMostWatchedURL}" target="_blank" rel="noopener">Simkl</a>`
	  : `Preview of what this job will fetch from ${sourceLabel}`;
    const deliveryLabel = data.mode === 'jellyseerr' ? 'request' : 'add';
    const readyLabel = `Will ${deliveryLabel}`;
    const passedFilters = data.total_found - data.filtered_out;
    document.getElementById('preview-stats').innerHTML = `
      <div><b>${data.total_found}</b><span>Found</span></div>
      <div class="preview-stat-passed"><b>${passedFilters}</b><span>Passed filters</span></div>
      <div class="preview-stat-${deliveryLabel}"><b>${data.will_add}</b><span>${readyLabel}</span></div>
      <div class="preview-stat-skipped"><b>${data.already_exists}</b><span>Skipped</span></div>
      <div class="preview-stat-rejected"><b>${data.filtered_out}</b><span>Rejected</span></div>
    `;
    document.getElementById('filter-will_add').textContent = readyLabel;
    document.getElementById('preview-tabs').classList.remove('hidden');

    let filteredItems;
    if (data.items && Array.isArray(data.items)) {
      filteredItems = data.items;
      if (filter === 'will_add') {
        filteredItems = filteredItems.filter(item => !item.already_exists && !item.filtered_out);
      } else if (filter === 'already_exists') {
        filteredItems = filteredItems.filter(item => item.already_exists);
      } else if (filter === 'filtered_out') {
        filteredItems = filteredItems.filter(item => item.filtered_out);
      }
    } else {
      filteredItems = [];
    }

    if (!filteredItems || filteredItems.length === 0) {
      document.getElementById('preview-content').innerHTML = `
        <div class="flex flex-col items-center justify-center py-12" style="min-height: 300px;">
          <svg class="w-12 h-12 mb-3 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"></path>
          </svg>
          <p class="text-slate-300 text-sm font-medium">No items found for this filter.</p>
          <p class="text-slate-500 text-xs mt-1">Try another filter or adjust your job settings.</p>
        </div>
      `;
      return;
    }

    const rowsHtml = filteredItems.map(item => {
      let badge;
      let rowClass = '';
      if (item.filtered_out) {
        badge = '<span class="preview-badge preview-badge-filtered"><i data-lucide="x" aria-hidden="true"></i>Rejected</span>';
        rowClass = 'preview-row-muted';
      } else if (item.already_exists) {
        badge = '<span class="preview-badge preview-badge-exists"><i data-lucide="circle-minus" aria-hidden="true"></i>Skipped</span>';
      } else {
        badge = `<span class="preview-badge preview-badge-${deliveryLabel}"><i data-lucide="arrow-right" aria-hidden="true"></i>${readyLabel}</span>`;
      }

      const posterHtml = data.has_posters && item.poster_url
        ? `<img src="${escapeHTML(item.poster_url)}" alt="" loading="lazy">`
        : '<i data-lucide="film" aria-hidden="true"></i>';

      const metaParts = [];
      if (item.rating) metaParts.push(`<span><i data-lucide="star" class="inline h-3 w-3 text-yellow-400" aria-hidden="true"></i> ${item.rating.toFixed(1)}${item.votes ? ` (${item.votes.toLocaleString()})` : ''}</span>`);
      if (item.tmdb_id) metaParts.push(`<a href="https://www.themoviedb.org/${item.tvdb_id ? 'tv' : 'movie'}/${item.tmdb_id}" target="_blank" rel="noopener">TMDB</a>`);
      if (item.tvdb_id) metaParts.push(`<a href="https://www.thetvdb.com/dereferrer/series/${item.tvdb_id}" target="_blank" rel="noopener">TVDB</a>`);

      const decisionTone = item.filtered_out ? 'rejected' : item.already_exists ? 'skipped' : data.mode === 'jellyseerr' ? 'requested' : 'accepted';

      return `
        <div class="preview-row ${rowClass}">
          <div class="preview-row-poster">${posterHtml}</div>
          <div class="preview-row-main">
            <div class="preview-row-title">
              <h3>${escapeHTML(item.title)}</h3>
              ${item.year ? `<span>${item.year}</span>` : ''}
            </div>
            ${metaParts.length ? `<div class="preview-row-meta">${metaParts.join('')}</div>` : ''}
            ${item.genres && item.genres.length ? `<div class="preview-row-genres">${item.genres.slice(0, 4).map(genre => `<span>${escapeHTML(genre)}</span>`).join('')}</div>` : ''}
            ${item.decision_reason ? `<p class="preview-decision preview-decision-${decisionTone}">${escapeHTML(item.decision_reason)}</p>` : item.overview ? `<p class="preview-row-overview">${escapeHTML(item.overview)}</p>` : ''}
            ${item.filter_checks && item.filter_checks.length ? `
              <details class="preview-checks">
                <summary>Rule evaluation · ${item.filter_checks.filter(check => check.passed).length}/${item.filter_checks.length} checks passed</summary>
                <div>
                  ${item.filter_checks.map(check => `
                    <p class="${check.passed ? 'preview-check-passed' : 'preview-check-failed'}">
                      <i data-lucide="${check.passed ? 'check' : 'x'}" aria-hidden="true"></i>
                      <span><strong>${escapeHTML(check.name)}</strong> ${escapeHTML(check.message)}</span>
                    </p>
                  `).join('')}
                </div>
              </details>
            ` : ''}
          </div>
          <div class="preview-row-aside">
            ${badge}
            ${item.popularity ? `<div>Popularity ${item.popularity}</div>` : ''}
            ${item.runtime ? `<div>${item.runtime}m</div>` : ''}
          </div>
        </div>
      `;
    }).join('');

    document.getElementById('preview-content').innerHTML = `<div class="preview-list">${rowsHtml}</div>`;
    window.renderLucideIcons?.(document.getElementById('preview-content'));
  }
