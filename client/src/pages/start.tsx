import Navbar from "../components/Navbar";
import Code from "../components/Code";

function Start() {
  document.title = "TStream - How to start streaming";
  return (
    <>
      <Navbar />
      <div className="container m-auto justify-center flex px-2">

        <article className="prose lg:prose-xl mt-4 text-white overflow-hidden">
          <h1 className="text-white">How to start streaming</h1>
          <ol>
            <li>Download the <span className="font-bold">tstream</span> package from our <a href="https://github.com/qnkhuat/tstream/releases" className="text-green-term">Release</a> page. Make sure you download the version that matches your Operating System.</li>
            <li>Unpack it <ul>
              <li><Code className="text-white" text="tar -xzf tstream_{version}_{os}_{arch}.tar.gz"></Code></li></ul>
            </li>
            <li>(Optional) Setup TStream to run it anywhere
              <ul>
                <li><Code className="text-white" text="cp tstream /usr/local/bin"></Code></li>
              </ul>
            </li>
            <li>Start <span className="font-bold">tstream</span> 
              <ul>
                <li><Code className="text-white" text="tstream"></Code></li>
                <li>If you skipped step 3: <Code className="text-white" text="./tstream"></Code></li>
              </ul>
            </li>
          </ol>

          <h3 className="text-white">Chat from terminal</h3>
          <p>Want to chat with your viewers but not willing to leave terminal?</p>
          <p>We've got you covered!</p> 
          <p>Just type: <Code className="text-white" text="tstream -chat"></Code> after you started your stream session</p>
          
          <img alt="chat-demo" className="w-4/5 m-auto"src="./chat.gif"/>

          <h3 className="text-gray-400">Voice chat - Coming soon</h3>

          <h2 className="text-white">Tips</h2>
          <p>The current version of <span className="font-bold">tstream</span> can only work on one terminal tap</p>
          <p>In order for Streamers to stream with multiple taps, we suggest using a terminal multiplexer like <a className="text-green-term" href="https://github.com/tmux/tmux/wiki/Installing">tmux</a> or <a className="text-green-term" href="https://www.byobu.org">byobu</a></p>
          <p>Just make sure you start tstream <span className="font-bold">before</span> you start your terminal multiplexer.</p>
          <p>If you're a new tmux user, <a className="text-green-term" href ="https://thoughtbot.com/blog/a-tmux-crash-course">this</a> is a simple tutorial that provides all you need to know to get started with tmux.</p>

          <h3 className="text-red-400">Happy streaming!</h3>
          <br></br>
          <br></br>

        </article>

      </div>
    </>
  )
}

export default Start;
