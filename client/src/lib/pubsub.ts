export default class PubSub {

  // store callbaccks
  _handlers: {[key: string]: (data: any) => void} = {};

  // strict will allow only one consumer to consume one topic name
  _strict: boolean = false;

  constructor(strict: boolean = false) {
    this._strict = strict;
  }

  sub(topic: string, cb: (data: any) => void) {
    if (!this._handlers[topic]) this._handlers[topic] = cb;
    else if (!this._strict) this.sub(`topic_${Math.floor(Math.random() * 1000)}`, cb);
    else throw `${topic} are already subscribed!`;
  }

  pub(topic: string, msg: any) {
    if(this._strict) {
      if (topic in this._handlers) this._handlers[topic](msg);
    } else {
      for (const k in this._handlers) if (k.startsWith(k)) this._handlers[k](msg);
    }
  }
}
