import {
  BrowserRouter,
  Switch,
  Route
} from "react-router-dom";

import Home from './pages';
import Room from './pages/[room]';
import Start from './pages/start';

const Router = () => {
  return (
    <BrowserRouter>
      <Switch>
        <Route path="/start-streaming" ><Start/></Route>
        <Route path="/:roomID" ><Room/></Route>
        <Route path="/" ><Home/></Route>
      </Switch>
    </BrowserRouter>
  )
}

export default Router;
