/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{ts,tsx}"],
  theme: {
    extend: {
      colors: {
        board: {
          bg: "#f5f3ee",
          panel: "#fff9ed",
          ink: "#1a2e2c",
          accent: "#e2711d",
          accentSoft: "#f9c784",
        },
      },
      boxShadow: {
        card: "0 12px 24px -12px rgba(0, 0, 0, 0.35)",
      },
    },
  },
  plugins: [],
};
