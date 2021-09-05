import { useState, useCallback} from "react";
import { Dimension } from "../types";

// Get dimension of a DOM when it's mounted
const useDimension = (): [Dimension, (args: any) => void ] => {
  const [ dimension, setDimension ] = useState<Dimension>({width: 0, height: 0});

  const refCallback = useCallback(node => {
    if (node !== null) {
      setDimension({height: node.offsetHeight, width: node.offsetWidth});
    }
  }, []);
  
  return [ dimension, refCallback ];
}

export default useDimension;
