package room

import (
	"bufio"
	"compress/gzip"
	"os"
)

type F struct {
	f  *os.File
	gf *gzip.Writer
	bf *bufio.Writer
}

func CreateWriteCloseGZ(path string, data []byte) error {
	f, err := CreateGZ(path)
	defer CloseGZ(f)
	if err != nil {
		return err
	}

	if err = WriteGZ(f, data); err != nil {
		return err
	}

	return nil
}

func CreateGZ(path string) (F, error) {
	var f F
	fi, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0660)
	if err != nil {
		return f, err
	}
	gf := gzip.NewWriter(fi)
	bf := bufio.NewWriter(gf)
	f = F{fi, gf, bf}
	return f, nil
}

func WriteGZ(f F, data []byte) error {
	n, err := f.bf.Write(data)
	if n != len(data) || err != nil {
		return err
	}
	return nil
}

func CloseGZ(f F) {
	f.bf.Flush()
	// Close the gzip first.
	f.gf.Close()
	f.f.Close()
}
