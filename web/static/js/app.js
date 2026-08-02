(function () {
  "use strict";

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
