import React from 'react';
import ReactDOM from 'react-dom';
import App from './app';
import ReactGA from 'react-ga';

// Google analytics
const GA_KEY = process.env.REACT_APP_GA;
if (GA_KEY && GA_KEY.length > 0) {
  ReactGA.initialize(GA_KEY);
  ReactGA.pageview(window.location.pathname + window.location.search);
}

ReactDOM.render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
  document.getElementById('root')
);
