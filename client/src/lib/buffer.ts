export function str2ab(input:string): Uint8Array{
  let binary_string =  window.atob(input);
  let len = binary_string.length;
  let bytes = new Uint8Array(len);
  for (let i = 0; i < len; i++)        {
    bytes[i] = binary_string.charCodeAt(i);
  }
  return bytes;
}

export function ab2str(buf: any): string{
  return new TextDecoder().decode(buf);
}

export function concatab(array: Uint8Array[]): Uint8Array {
  let len = 0;
  array.forEach((a) => { len += a.byteLength; });

  let result = new Uint8Array(len);
  let runningIndex = 0
  array.forEach((a) => {
    result.set(new Uint8Array(a), runningIndex);
    runningIndex += a.byteLength;
  });
  return result;

}
