import React, { useEffect } from 'react';
import "./app.css";
import Router from "./router";

import { createTheme, ThemeProvider } from '@material-ui/core/styles';
import CssBaseline from '@material-ui/core/CssBaseline';
import { createBrowserHistory } from 'history'

const darkTheme = createTheme({
  palette: {
    mode: "dark",
  },
});


function App() {

  return (
    <ThemeProvider theme={darkTheme}>
      <CssBaseline />
      <Router/>
    </ThemeProvider>
  )
}

export default App;
