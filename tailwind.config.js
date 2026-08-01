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
    "ring-green-500/40", "ring-orange-500/40", "ring-indigo-500/40", "ring-teal-500/40"
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
