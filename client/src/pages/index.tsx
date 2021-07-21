import React, { useState, useEffect } from "react";
import { Link } from "react-router-dom"
import urljoin from "url-join";
import axios from "axios";
import dayjs from "dayjs";
import relativeTime from "dayjs/plugin/relativeTime";
import { findBestMatch } from "string-similarity";

import StreamPreview from "../components/StreamPreview";
import Loading from "../components/Loading";
import Navbar from "../components/Navbar";
import * as util from "../lib/util";

import PersonIcon from '@material-ui/icons/Person';
import TextField from '@material-ui/core/TextField';
dayjs.extend(relativeTime)

// max number of streampreview to display
const NDisplayLiveStreams = 6;

interface Room {
  StreamerID: string;
  LastActiveTime: string;
  StartedTime:string;
  StoppedTime:string;
  NViewers: number;
  AccNViewers:number;
  Title: string;
}

interface RatingExtended extends stringSimilarity.Rating {
  type?: string;
}

function Home() {

  const [ liveStreams, setLiveStreams ] = useState<Room[]>();
  const [ displayLiveStreams, setDisplayLiveStreams ] = useState<Room[]>();
  const [ recentStreams, setRecentStreams ] = useState<Room[]>();

  useEffect(() => {

    axios.get<Room[]>(urljoin(process.env.REACT_APP_API_URL as string, "/api/rooms?status=Streaming")).then((res) => {
      // get all for search purpose
      setLiveStreams(res.data);

      // display part of it because each streampreview will create an xterm to display preview
      // and xterm require a lot of memmory
      setDisplayLiveStreams(getDisplayStreams(res.data));
    }).catch((e) => console.error("Failed to get streaming rooms: ", e))

    axios.get<Room[]>(urljoin(process.env.REACT_APP_API_URL as string, "/api/rooms?status=Stopped&n=30")).then((res) => {
      let displayRecentStreams = res.data;

      // filter stream with duration more than 5 minutes
      displayRecentStreams = displayRecentStreams.filter((stream) => dayjs(stream.LastActiveTime).diff(dayjs(stream.StartedTime), "minute") > 5);
      // sort by descending started time
      displayRecentStreams = displayRecentStreams.sort((a, b) => dayjs(b.StartedTime).diff(dayjs(a.StartedTime)));
      setRecentStreams(displayRecentStreams);
    }).catch((e) => console.error("Failed to get streaming rooms: ", e))

  }, []);


  const getDisplayStreams = (streams: Room[]): Room[] => {
    if (!streams) return [];
    let newDisplayLiveStreams = [...streams];
    newDisplayLiveStreams.sort(() => 0.5 - Math.random());
    newDisplayLiveStreams = newDisplayLiveStreams.slice(0, NDisplayLiveStreams);
    return newDisplayLiveStreams
  }

  const handleSearch = (event: React.ChangeEvent<HTMLInputElement>) => {
    // TODO: schedule the serach, don't search on every type

    if (!liveStreams) return;
    const value = event.target.value;

    if (!value || value.length < 1) {
      setDisplayLiveStreams(getDisplayStreams(liveStreams));
      return
    }

    const findUniqueStreamerId = (value: string, type: string, matchedLiveStreams: Room[]): Room[] => {
      if(!liveStreams) return [];

      let fieldType = type as keyof Room;
      const result = liveStreams.filter((stream) => {
        return !matchedLiveStreams.includes(stream) && stream[fieldType] == value;
      });
      return result;
    }

    const streamerIDs = liveStreams.map((s) => s.StreamerID);
    const titles = liveStreams.map((s) => s.Title);

    let streamerIDRatings = findBestMatch(value, streamerIDs).ratings as RatingExtended[];
    streamerIDRatings.forEach((e) => e.type= "StreamerID");

    let titleRatings = findBestMatch(value, titles).ratings as RatingExtended[];
    titleRatings.forEach((e) => e.type= "Title");

    var mergeRatings = streamerIDRatings.concat(titleRatings) as RatingExtended[];;
    mergeRatings.sort((a, b) => b.rating - a.rating); // sort descending

    let matchedLiveStreams: Room[] = [];

    for (let e of mergeRatings){
      if (e.rating < .5) break;
      matchedLiveStreams = matchedLiveStreams.concat(findUniqueStreamerId(e.target, e.type as string, matchedLiveStreams));
    }
    setDisplayLiveStreams(matchedLiveStreams);

  }


  if (!displayLiveStreams || !recentStreams) return <Loading/>;

  let displayRecentStreams: Room[] = recentStreams;

  return (
    <>
      <Navbar />
      <div id="home" className="container m-auto text-white px-2">
        <div id="body" className="mt-8">
          <div id="intro">
            <p className="text-2xl mb-8 text-center font-bold">TStream - Live Stream from your terminal</p>
            <img className="border-2 border-gray-200 rounded-xl m-auto mt-4 w-4/5 xl:w-3/5" src="./demo.gif" />
          </div>

          <div id="previews">
            {!liveStreams && <p className="text-2xl mt-8 text-center font-bold">No one is live streaming 😅</p>}
            {liveStreams && 
              <>
                <div className="flex justify-center flex-wrap mt-8">
                  <p className="text-2xl w-full text-center font-bold">Live streaming</p>
                  <TextField className="m-auto mt-4 w-full mx-4 sm:w-96" label="Search" variant="standard" onChange={handleSearch}/>
                  <div id="listings" className="flex w-full justify-around my-5 flex-wrap">
                    {displayLiveStreams.map((r, i) =>
                    <Link key={i} to={`/${r.StreamerID}`}>
                      <StreamPreview
                        key={i} title={r.Title} streamerID={r.StreamerID}
                        startedTime={r.StartedTime} lastActiveTime={r.LastActiveTime}
                        wsUrl={util.getWsUrl(r.StreamerID)}
                        nViewers={r.NViewers}
                      />
                    </Link>
                    )}
                  </div>
                </div>
              </>}

            <p className="text-2xl mt-8 text-center font-bold">Recent Broadcasts</p>
            <div id="listings" className="flex w-full justify-around mt-5 flex-wrap mb-10">
              {displayLiveStreams && displayRecentStreams.map((r, i) =>
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

          </div>
        </div>
      </div>
    </>
  )
}

export default Home;
