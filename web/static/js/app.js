(function () {
  "use strict";

  let latestRelease;
  const dialogState = new WeakMap();
  const focusableSelector = 'button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [href], [tabindex]:not([tabindex="-1"])';

  function visibleFocusable(dialog) {
    return [...dialog.querySelectorAll(focusableSelector)].filter((element) => element.getClientRects().length > 0);
  }

  function setPageInert(dialog, inert) {
    [...document.body.children].forEach((element) => {
      if (element === dialog || element.tagName === 'SCRIPT') return;
      if (inert) {
        element.dataset.dialogWasInert = String(element.inert);
        element.inert = true;
      } else if (element.dataset.dialogWasInert !== undefined) {
        element.inert = element.dataset.dialogWasInert === 'true';
        delete element.dataset.dialogWasInert;
      }
    });
  }

  function requestDialogClose(dialog) {
    const state = dialogState.get(dialog);
    if (state?.onRequestClose) state.onRequestClose();
    else window.blockbusterrDialog.close(dialog);
  }

  window.blockbusterrDialog = {
    open(dialogOrID, options = {}) {
      const dialog = typeof dialogOrID === 'string' ? document.getElementById(dialogOrID) : dialogOrID;
      if (!dialog || !dialog.classList.contains('hidden')) return;
      const trigger = options.trigger || document.activeElement;
      if (dialog.parentElement !== document.body) document.body.appendChild(dialog);
      dialogState.set(dialog, { trigger, onRequestClose: options.onRequestClose });
      setPageInert(dialog, true);
      dialog.classList.remove('hidden');
      document.body.style.overflow = 'hidden';
      requestAnimationFrame(() => (options.initialFocus || visibleFocusable(dialog)[0])?.focus());
    },
    close(dialogOrID) {
      const dialog = typeof dialogOrID === 'string' ? document.getElementById(dialogOrID) : dialogOrID;
      if (!dialog || dialog.classList.contains('hidden')) return;
      const state = dialogState.get(dialog);
      dialog.classList.add('hidden');
      setPageInert(dialog, false);
      document.body.style.overflow = document.querySelector('.job-dialog:not(.hidden)') ? 'hidden' : '';
      dialogState.delete(dialog);
      state?.trigger?.focus?.();
    }
  };

  document.addEventListener('keydown', (event) => {
    const dialog = document.querySelector('.job-dialog:not(.hidden)');
    if (!dialog) return;
    if (event.key === 'Escape') {
      event.preventDefault();
      requestDialogClose(dialog);
      return;
    }
    if (event.key !== 'Tab') return;
    const focusable = visibleFocusable(dialog);
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
  });

  document.addEventListener('click', (event) => {
    if (event.target.matches('.job-dialog:not(.hidden)')) requestDialogClose(event.target);
    const tab = event.target.closest('[role="tab"]');
    if (!tab) return;
    const tabs = tab.closest('[role="tablist"]')?.querySelectorAll(':scope > [role="tab"]') || [];
    tabs.forEach((candidate) => { candidate.tabIndex = candidate === tab ? 0 : -1; });
    const panel = document.getElementById(tab.getAttribute('aria-controls'));
    if (panel && tab.id) panel.setAttribute('aria-labelledby', tab.id);
  });

  document.addEventListener('keydown', (event) => {
    const current = event.target.closest('[role="tab"]');
    const tablist = current?.closest('[role="tablist"]');
    if (!tablist || !['ArrowLeft', 'ArrowRight', 'Home', 'End'].includes(event.key)) return;
    const tabs = [...tablist.querySelectorAll(':scope > [role="tab"]:not([disabled])')];
    const index = tabs.indexOf(current);
    if (index < 0) return;
    event.preventDefault();
    const next = event.key === 'Home' ? tabs[0]
      : event.key === 'End' ? tabs[tabs.length - 1]
      : tabs[(index + (event.key === 'ArrowRight' ? 1 : -1) + tabs.length) % tabs.length];
    next.focus();
    next.click();
  });

  window.togglePassword = function (inputId, button) {
    const input = document.getElementById(inputId);
    const eyeOpen = button.querySelector(".eye-open");
    const eyeClosed = button.querySelector(".eye-closed");
    const showing = input.type === "password";

    input.type = showing ? "text" : "password";
    eyeOpen.classList.toggle("hidden", showing);
    eyeClosed.classList.toggle("hidden", !showing);
    button.setAttribute("aria-pressed", String(showing));
  };

  window.showNotification = function (message, type = "success") {
    const container = document.getElementById("notification-container");
    const notification = document.createElement("div");
    const icon = document.createElementNS("http://www.w3.org/2000/svg", "svg");
    const path = document.createElementNS("http://www.w3.org/2000/svg", "path");
    const text = document.createElement("span");
    const close = document.createElement("button");

    notification.className = `${type === "success" ? "bg-green-500" : type === "error" ? "bg-red-500" : "bg-yellow-500"} text-white px-6 py-3 rounded-lg shadow-lg flex items-center gap-3 min-w-[300px] transform transition-all duration-300 translate-x-full`;
    notification.setAttribute("role", type === "error" ? "alert" : "status");
    icon.setAttribute("class", "w-5 h-5");
    icon.setAttribute("fill", "none");
    icon.setAttribute("stroke", "currentColor");
    icon.setAttribute("viewBox", "0 0 24 24");
    path.setAttribute("stroke-linecap", "round");
    path.setAttribute("stroke-linejoin", "round");
    path.setAttribute("stroke-width", "2");
    path.setAttribute("d", type === "success" ? "M5 13l4 4L19 7" : "M6 18L18 6M6 6l12 12");
    text.className = "flex-1";
    // Messages may contain upstream error text, so keep them out of innerHTML.
    text.textContent = String(message);
    close.type = "button";
    close.className = "hover:bg-white/20 rounded p-1";
    close.setAttribute("aria-label", "Dismiss notification");
    close.innerHTML = '<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24" aria-hidden="true"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"></path></svg>';
    close.addEventListener("click", () => notification.remove());
    icon.appendChild(path);
    notification.append(icon, text, close);
    container.appendChild(notification);

    setTimeout(() => notification.classList.remove("translate-x-full"), 10);
    setTimeout(() => {
      notification.classList.add("translate-x-full");
      setTimeout(() => notification.remove(), 300);
    }, 5000);
  };

  window.renderLucideIcons = function (root) {
    if (!window.lucide || typeof window.lucide.createIcons !== "function") return;
    if (root && root.querySelectorAll) {
      window.lucide.createIcons({ nodes: root.querySelectorAll("[data-lucide]") });
      return;
    }
    window.lucide.createIcons();
  };

  function initialize(root) {
    window.renderLucideIcons(root);

    root.querySelectorAll('[role="tablist"]').forEach((tablist) => {
      const tabs = [...tablist.querySelectorAll(':scope > [role="tab"]')];
      tabs.forEach((tab, index) => { tab.tabIndex = tab.getAttribute('aria-selected') === 'true' || (!tabs.some((item) => item.getAttribute('aria-selected') === 'true') && index === 0) ? 0 : -1; });
    });

    const versionLinks = [...root.querySelectorAll('[data-latest-version]')];
    if (versionLinks.length) {
      latestRelease ||= fetch('https://api.github.com/repos/mahcks/blockbusterr/releases/latest')
        .then((response) => {
          if (!response.ok) throw new Error('release lookup failed');
          return response.json();
        });
      latestRelease
        .then((release) => versionLinks.forEach((link) => {
          link.textContent = release.tag_name || 'unavailable';
          if (release.html_url) link.href = release.html_url;
        }))
        .catch(() => versionLinks.forEach((link) => { link.textContent = 'unavailable'; }));
    }

    root.querySelectorAll('[id^="info-alert-"]').forEach((alert) => {
      const id = alert.id.replace('info-alert-', '');
      if (!localStorage.getItem(`infoAlertDismissed_${id}`)) alert.style.display = '';
    });
  }

  document.addEventListener('click', (event) => {
    const button = event.target.closest('[data-dismiss-info-alert]');
    if (!button) return;
    const id = button.dataset.dismissInfoAlert;
    localStorage.setItem(`infoAlertDismissed_${id}`, 'true');
    document.getElementById(`info-alert-${id}`)?.remove();
  });

  // HTMX swaps do not rerun DOMContentLoaded, so initialize each injected subtree.
  document.addEventListener("DOMContentLoaded", () => initialize(document));
  document.addEventListener("htmx:afterSwap", (event) => initialize(event.detail.target));
})();
