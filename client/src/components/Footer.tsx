import React, { FC } from "react";

interface Props {
}

const Footer: FC<Props> = (props) =>  {
  return (
    <>
      
      <div id="footer"
        className="mt-8 py-4 text-center text-gray-500 text-sm border-t border-gray-800">
        <a href="https://www.producthunt.com/posts/tstream?utm_source=badge-featured&utm_medium=badge&utm_souce=badge-tstream" target="_blank">
          <img className="mx-auto mb-4" src="https://api.producthunt.com/widgets/embed-image/v1/featured.svg?post_id=305917&theme=light" alt="TStream - Live streaming from your terminal | Product Hunt" width="250" height="54" /></a>
        <a href={`https://github.com/qnkhuat/tstream`}><span className="underline">Github</span></a>
        <span> | </span>
        <a href={`https://discord.gg/qATHjk6ady`}><span className="underline">Discord</span></a>
        <br></br>
        <p id="copy-right" className="text-gray-500 text-xs">© {new Date().getFullYear()} TStream</p>
      </div> 
    </>
  )
}

export default Footer;
