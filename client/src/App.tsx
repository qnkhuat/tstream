import React, { useEffect, useState } from 'react';
import Home from './pages/home';
import Room from './pages/[room]';
import {
  BrowserRouter as Router,
  Switch,
  Route,
  Link
} from "react-router-dom";


function App() {
  return (
    <>
      <Router>
        <Switch>
          <Route path="/:username" ><Room/></Route>
          <Route path="/" ><Home/></Route>
        </Switch>
      </Router>
    </>
  )
}


export default App;
