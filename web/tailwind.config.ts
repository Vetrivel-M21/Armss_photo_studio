import type { Config } from "tailwindcss";

// ARMSS brand palette: maroon/red/coral/yellow with a warm cream surface,
// plus an independent danger scale kept deliberately distinct from the brand
// hues so errors never look brand-colored. Poster previews are meant to be
// the only loud thing on the page, so UI chrome leans on ink/cream/neutral
// tones with the brand colors as accents.
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        brand: {
          primary: "#FF2A5F", // Neon Crimson
          secondary: "#E11D48", // Radiant Ruby
          accent: "#FB7185", // Electric Rose
          violet: "#8B5CF6", // Electric Violet
          highlight: "#FBBF24", // Vibrant Amber
          dark: "#090B10", // Obsidian Dark Canvas
          surface: "#11141E", // Dark Glass Surface
          surfaceHover: "#181D2B", // Elevated Surface
          border: "#202637", // Crisp Dark Border
        },
        dark: {
          50: "#F8FAFC",
          100: "#F1F5F9",
          200: "#E2E8F0",
          300: "#94A3B8",
          400: "#64748B",
          500: "#475569",
          600: "#334155",
          700: "#1E293B",
          800: "#141926",
          900: "#0E121C",
          950: "#090B10",
        },
        danger: {
          DEFAULT: "#EF4444",
          tint: "rgba(239, 68, 68, 0.12)",
        },
      },
      fontFamily: {
        sans: ['"Plus Jakarta Sans"', '"Inter"', "system-ui", "sans-serif"],
        serif: ['"Fraunces"', "Georgia", "serif"],
      },
      boxShadow: {
        glow: "0 0 30px -5px rgba(255, 42, 95, 0.35)",
        "glow-sm": "0 0 15px -3px rgba(255, 42, 95, 0.25)",
        "glow-violet": "0 0 30px -5px rgba(139, 92, 246, 0.35)",
        "glow-amber": "0 0 30px -5px rgba(251, 191, 36, 0.35)",
        glass: "0 8px 32px 0 rgba(0, 0, 0, 0.45)",
      },
      keyframes: {
        gradientShift: {
          "0%, 100%": { backgroundPosition: "0% 50%" },
          "50%": { backgroundPosition: "100% 50%" },
        },
        shimmer: {
          "0%": { transform: "translateX(-100%)" },
          "100%": { transform: "translateX(100%)" },
        },
        float: {
          "0%, 100%": { transform: "translateY(0)" },
          "50%": { transform: "translateY(-6px)" },
        },
        pulseGlow: {
          "0%, 100%": { opacity: "0.6", transform: "scale(1)" },
          "50%": { opacity: "0.9", transform: "scale(1.05)" },
        },
      },
      animation: {
        "gradient-shift": "gradientShift 8s ease infinite",
        shimmer: "shimmer 1.8s linear infinite",
        float: "float 4s ease-in-out infinite",
        "pulse-glow": "pulseGlow 3s ease-in-out infinite",
      },
    },
  },
  plugins: [require("tailwindcss-animate")],
} satisfies Config;
