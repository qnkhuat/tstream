import { Link } from "react-router-dom";
import React from "react";

interface Props {
}

const Navbar: React.FC<Props> = () => {
  return (
    <div id="navbar" className="flex justify-center py-2 border-b border-black shadow"
      style={{background:"#18181b"}} >
      <div className="container flex justify-between">
        <Link to="/">
          <img alt={"logo"} className="w-10 h-10" src="./logo.svg" />
        </Link>
        <div className="flex items-center font-bold text-gray-100">
          <Link to="/how-to" className="border-r border-white pr-4">Start streaming</Link>
          <a href="https://github.com/qnkhuat/tstream" className="pl-4">Github</a>
        </div>
      </div>
    </div>
  )
}

export default Navbar;
