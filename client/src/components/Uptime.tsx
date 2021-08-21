import React, { useState, useEffect } from "react";
import * as utils from "./../utils";
import dayjs from "dayjs";

interface Props {
  startTime: Date;
  className?: string;
}

const Uptime: React.FC<Props> = ({startTime, className=""}) => {
  const [ upTime, setUptime ] = useState(utils.formatDuration(dayjs(), dayjs(startTime)));
  useEffect(() => {
    let id = setInterval(() => {
      setUptime(utils.formatDuration(dayjs(), dayjs(startTime)));
    }, 1000)
    return () => {
      clearInterval(id);
    }

  }, []);

  return  <p className={className} >{upTime}</p>;
}

export default Uptime;
