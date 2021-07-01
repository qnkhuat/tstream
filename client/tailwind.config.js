module.exports = {
  purge: ['./src/**/*.{js,jsx,ts,tsx}', './public/index.html'],
  darkMode: false, // or 'media' or 'class'
  theme: {
    extend: {
      colors:{
        green: {
          "term": "#00c200",
        }
      }
    },
  },
  variants: {
    extend: {},
  },
  plugins: [],
}
