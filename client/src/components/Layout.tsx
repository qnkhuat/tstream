import React, { FC, ReactElement } from "react";

interface Props {
  children: React.ReactNode;
  ref: React.RefObject<HTMLDivElement>;
}

const Layout: FC<Props> = ({ children, ref }) =>  {
  return (
    <>
      <div id="navbar"
        className="fixed top-0 left-0 w-screen">
        <h3>This is a layout </h3>
        <h3>This is a layout </h3>
        <h3>This is a layout </h3>
        <h3>This is a layout </h3>
        <h3>This is a layout </h3>
        <h3>This is a layout </h3>
        <h3>This is a layout </h3>
        <h3>This is a layout </h3>
        <h3>This is a layout </h3>
        <h3>This is a layout </h3>
        <h3>This is a layout </h3>
      </div>
      <div className="body pt-10">
        {children}
      </div>
    </>
  )
}

export default Layout;
