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

  window.toggleAdvanced = function (jobName, button) {
    const advancedDiv = document.getElementById(`${jobName}-advanced`);
    const icon = button.querySelector("svg");
    const opening = advancedDiv.classList.contains("hidden");

    advancedDiv.classList.toggle("hidden", !opening);
    icon.style.transform = opening ? "rotate(180deg)" : "rotate(0deg)";
    button.setAttribute("aria-expanded", String(opening));
  };

  window.testConnection = function (service, button) {
    const form = document.querySelector("form");
    const formData = new FormData(form);
    const services = {
      trakt: ["/v1/trakt/validate", "trakt.client_id", "trakt.client_secret", "Trakt Client ID is required"],
      radarr: ["/v1/radarr/validate", "radarr.url", "radarr.api_key", "Radarr URL and API Key are required"],
      sonarr: ["/v1/sonarr/validate", "sonarr.url", "sonarr.api_key", "Sonarr URL and API Key are required"],
      jellyseerr: ["/v1/jellyseerr/validate", "jellyseerr.url", "jellyseerr.api_key", "Jellyseerr URL and API Key are required"]
    };
    const config = services[service];
    if (!config) return;

    const first = String(formData.get(config[1]) || "");
    const second = String(formData.get(config[2]) || "");
    if (!first.trim() || (service !== "trakt" && !second.trim())) {
      window.showNotification(config[3], "error");
      return;
    }

    const params = service === "trakt"
      ? { client_id: first, client_secret: second, test: "true" }
      : { url: first, api_key: second, test: "true" };
    const originalText = button.innerHTML;
    button.disabled = true;
    button.innerHTML = '<svg class="animate-spin h-5 w-5 mx-auto" fill="none" viewBox="0 0 24 24" aria-label="Testing connection"><circle class="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4"></circle><path class="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path></svg>';

    fetch(`${config[0]}?${new URLSearchParams(params)}`)
      .then((response) => response.json())
      .then((data) => {
        if (data.connected || data.message) {
          window.showNotification(data.message || `${service.charAt(0).toUpperCase() + service.slice(1)} connection successful!`, "success");
        } else if (data.error) {
          window.showNotification(data.error, "error");
        }
      })
      .catch((error) => window.showNotification(`Failed to connect to ${service}: ${error.message}`, "error"))
      .finally(() => {
        button.disabled = false;
        button.innerHTML = originalText;
      });
  };

  function initialize(root) {
    window.renderLucideIcons(root);
  }

  // HTMX swaps do not rerun DOMContentLoaded, so initialize each injected subtree.
  document.addEventListener("DOMContentLoaded", () => initialize(document));
  document.addEventListener("htmx:afterSwap", (event) => initialize(event.detail.target));
})();
