import "./app.css";
import Router from "./router";

import { createTheme, ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import termlog from "termlog";
termlog();

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
