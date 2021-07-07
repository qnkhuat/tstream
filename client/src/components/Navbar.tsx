import { Link } from "react-router-dom";
import React from "react";

interface Props {
}

const Navbar: React.FC<Props> = () => {
  return (
    <Link to="/">
      <div id="navbar" className="flex justify-center py-2 border-b border-black"
        style={{background:"#18181b"}} >
        <div className="container" >
          <img alt={"logo"} className="w-9 h-9" src="./logo2.svg" />
        </div>
      </div>
    </Link>
  )
}

export default Navbar;
