export default function Loading() {
  return (
    <div className="fixed w-screen h-screen top-0 left-0 bg-white flex items-center place-content-center z-10">
      <img src="./tstream-black.svg"  alt="logo"
        className="w-44 animate-pulse" />
    </div>
  )
}
