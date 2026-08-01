/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./web/templates/**/*.html", "./web/static/**/*.js"],
  // Activity job definitions assemble these finite class fragments at runtime.
  safelist: [
    "bg-gradient-to-r",
    "from-blue-600", "to-cyan-600", "from-purple-600", "to-pink-600",
    "from-green-600", "to-emerald-600", "from-orange-600", "to-red-600",
    "from-indigo-600", "to-purple-600", "from-teal-600", "to-blue-600",
    "text-blue-400", "text-purple-400", "text-green-400", "text-orange-400",
    "text-indigo-400", "text-teal-400", "ring-blue-500/40", "ring-purple-500/40",
    "ring-green-500/40", "ring-orange-500/40", "ring-indigo-500/40", "ring-teal-500/40",
    // Activity outcome classes are assembled from the log/run status string
    // (Go template interpolation or JS template literals), so the literal
    // class names never appear in scanned content and must be safelisted.
    "activity-status-added", "activity-status-requested", "activity-status-rejected",
    "activity-status-failed", "activity-status-skipped", "activity-status-blocked",
    "activity-reason-added", "activity-reason-requested", "activity-reason-rejected",
    "activity-reason-failed", "activity-reason-skipped", "activity-reason-blocked",
    "run-flow-found", "run-flow-passed", "run-flow-added", "run-flow-requested",
    "run-flow-rejected", "run-flow-skipped", "run-flow-failed",
    "run-distribution-added", "run-distribution-requested", "run-distribution-rejected",
    "run-distribution-skipped", "run-distribution-failed",
    "run-entry-filter-added", "run-entry-filter-requested", "run-entry-filter-rejected",
    "run-entry-filter-skipped", "run-entry-filter-failed",
    "run-row-added", "run-row-requested", "run-row-rejected", "run-row-skipped", "run-row-failed"
  ],
  theme: {
    extend: {
      colors: {
        primary: "#3b82f6",
        secondary: "#1e293b"
      },
      animation: {
        "slide-in": "slide-in 0.3s ease-out",
        "fade-in": "fade-in 0.2s ease-out"
      },
      keyframes: {
        "slide-in": {
          "0%": { transform: "translateX(100%)", opacity: "0" },
          "100%": { transform: "translateX(0)", opacity: "1" }
        },
        "fade-in": {
          "0%": { opacity: "0" },
          "100%": { opacity: "1" }
        }
      }
    }
  }
};
