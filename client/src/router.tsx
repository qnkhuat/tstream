import {
  BrowserRouter,
  Switch,
  Route,
  Link
} from "react-router-dom";

import Home from './pages/home';
import Room from './pages/[room]';

const Router = () => {
  return (
    <BrowserRouter>
      <Switch>
        <Route path="/:username" ><Room/></Route>
        <Route path="/" ><Home/></Route>
      </Switch>
    </BrowserRouter>
  )
}

export default Router;
