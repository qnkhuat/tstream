import { Link } from "react-router-dom"
import React, { useState, useEffect } from "react";
import StreamPreview from "../components/StreamPreview";
import Navbar from "../components/Navbar";
import * as util from "../lib/util";
import urljoin from "url-join";
import axios from "axios";

//import SearchIcon from '@material-ui/icons/Search';
//import FormControl from '@material-ui/core/FormControl';
//import InputLabel from '@material-ui/core/InputLabel';
//import TextField from '@material-ui/core/TextField';
//import Input from '@material-ui/core/Input';
//import InputAdornment from '@material-ui/core/InputAdornment';
//import Button from '@material-ui/core/Button';


interface Room {
  streamerID: string;
  lastActiveTime: string;
  startedTime:string;
  nViewers: number;
  title: string;
}

function Home() {

  const [ rooms, setRooms ] = useState<Room[]>();
  console.log(process.env.REACT_APP_API_URL);
  useEffect(() => {
    axios.get<Room[]>(urljoin(process.env.REACT_APP_API_URL as string, "/api/rooms")).then((res) => {
      setRooms(res.data);
    })
  }, []);

  return (
    <>
      <Navbar />
      <div id="home" className="container m-auto text-white">
        <div id="body">
          <div id="intro">
          </div>

          <div id="previews"
            className="flex-row items-center justify-center">
            <div className="flex justify-center">
              {/*
                <FormControl variant="standard"
                className="w-96"
                color="success">
                <Input
                  id="standard-adornment-amount"
                  placeholder="Search"
                  startAdornment={<InputAdornment position="start"><span className="font-bold pb-1 text-green-term">{'>'}</span></InputAdornment>}
                />
              </FormControl>
                */}
            </div>
            <div id="listings" className="flex w-full justify-around m-5 flex-wrap">
              {rooms?.map((r, i) =>
              <Link to={`/${r.streamerID}`}>
                <StreamPreview
                  key={i} title={r.title} streamerID={r.streamerID}
                  startedTime={r.startedTime} lastActiveTime={r.lastActiveTime}
                  wsUrl={util.getWsUrl(r.streamerID)}
                  nViewers={r.nViewers}
                />
              </Link>
              )}
            </div>
          </div>
        </div>
      </div>
    </>
  )
}

export default Home;