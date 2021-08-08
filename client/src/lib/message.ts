export interface TermWriteBlock {
  Data: string;
  Duration: number;
  StartTime: string;
}

export interface TermWrite {
  Data: string;
  Offset: number;
}

export interface ChatMsg {
  Name: string;
  Content: string;
  Color: string;
  Time: string;
}
