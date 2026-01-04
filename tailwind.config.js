/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./templates/**/*.templ", "./templates/**/*_templ.go"],
  theme: {
    extend: {
      colors: {
        "hn-orange": "#ff6600",
        "hn-beige": "#f6f6ef",
      },
    },
  },
  plugins: [],
};
