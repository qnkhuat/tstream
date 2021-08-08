//import * as pako from "pako";
//import * as constants from "./constants";
//import * as base64 from "./base64";
//import PubSub from "./pubsub";

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



// *** Message handlers ***
//export function hanldeWriteBlockMessage(msgData: string, msgManager: PubSub) {
//  let blockMsg: TermWriteBlock = JSON.parse(window.atob(msgData));
//  // this is a big chunk of encoding/decoding
//  // Since we have to : reduce message size by usign gzip and also
//  // every single termwrite have to be decoded, or else the rendering will screw up
//  // the whole block often took 9-20 milliseconds to decode a 3 seconds block of message
//  let data = pako.ungzip(base64.str2ab(blockMsg.Data));
//  let dataArr: string[] = JSON.parse(base64.ab2str(data));
//  //let bufferArray: Uint8Array[] = [];
//  dataArr.forEach((data: string) => {
//    let writeMsg: TermWrite = JSON.parse(window.atob(data));
//    let buffer = base64.str2ab(writeMsg.Data)
//    //bufferArray.push(buffer);
//    setTimeout(() => {
//      msgManager.pub(, buffer);
//    }, writeMsg.Offset as number);
//  })
//  //console.log(base64.concatab(bufferArray).length);
//  //msgManager.pub(msg.Type, base64.concatab(bufferArray));
//}
