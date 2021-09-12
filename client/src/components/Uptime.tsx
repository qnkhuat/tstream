import React, { useState, useEffect } from "react";
import * as utils from "./../utils";
import dayjs from "dayjs";

interface Props {
  startTime: Date;
  className?: string;
}

const Uptime: React.FC<Props> = ({startTime, className=""}) => {
  const [ upTime, setUptime ] = useState(utils.formatDuration(dayjs().diff(dayjs(startTime), "second"), true) );
  useEffect(() => {
    let id = setInterval(() => {
      setUptime(utils.formatDuration(dayjs().diff(dayjs(startTime), "second"), true));
    }, 1000)
    return () => {
      clearInterval(id);
    }

  }, []);

  return  <p className={className} >{upTime}</p>;
}

export default Uptime;
