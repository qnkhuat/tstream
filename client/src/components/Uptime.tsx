import React, { useState, useEffect } from "react";
import * as util from "./../lib/util";
import dayjs from "dayjs";

interface Props {
  startTime: Date;
  className?: string;
}

const Uptime: React.FC<Props> = ({startTime, className=""}) => {
  const [ upTime, setUptime ] = useState(util.getUpTime(dayjs(startTime)));
  useEffect(() => {
    setInterval(() => {
      setUptime(util.getUpTime(dayjs(startTime)));
    }, 1000)

  }, []);

  return  <p className={className} >{upTime}</p>;
}

export default Uptime;
