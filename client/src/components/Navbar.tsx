import { Link } from "react-router-dom";
import React from "react";

interface Props {
}

const Navbar: React.FC<Props> = () => {
  return (
    <Link to="/">
      <div id="navbar" className="flex justify-center items-center py-1 bg-black border-b border-green-400">
        <img alt={"logo"} className="h-12 mr-2" src="./tstream-green.svg" />
        <p className="text-center text-2xl text-green-term font-bold">TStream</p>
      </div>
    </Link>
  )
}

export default Navbar;
