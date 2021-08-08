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
