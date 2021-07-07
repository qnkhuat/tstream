import {
  BrowserRouter,
  Switch,
  Route
} from "react-router-dom";

import Home from './pages';
import Room from './pages/[room]';
import HowTo from './pages/how-to';

const Router = () => {
  return (
    <BrowserRouter>
      <Switch>
        <Route path="/how-to" ><HowTo/></Route>
        <Route path="/:username" ><Room/></Route>
        <Route path="/" ><Home/></Route>
      </Switch>
    </BrowserRouter>
  )
}

export default Router;
