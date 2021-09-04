import {
  BrowserRouter,
  Switch,
  Route,
} from "react-router-dom";

import Home from './pages';
import Room from './pages/[room]';
import Records from './pages/records/[id]';
import Start from './pages/start';

const Router = () => {
  return (
    <BrowserRouter>
      <Switch>
        <Route path="/start-streaming" ><Start/></Route>
        <Route path="/records/:Id" render={(props) => <Records {...props}/>}></Route>
        <Route path="/:roomID" ><Room/></Route>
        <Route path="/" ><Home/></Route>
      </Switch>
    </BrowserRouter>
  )
}

export default Router;
