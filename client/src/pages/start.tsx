import Navbar from "../components/Navbar";
import Code from "../components/Code";

function Start() {
  document.title = "TStream - How to start streaming";
  return (
    <>
      <Navbar />
      <div className="container m-auto justify-center flex">

        <article className="prose lg:prose-xl mt-4 text-white">
          <h1 className="text-white">How to start streaming</h1>
          <ol>
            <li>Download the <span className="font-bold">tstream</span> package from our <a href="https://github.com/qnkhuat/tstream/releases" className="text-green-term">Release</a> page. Make sure you download the version that matches your Operating System.</li>
            <li>Unpack <Code className="text-white" text="tar -xzf tstream_{version}_{os}_{arch}.tar.gz"></Code></li>
            <li>(Optional) Setup TStream so you can run it anywhere <Code className="text-white" text="cp tstream_{version}_{os}_{arch}/tstream /usr/local/bin"></Code></li>
            <li>Start <span className="font-bold">tstream</span> <Code className="text-white" text="tstream"></Code> or <Code className="text-white" text="./tstream_{version}_{os}_{arch}/tstream/tstream"></Code> if you skipped step 3</li>
          </ol>

          <h1 className="text-white">Tips</h1>
          <p>The current version of <span className="font-bold">tstream</span> can only work on one terminal tap.</p>
          <p>So in order for Streamer to stream multiple taps, we suggest using a terminal multiplexer like <a className="text-green-term" href="https://github.com/tmux/tmux/wiki/Installing">tmux</a> or <a className="text-green-term" href="https://www.byobu.org">byobu</a></p>
          <p>Just make sure you start <span className="font-bold">tstream</span> before you start your terminal multiplexer.</p>
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
