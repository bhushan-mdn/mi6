/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    // Watch all Go and templ files for Tailwind classes
    "./**/*.{go,templ,html}",
  ],
  theme: {
    extend: {},
  },
  plugins: [
    require('daisyui'),
  ],
  // DaisyUI configuration
  daisyui: {
    themes: ["dark", "light"], // Use a couple of themes
  },
}
