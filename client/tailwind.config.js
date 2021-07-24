module.exports = {
  important:true,
  purge: ['./src/**/*.{js,jsx,ts,tsx}', './public/index.html'],
  darkMode: 'class', // or 'media' or 'class'
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
    extend: {
      display: ['group-hover'],
    }
  },
  plugins: [
    require('@tailwindcss/typography'),
  ],
}
