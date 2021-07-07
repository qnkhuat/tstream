import React from "react";

interface Props {
  text: string,
    className?: string,
}
const Code: React.FC<Props> = ({text, className=""}) => {
  return <code className={`whitespace-pre rounded-md bg-gray-400 p-1 ${className}`}>{text}</code>;
}

export default Code;
