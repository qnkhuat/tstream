import { Link } from "react-router-dom";
import React from "react";

interface Props {
}

const Navbar = React.forwardRef<HTMLDivElement, Props>(({}, ref) => {
  return (
    <div ref={ref} id="navbar" className="flex justify-center py-2 border-b border-black shadow"
      style={{background:"#18181b"}} >
      <div className="container flex justify-between px-2">
        <Link to="/">
          <img alt={"logo"} className="w-10 h-10" src="/logo.svg" />
        </Link>
        <div className="flex items-center font-bold text-gray-100">
          <Link to="/start-streaming" className="border-r border-white pr-4">Start streaming</Link>
          <a href="https://github.com/qnkhuat/tstream" className="pl-4">GitHub</a>
        </div>
      </div>
    </div>
  )
})

export default Navbar;
