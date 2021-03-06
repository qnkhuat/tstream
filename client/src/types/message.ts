export interface Wrapper {
  Type: string;
  Data: any;
  Delay: number;
}

export interface TermWriteBlock {
  Data: string;
  Duration: number;
  StartTime: string;
}

export interface ChatMsg {
  Name: string;
  Content: string;
  Color: string;
  Time: string;
}

export enum RoomStatus {
  Streaming = "Streaming",
    Stopped = "Stopped",
}

export interface RoomInfo {
  StreamerID: string;
  LastActiveTime: string;
  StartedTime:string;
  StoppedTime:string;
  NViewers: number;
  AccNViewers:number;
  Title: string;
}

