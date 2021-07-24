import React, { useState, useEffect } from "react";
import { Link } from "react-router-dom"
import urljoin from "url-join";
import axios from "axios";
import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";

import StreamPreview from "../components/StreamPreview";
import Navbar from "../components/Navbar";
import * as util from "../lib/util";

import PersonIcon from '@material-ui/icons/Person';
dayjs.extend(relativeTime)


interface Room {
  StreamerID: string;
  LastActiveTime: string;
  StartedTime:string;
  StoppedTime:string;
  NViewers: number;
  AccNViewers:number;
  Title: string;
}

function Home() {

  const [ liveRooms, setLiveRooms ] = useState<Room[]>();
  const [ recentStreams, setRecentStreams ] = useState<Room[]>();
  useEffect(() => {

    axios.get<Room[]>(urljoin(process.env.REACT_APP_API_URL as string, "/api/rooms?status=Streaming&n=6")).then((res) => {
      setLiveRooms(res.data);
    }).catch((e) => console.error("Failed to get streaming rooms: ", e))

    axios.get<Room[]>(urljoin(process.env.REACT_APP_API_URL as string, "/api/rooms?status=Stopped&n=20")).then((res) => {
      setRecentStreams(res.data);
    }).catch((e) => console.error("Failed to get streaming rooms: ", e))

  }, []);

  return (
    <>
      <Navbar />
      <div id="home" className="container m-auto text-white px-2">
        <div id="body" className="mt-8">
          <div id="intro">

            <p className="text-2xl mb-8 text-center font-bold">TStream - Live Stream from your terminal</p>
            <img className="border-2 border-gray-200 rounded-xl m-auto mt-4 w-4/5 xl:w-3/5" src="./demo.gif" />

          </div>
          <div id="previews"
            className="flex-row items-center justify-center">
            {!liveRooms && <p className="text-2xl mt-8 text-center font-bold">No one is live streaming 😅</p>}
            {liveRooms && 
              <>
                <p className="text-2xl mt-8 text-center font-bold">Live streaming</p>
                <div id="listings" className="flex w-full justify-around my-5 flex-wrap">
                  {liveRooms.map((r, i) =>
                  <Link to={`/${r.StreamerID}`}>
                    <StreamPreview
                      key={i} title={r.Title} streamerID={r.StreamerID}
                      startedTime={r.StartedTime} lastActiveTime={r.LastActiveTime}
                      wsUrl={util.getWsUrl(r.StreamerID)}
                      nViewers={r.NViewers}
                    />
                  </Link>
                  )}
                </div>
              </>}

            {recentStreams && 
              <>
                <p className="text-2xl mt-8 text-center font-bold">Recent Broadcasts</p>
                <div id="listings" className="flex w-full justify-around mt-5 flex-wrap mb-10">
                  {recentStreams.sort((a, b) => dayjs(b.StartedTime).diff(dayjs(a.StartedTime))).map((r, i) =>
                  <div key={i} className="w-full sm:w-5/12 lg:w-3/12 bg-gray-600 p-4 rounded-lg flex justify-between m-4 flex-wrap relative">

                    <div className="left mr-20">
                      <p className="font-bold ">{r.Title}</p>
                      <p>@{r.StreamerID}</p>
                      <p>{dayjs(r.StartedTime).fromNow()}</p>
                    </div>

                    <p className="absolute top-4 right-4 bg-gray-800 p-1 rounded-md font-semibold">{util.formatDuration(dayjs(r.LastActiveTime), dayjs(r.StartedTime))}</p>
                    <p className="absolute bottom-4 right-4 text-whtie font-semibold text-right mt-4"><PersonIcon/> {r.AccNViewers}</p>

                  </div>)}
                </div>
              </>}

          </div>
        </div>
      </div>
    </>
  )
}

export default Home;
