import * as message from "../types/message";
import urljoin from "url-join";
import axios from "axios";

interface getRoomsArg {
  status: message.RoomStatus;
  n?: number;
}

export const getRooms = async (arg: getRoomsArg): Promise<message.RoomInfo[]> => {
  let url = urljoin(process.env.REACT_APP_API_URL as string, "/api/rooms");
  return axios.get<message.RoomInfo[]>(url, { params:arg }).then(( res ) => res.data);
}
