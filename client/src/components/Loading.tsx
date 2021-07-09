import Navbar from "./Navbar";

export default function Loading() {
  return (
    <>
      <div className="fixed w-screen h-screen top-0 left-0 bg-black flex items-center place-content-center z-50">
        <div className="w-full fixed top-0 left-0">
          <Navbar />
        </div>
        <img src="./logo.svg"  alt="logo"
          className="w-44 animate-pulse" />
      </div>
    </>
  )
}
