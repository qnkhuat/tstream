const sep = "_" // sepperator used for seperate topic prefix and aliased name

export default class PubSub {

  // set to true if one topic can only have one subscriber
  _one = false;

  constructor(one:boolean = false){
    this._one = one;
  }

  // store callbaccks
  _handlers: {[key: string]: ((data: any) => void)[]} = {};

  // return the index of callback in list
  // used to delete if needed
  sub(topic: string, cb: (data: any) => void): number {
    if (!this._handlers[topic]) {
      this._handlers[topic] = [cb];
      return 0;
    } else if (!this._one){
      this._handlers[topic].push(cb);
      return this._handlers[topic].length - 1
    } else {
      throw "Only one subscribeer allowed";
    }
  }

  // Unsubscribe a topic
  // provide index if want to unsub a specific subscribe
  // otherwise will unsubscribe all
  unsub(topic: string, index: (number | null) = null) {
    if (index == null) delete this._handlers[topic]
    else delete this._handlers[topic][index]
  }

  pub(topic: string, msg: any) {
    if (this._handlers[topic]) this._handlers[topic].forEach((handler) => handler(msg))
  }
}
