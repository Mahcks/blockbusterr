  // Owns Jobs page state and the /v1/jobs/* interactions. The module stays
  // global-compatible while templates migrate away from inline handlers.
  let allJobs = [];
  let jobTypes = {};
  let jobTemplates = [];
  let ruleSets = [];
  let currentJob = null;
  let currentTemplateCategory = 'Movies';
  let templateSearchQuery = '';
  let previewData = null;
  let currentFilter = 'all';
  const dialogTriggers = [];
  const globalMode = document.getElementById("jobs-page")?.dataset.globalMode || "direct";

  const dialogIds = ['preview-modal', 'job-modal', 'add-job-modal'];
  const focusableSelector = 'button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [href], [tabindex]:not([tabindex="-1"])';

  function openDialog(id, trigger = document.activeElement) {
    const dialog = document.getElementById(id);
    if (!dialog) return;
    dialogTriggers.push(trigger);
    dialog.classList.remove('hidden');
    document.body.style.overflow = 'hidden';
    requestAnimationFrame(() => dialog.querySelector(focusableSelector)?.focus());
  }

  function closeDialog(id) {
    document.getElementById(id)?.classList.add('hidden');
    if (!dialogIds.some(dialogId => !document.getElementById(dialogId)?.classList.contains('hidden'))) {
      document.body.style.overflow = '';
    }
    dialogTriggers.pop()?.focus();
  }

  function handleDialogKeyboard(event) {
    const dialog = dialogIds.map(id => document.getElementById(id)).find(element => element && !element.classList.contains('hidden'));
    if (!dialog) return;
    if (event.key === 'Escape') {
      event.preventDefault();
      if (dialog.id === 'add-job-modal') closeAddJobModal();
      if (dialog.id === 'job-modal') closeJobModal();
      if (dialog.id === 'preview-modal') closePreviewModal();
      return;
    }
    if (event.key !== 'Tab') return;
    const focusable = [...dialog.querySelectorAll(focusableSelector)].filter(element => element.offsetParent !== null);
    if (!focusable.length) return;
    const first = focusable[0];
    const last = focusable[focusable.length - 1];
    if (event.shiftKey && document.activeElement === first) {
      event.preventDefault();
      last.focus();
    } else if (!event.shiftKey && document.activeElement === last) {
      event.preventDefault();
      first.focus();
    }
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
      smart_popular: 'brain'
    };
    const iconName = iconNames[type] || iconNames.popular;
    return `<i data-lucide="${iconName}" class="${cls}"></i>`;
  }

  const typeStyles = {
    'trending': { color: 'text-yellow-400', bg: 'from-yellow-900/50' },
    'popular': { color: 'text-red-400', bg: 'from-red-900/50' },
    'watched': { color: 'text-cyan-400', bg: 'from-cyan-900/50' },
    'collected': { color: 'text-amber-400', bg: 'from-amber-900/50' },
    'favorited': { color: 'text-pink-400', bg: 'from-pink-900/50' },
    'played': { color: 'text-blue-400', bg: 'from-blue-900/50' },
    'anticipated': { color: 'text-purple-400', bg: 'from-purple-900/50' },
    'box_office': { color: 'text-green-400', bg: 'from-green-900/50' },
    'smart_popular': { color: 'text-indigo-400', bg: 'from-indigo-900/50' }
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
    document.getElementById('custom-media')?.addEventListener('change', updateCustomRuleSets);
    document.getElementById('modal-type')?.addEventListener('change', onJobTypeChange);
    document.getElementById('modal-source')?.addEventListener('change', updateModalTypeFields);
    document.getElementById('modal-rule-set')?.addEventListener('change', updateRuleSetSummary);
    document.addEventListener('keydown', handleDialogKeyboard);
    document.addEventListener('click', handleJobsAction);

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
        if (template) createJobFromTemplate(template.type, template.media, template.name, template.limit, template.period || '');
      },
      'close-preview': () => closePreviewModal(),
      'filter-preview': () => filterPreview(target.dataset.filter),
      'close-job': () => closeJobModal(),
      'toggle-advanced': () => toggleModalAdvanced(),
      'preview-job': () => previewJob(),
      'run-job': () => runJobNow(),
      'toggle-job': () => toggleJobEnabled(),
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
	const enabledJobs = allJobs.filter(job => job.enabled);
	const activeJobs = enabledJobs.filter(isJobAvailable);
	const unavailableJobs = enabledJobs.filter(job => !isJobAvailable(job));

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
                  ${unavailable ? '<span class="text-xs font-medium text-amber-300">Setup required</span>' : ''}
                </span>
                <span class="mt-0.5 block truncate text-xs text-slate-500">${escapeHTML(capitalize(job.type))} · ${job.limit} items</span>
              </span>
            </button>
            <span class="job-row-value">${escapeHTML(sourceLabel)}</span>
            <span class="job-row-value">${job.media === 'movie' ? 'Movie' : 'TV show'}</span>
            <button type="button" class="job-filter-policy" data-action="open-job-filters" data-job-id="${escapeHTML(job.id)}" ${isLegacy ? 'disabled' : ''}>
              <span>${escapeHTML(assignedRules?.name || 'Rules unavailable')}</span>
              <small>${assignedRules ? `Revision ${assignedRules.revision}` : 'Check assignment'}</small>
            </button>
            <div class="job-row-actions">
				  ${unavailable ? '' : `<button
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

    if (legacyJobs.length > 0) {
      countEl.textContent = legacyJobs.length;
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
    if (!confirm('This will convert all legacy jobs to the new dynamic format. Your job settings will be preserved. Continue?')) {
      return;
    }

    try {
      const response = await fetch('/v1/jobs/migrate', {
        method: 'POST'
      });

      const data = await response.json();

      if (!response.ok) {
        throw new Error(data.error || 'Migration failed');
      }

      if (data.count === 0) {
        showNotification('No legacy jobs to migrate', 'info');
      } else {
        showNotification(`Successfully migrated ${data.count} jobs!`, 'success');
        // Reload jobs data
        await loadData();
        renderJobsList();
      }

      // Hide the banner
      document.getElementById('migration-banner').classList.add('hidden');
    } catch (error) {
      showNotification(error.message, 'error');
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
    openDialog('add-job-modal', trigger);
  }

  function closeAddJobModal() {
    closeDialog('add-job-modal');
    // Reset custom form
    document.getElementById('custom-job-form').reset();
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
    typeSelect.innerHTML = Object.entries(jobTypes).filter(([, def]) => def.sources?.length).map(([key, def]) => {
      return `<option value="${key}">${def.name}</option>`;
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

    const selectedType = typeSelect.value;
    const typeDef = jobTypes[selectedType];

    if (!typeDef) return;
	const currentSource = sourceSelect.value;
	const sources = typeDef.sources || [typeDef.source || 'trakt'];
	sourceSelect.innerHTML = sources.map(source => `<option value="${source}">${source === 'tmdb' ? 'TMDB' : capitalize(source)}</option>`).join('');
	if (sources.includes(currentSource)) sourceSelect.value = currentSource;
	document.getElementById('custom-source-help').textContent = sourceSelect.value === 'simkl'
	  ? 'Simkl attribution will be shown with sourced results.'
	  : 'Only configured discovery sources are shown.';
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
    mediaSelect.innerHTML = '';
    if (supportedMedia.includes('movie')) {
      mediaSelect.innerHTML += '<option value="movie">Movie</option>';
    }
    if (supportedMedia.includes('show')) {
      mediaSelect.innerHTML += '<option value="show">TV Show</option>';
    }

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
		  source: formData.get('source') || 'trakt',
          rule_set_id: formData.get('rule_set_id')
        };

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
	  .filter(t => jobTypes[t.type]?.sources?.length)
      .filter(t => t.category === currentTemplateCategory)
      .filter(t => {
        if (!templateSearchQuery) return true;
        const name = String(t.name || '').toLowerCase();
        const type = String(t.type || '').toLowerCase();
        const description = String(t.description || '').toLowerCase();
        return name.includes(templateSearchQuery) || type.includes(templateSearchQuery) || description.includes(templateSearchQuery);
      })
      .sort((a, b) => String(a.name || '').localeCompare(String(b.name || '')));

    if (count) {
      count.textContent = `${filtered.length} template${filtered.length === 1 ? '' : 's'} available`;
    }

    if (!filtered.length) {
      grid.innerHTML = `
        <div class="p-6 text-center">
          <p class="text-sm font-medium text-slate-200">No templates found</p>
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

      return `
        <button
          type="button"
          data-action="create-from-template"
          data-template-index="${jobTemplates.indexOf(template)}"
          class="template-row"
        >
          <div class="flex items-center gap-3">
            <span class="job-row-icon ${style.color}">${getTypeIcon(template.type, 'w-4 h-4')}</span>
            <div class="flex-1">
              <div class="flex items-center justify-between gap-3">
                <h3 class="text-sm font-medium text-slate-100">${escapeHTML(template.name)}</h3>
                <span class="shrink-0 text-xs text-slate-500">${template.limit} items${template.period ? ` · ${escapeHTML(capitalize(template.period))}` : ''}</span>
              </div>
              <p class="mt-0.5 text-xs text-slate-500">${escapeHTML(template.description)}</p>
            </div>
          </div>
        </button>
      `;
    }).join('');

    if (window.renderLucideIcons) {
      window.renderLucideIcons(grid);
    }
  }

  async function createJobFromTemplate(type, media, name, limit, period) {
	const source = jobTypes[type]?.sources?.[0];
	if (!source) {
	  showNotification('Configure a discovery source that supports this job type first.', 'error');
	  return;
	}
    const job = {
      name: name,
      type: type,
      media: media,
      enabled: true,
      limit: limit,
	  source,
      rule_set_id: `default-${media === 'show' ? 'shows' : 'movies'}`
    };

    if (period) {
      job.period = period;
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
      showNotification(`${name} created with the default ${media === 'show' ? 'show' : 'movie'} rules.`, 'success');

      // Open the edit modal for the new job
      openJobModal(newJob.id);
    } catch (error) {
      showNotification(error.message, 'error');
    }
  }

  // Job Configuration Modal
  function openJobModal(jobId, trigger) {
    const job = allJobs.find(j => j.id === jobId);
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
    document.getElementById('modal-interval').value = job.sync_interval || '';
    document.getElementById('modal-mode').value = job.mode || '';
    document.getElementById('modal-source').value = job.source || 'trakt';

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
    } else {
      deleteZone.classList.remove('hidden');
    }

    openDialog('job-modal', trigger);
  }

  function openJobFilters(jobId, trigger) {
    openJobModal(jobId, trigger);
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

    const selectedType = typeSelect.value;
    const typeDef = jobTypes[selectedType];

    if (!typeDef) return;
	const currentSource = sourceSelect.value || currentJob?.source || 'trakt';
	const sources = typeDef.sources || [typeDef.source || 'trakt'];
	sourceSelect.innerHTML = sources.map(source => `<option value="${source}">${source === 'tmdb' ? 'TMDB' : capitalize(source)}</option>`).join('');
	if (sources.includes(currentSource)) sourceSelect.value = currentSource;
	document.getElementById('modal-source-help').textContent = sources.length
	  ? (sourceSelect.value === 'simkl' ? 'Simkl attribution will be shown with sourced results.' : 'Only configured discovery sources are shown.')
	  : 'This job requires a discovery source that is not configured. Change its type or delete it.';
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
      if (!baseRating.value || baseRating.value === '0') {
        baseRating.value = currentJob?.base_min_rating || 6.0;
      }
      if (!adjustmentFactor.value || adjustmentFactor.value === '0') {
        adjustmentFactor.value = currentJob?.adjustment_factor || 0.5;
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

  function describeRuleSet(rules) {
    const values = rules[rules.media === 'show' ? 'shows' : 'movies'] || {};
    const parts = [];
    if (values.blacklisted_min_year || values.blacklisted_max_year) parts.push(`${values.blacklisted_min_year || 'Any'}-${values.blacklisted_max_year || 'now'}`);
    if (values.min_rating) parts.push(`rating ${values.min_rating}+`);
    if (values.allowed_languages?.length) parts.push(values.allowed_languages.join(', '));
    if (values.blacklisted_genres?.length) parts.push(`blocks ${values.blacklisted_genres.slice(0, 2).join(', ')}${values.blacklisted_genres.length > 2 ? ` +${values.blacklisted_genres.length - 2}` : ''}`);
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

  function closeJobModal() {
    closeDialog('job-modal');
    document.getElementById('modal-advanced').classList.add('hidden');
    document.getElementById('advanced-chevron').style.transform = '';
    currentJob = null;
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
      sync_interval: document.getElementById('modal-interval').value || '',
      mode: document.getElementById('modal-mode').value || ''
    };
    updatedJob.rule_set_id = document.getElementById('modal-rule-set').value;
    updatedJob.use_custom_filters = false;

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

      // Update local state
      const index = allJobs.findIndex(j => j.id === currentJob.id);
      if (index !== -1) {
        allJobs[index] = updatedJob;
      }
      currentJob = updatedJob;

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

      closeJobModal();
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
      closeJobModal();
      deleteJob(currentJob.id);
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
    if (currentJob) {
      triggerJob(currentJob.id);
    }
  }

  // Preview functionality
  function previewJob() {
    if (!currentJob) return;

    const job = currentJob;
    closeJobModal();

    setTimeout(async () => {
      currentFilter = 'all';

      document.getElementById('preview-title').textContent = job.name;
      openDialog('preview-modal');

      document.getElementById('preview-content').innerHTML = `
        <div class="flex items-center justify-center py-12">
          <div class="h-12 w-12 animate-spin rounded-full border-b-2 border-blue-400"></div>
        </div>
      `;

      try {
        const response = await fetch(`/v1/jobs/${job.id}/preview`, {
          method: 'POST'
        });

        previewData = await response.json();
        renderPreview(previewData);
      } catch (error) {
        document.getElementById('preview-content').innerHTML = `
          <div class="text-center py-12">
            <svg class="w-16 h-16 mx-auto mb-4 text-red-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z"></path>
            </svg>
            <p class="text-slate-300 text-lg font-medium">Failed to load preview</p>
            <p class="text-slate-400 text-sm mt-2">${escapeHTML(error.message)}</p>
          </div>
        `;
      }
    }, 150);
  }

  function filterPreview(filter) {
    currentFilter = filter;

    ['all', 'will_add', 'already_exists', 'filtered_out'].forEach(f => {
      const btn = document.getElementById(`filter-${f}`);
      if (f === filter) {
        btn.className = 'rounded-md bg-red-600 px-3 py-1.5 text-sm font-medium text-white';
      } else {
        btn.className = 'rounded-md bg-transparent px-3 py-1.5 text-sm font-medium text-slate-400 hover:bg-zinc-800 hover:text-slate-100';
      }
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
    const willAdd = data.total_found - data.already_exists - data.filtered_out;
    document.getElementById('preview-stats').innerHTML = `
      <div class="rounded-lg border border-slate-700 bg-slate-800 p-3">
        <div class="text-xl font-semibold text-white">${data.total_found}</div>
        <div class="mt-1 text-xs text-slate-400">Found</div>
      </div>
      <div class="rounded-lg border border-green-800 bg-green-950/40 p-3">
        <div class="text-xl font-semibold text-green-200">${willAdd}</div>
        <div class="mt-1 text-xs text-green-300">Will add</div>
      </div>
      <div class="rounded-lg border border-slate-700 bg-slate-800 p-3">
        <div class="text-xl font-semibold text-slate-200">${data.already_exists}</div>
        <div class="mt-1 text-xs text-slate-400">Already exists</div>
      </div>
      <div class="rounded-lg border border-red-900 bg-red-950/30 p-3">
        <div class="text-xl font-semibold text-red-200">${data.filtered_out}</div>
        <div class="mt-1 text-xs text-red-300">Filtered out</div>
      </div>
    `;

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
          <svg class="w-16 h-16 mb-4 text-slate-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M20 13V6a2 2 0 00-2-2H6a2 2 0 00-2 2v7m16 0v5a2 2 0 01-2 2H6a2 2 0 01-2-2v-5m16 0h-2.586a1 1 0 00-.707.293l-2.414 2.414a1 1 0 01-.707.293h-3.172a1 1 0 01-.707-.293l-2.414-2.414A1 1 0 006.586 13H4"></path>
          </svg>
          <p class="text-slate-400 text-lg font-semibold">No items found for this filter.</p>
          <p class="text-slate-500 text-sm mt-2">Try another filter or adjust your job settings.</p>
        </div>
      `;
      return;
    }

    const itemsHtml = filteredItems.map(item => {
      let statusBadge = '';
      let statusClass = '';

      if (item.filtered_out) {
        statusBadge = `<span class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-red-500/20 text-red-300 rounded-lg text-xs font-semibold ring-1 ring-red-500/30"><svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>Filtered</span>`;
        statusClass = 'opacity-60';
      } else if (item.already_exists) {
        statusBadge = `<span class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-blue-500/20 text-blue-300 rounded-lg text-xs font-semibold ring-1 ring-blue-500/30"><svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>Exists</span>`;
        statusClass = 'opacity-85';
      } else {
        statusBadge = `<span class="inline-flex items-center gap-1.5 px-3 py-1.5 bg-green-500/20 text-green-300 rounded-lg text-xs font-semibold ring-1 ring-green-500/30"><svg class="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 5v14m-7-7h14"></path></svg>Will Add</span>`;
      }

      const posterHtml = data.has_posters && item.poster_url
        ? `<img src="${escapeHTML(item.poster_url)}" alt="${escapeHTML(item.title)}" class="w-full h-full object-cover" onerror="this.parentElement.innerHTML='<div class=\\'flex items-center justify-center h-full bg-slate-700\\'><svg class=\\'w-8 h-8 text-slate-500\\' fill=\\'currentColor\\' viewBox=\\'0 0 20 20\\'><path d=\\'M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4z\\' /></svg></div>'">`
        : `<div class="flex items-center justify-center h-full bg-slate-700"><svg class="w-8 h-8 text-slate-500" fill="currentColor" viewBox="0 0 20 20"><path d="M4 3a2 2 0 00-2 2v10a2 2 0 002 2h12a2 2 0 002-2V5a2 2 0 00-2-2H4z" /></svg></div>`;

      return `
        <div class="overflow-hidden rounded-lg border border-slate-700/80 bg-slate-900 transition-colors hover:border-slate-600 ${statusClass}">
          <div class="flex gap-4 p-4">
            <div class="h-36 w-24 flex-shrink-0 overflow-hidden rounded-lg bg-slate-700 ring-1 ring-slate-600">
              ${posterHtml}
            </div>
            <div class="flex min-w-0 flex-1 flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
              <div class="min-w-0">
                <h3 class="text-lg font-semibold text-white truncate">${escapeHTML(item.title)}</h3>
                <div class="flex items-center gap-2 mt-1 text-sm text-slate-300 flex-wrap">
                  ${item.year ? `<span>${item.year}</span>` : ''}
                  ${item.rating ? `<span class="flex items-center gap-1 text-slate-200">
                    <svg class="w-4 h-4 text-yellow-400" fill="currentColor" viewBox="0 0 20 20"><path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z"></path></svg>
                    ${item.rating.toFixed(1)}
                  </span>` : ''}
                  ${item.votes ? `<span class="text-slate-400">(${item.votes.toLocaleString()} votes)</span>` : ''}
                  ${item.tmdb_id ? `<a href="https://www.themoviedb.org/${item.tvdb_id ? 'tv' : 'movie'}/${item.tmdb_id}" target="_blank" class="inline-flex items-center gap-1 text-xs font-medium text-blue-400 hover:text-blue-300 transition-colors">
                    <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20"><path d="M11 3a1 1 0 100 2h2.586l-6.293 6.293a1 1 0 101.414 1.414L15 6.414V9a1 1 0 102 0V4a1 1 0 00-1-1h-5z"></path><path d="M5 5a2 2 0 00-2 2v8a2 2 0 002 2h8a2 2 0 002-2v-3a1 1 0 10-2 0v3H5V7h3a1 1 0 000-2H5z"></path></svg>
                    TMDB
                  </a>` : ''}
                  ${item.tvdb_id ? `<a href="https://www.thetvdb.com/dereferrer/series/${item.tvdb_id}" target="_blank" class="inline-flex items-center gap-1 text-xs font-medium text-cyan-400 hover:text-cyan-300 transition-colors">
                    <svg class="w-3.5 h-3.5" fill="currentColor" viewBox="0 0 20 20"><path d="M11 3a1 1 0 100 2h2.586l-6.293 6.293a1 1 0 101.414 1.414L15 6.414V9a1 1 0 102 0V4a1 1 0 00-1-1h-5z"></path><path d="M5 5a2 2 0 00-2 2v8a2 2 0 002 2h8a2 2 0 002-2v-3a1 1 0 10-2 0v3H5V7h3a1 1 0 000-2H5z"></path></svg>
                    TVDB
                  </a>` : ''}
                </div>
                ${item.genres && item.genres.length > 0 ? `
                  <div class="flex flex-wrap gap-1 mt-2">
                    ${item.genres.slice(0, 4).map(genre => `<span class="px-2 py-0.5 text-xs rounded bg-slate-700/60 text-slate-200 ring-1 ring-slate-600/40">${escapeHTML(genre)}</span>`).join('')}
                  </div>
                ` : ''}
                ${item.overview ? `<p class="text-sm text-slate-300/90 mt-3 line-clamp-2">${escapeHTML(item.overview)}</p>` : ''}
                ${item.filter_reason && item.filtered_out ? `<p class="text-xs text-red-300/90 mt-2">Reason: ${escapeHTML(item.filter_reason)}</p>` : ''}
              </div>
              <div class="flex flex-col items-start lg:items-end gap-2">
                ${statusBadge}
                ${item.popularity ? `
                  <div class="text-xs text-slate-300/80">Popularity: ${item.popularity}</div>
                ` : ''}
                ${item.runtime ? `
                  <div class="text-xs text-slate-300/80">Runtime: ${item.runtime}m</div>
                ` : ''}
              </div>
            </div>
          </div>
        </div>
      `;
    }).join('');

    document.getElementById('preview-content').innerHTML = `
      <div class="space-y-3">
        ${itemsHtml}
      </div>
    `;
  }
