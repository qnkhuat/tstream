import "./app.css";
import Router from "./router";

import { createTheme, ThemeProvider } from '@material-ui/core/styles';
import CssBaseline from '@material-ui/core/CssBaseline';
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
