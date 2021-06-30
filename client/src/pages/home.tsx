import React, { useState, useEffect, useRef, RefObject } from "react";
import Xterm from "../components/Xterm";
import colors from "ansi-colors";

function Home() {
  const term = React.useRef() as RefObject<Xterm>;
  useEffect(() => {
    if (term) {
      term.current?.writeln("sup bitchse");
      term.current?.writeln(colors.red("YO"));
      term.current?.writeln(colors.blue("YO"));
      term.current?.prompt();
    }
  }, [term])
  return (
    <>
      <Xterm
        ref={term}
        onKey={(e) => {
          const ev = e.domEvent;
          const printable = !ev.altKey && !ev.ctrlKey && !ev.metaKey;

          if (ev.keyCode === 13) {
            term.current?.prompt();
          } else if (ev.keyCode === 8) {
            //if (term.current?.terminal.buffer.x > 2) {
            term.current?.write('\b \b');
            console.log(term.current?.terminal);
            //}
          } else if (printable) {
            term.current?.write(e.key);
          }

        }}
        onData={(data) => {
          console.log("Get data: ", data);
        }}

        onLineFeed={() => {
          console.log("Line feed");
        }}

      />
      <h3>Home</h3>
    </>
  )
}

export default Home;
