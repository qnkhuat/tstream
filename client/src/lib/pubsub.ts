const sep = "_"
export default class PubSub {

  // store callbaccks
  _handlers: {[key: string]: (data: any) => void} = {};

  // set to true if want allow multiple consumers for one topic
  _strict: boolean = false;

  constructor(strict: boolean = false) {
    this._strict = strict;
  }

  sub(topic: string, cb: (data: any) => void): string | null {
    if (!this._handlers[topic]) {
      this._handlers[topic] = cb;
      return topic
    } else if (!this._strict) {
      const topic_name = `${topic}${sep}${Math.floor(Math.random() * 1000)}`;
      return this.sub(topic_name, cb);
    } else { // topic already subscribed and strict is not set
      throw `Topic ${topic} already subscribed`;
    }
  }

  // Unsubscribe a topic
  // set all to true to unsubscribe all all topic. Only used when _strict is true
  unsub(topic: string, all: boolean = false) {
    if (this._handlers[topic]) delete this._handlers[topic] ;

    // search for all hanlders with topic as prefix and delete it
    if (!this._strict && all) for (const k in this._handlers) if ( k.startsWith(`${topic}${sep}`) ) delete this._handlers[k];
  }

  pub(topic: string, msg: any) {
    if(this._strict) {
      if (topic in this._handlers) this._handlers[topic](msg)
    }
    else {
      for (const k in this._handlers) if (k.startsWith(topic)) this._handlers[k](msg);
    }
  }
}
