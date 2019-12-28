package main

import (
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

func main() {
	type fileInfo struct {
		file string
		size int64
	}

	fileList := make(map[string]fileInfo, len(os.Args))
	for _, v := range os.Args[1:] {
		if f, err := os.Stat(v); err == nil || os.IsExist(err) {
			fileList[filepath.Base(v)] = fileInfo{file: v, size: f.Size()}
		}
	}
	if len(fileList) == 0 {
		log.Fatal(os.Args, "input error")
	}

	fw, err := os.Create("resources.go")
	if err != nil {
		log.Fatal(err)
	}
	defer fw.Close()

	fw.WriteString(`package main

import (
	"compress/zlib"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
)

func WriteBytesToFile(path, name string) error {
	ok, err := checkFileIsOk(path, name)
	if err != nil || ok {
		return err
	}
	zr, err := zlib.NewReader(base64.NewDecoder(base64.RawStdEncoding, getBytesFromMap(name)))
	if err != nil {
		return err
	}
	defer zr.Close()
	fw, err := os.Create(path)
	if err != nil {
		return err
	}
	defer fw.Close()
	_, err = io.Copy(fw, zr)
	return err
}

func checkFileIsOk(path, name string) (bool, error) {
	size, ok := map[string]int64{
`)

	for k, v := range fileList {
		fmt.Fprintf(fw, "		\"%s\":%d,\n", k, v.size)
	}

	fw.WriteString(`	}[name]
	if !ok {
		return false, errors.New("name not find")
	}
	f, err := os.Stat(path)
	return (err == nil || os.IsExist(err)) && size == f.Size(), nil
}

func getBytesFromMap(name string) io.Reader {
	data := map[string]string{
`)

	for k, v := range fileList {
		fmt.Fprintf(fw, "		\"%s\":\"", k)
		if err = writeDataToFile(fw, v.file); err != nil {
			log.Fatal(err)
		}
		fw.WriteString("\",\n")
	}

	fw.WriteString(`	}[name]
	return strings.NewReader(data)
}`)
}

func writeDataToFile(fw io.Writer, file string) error {
	fr, err := os.Open(file)
	if err != nil {
		return err
	}
	defer fr.Close()

	bw := base64.NewEncoder(base64.RawStdEncoding, fw)
	defer bw.Close()
	zw := zlib.NewWriter(bw)
	defer zw.Close()
	_, err = io.Copy(zw, fr)
	return err
}
