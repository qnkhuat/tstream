// str to buffer array
export function str2ab(input:string): Uint8Array{
  let binary_string =  window.atob(input);
  let len = binary_string.length;
  let bytes = new Uint8Array( len );
  for (let i = 0; i < len; i++)        {
    bytes[i] = binary_string.charCodeAt(i);
  }
  return bytes;
}

// array buffer to string
export function ab2str(buf: number[]): string{
  return String.fromCharCode.apply(null, buf);
}
