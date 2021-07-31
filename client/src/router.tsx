import {
  BrowserRouter,
  Switch,
  Route
} from "react-router-dom";

import Home from './pages';
import Room from './pages/[room]';
import Start from './pages/start';
import Tracker from "./components/Tracker";

const Router = () => {
  return (
    <BrowserRouter>
      <Tracker trackingId={process.env.REACT_APP_GA}>
        <Switch>
          <Route path="/start-streaming" ><Start/></Route>
          <Route path="/:roomID" ><Room/></Route>
          <Route path="/" ><Home/></Route>
        </Switch>
      </Tracker>
    </BrowserRouter>
  )
}

export default Router;
