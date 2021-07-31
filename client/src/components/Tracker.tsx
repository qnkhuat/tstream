import {Location, LocationListener, UnregisterCallback} from 'history'
import {useEffect} from 'react'
import ReactGA from 'react-ga'
import {useHistory} from 'react-router'

const sendPageView: LocationListener = (location: Location): void => {
  ReactGA.set({ page: location.pathname });
  ReactGA.pageview(location.pathname);
}

interface Props {
  children: JSX.Element;
  trackingId?: string;
}

const Tracker = ({ children, trackingId }: Props): JSX.Element => {
  const history = useHistory();
  useEffect((): UnregisterCallback | void => {
    if (trackingId && trackingId.length > 0) {
      ReactGA.initialize(trackingId);
      sendPageView(history.location, 'REPLACE');
      return history.listen(sendPageView);
    }
  }, [history, trackingId])

  return children;
}

export default Tracker;
