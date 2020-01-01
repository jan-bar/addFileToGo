package main

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/jan-bar/addFileToGo/LzmaSpec"
)

func main() {
	err := main2()
	if err != nil {
		fmt.Println(err)
	}
}

func main2() error {
	fr, err := os.Open(`E:\code_porject\java\7z\lzma-specification\examples\xx.lzma`)
	if err != nil {
		return err
	}
	defer fr.Close()
	fw, err := os.Create(`E:\code_porject\java\7z\lzma-specification\examples\a.txt`)
	if err != nil {
		return err
	}
	defer fw.Close()

	l := LzmaSpec.NewCLzmaDecoder(fr, fw)
	defer l.Close()

	val, unpackSize, unpackSizeDefined, err := l.DecodeProperties()
	if err != nil {
		return err
	}
	fmt.Printf("\nlc=%d, lp=%d, pb=%d", val[0], val[1], val[2])
	fmt.Printf("\nDictionary Size in properties = %d", val[3])
	fmt.Printf("\nDictionary Size for decoding  = %d\n", val[4])
	if unpackSizeDefined {
		fmt.Println("Uncompressed Size: ", unpackSize)
	} else {
		fmt.Println("End marker is expected")
	}

	processing := uint64(0)

	go func() {
		for { // 这样打印不费性能
			t := atomic.LoadUint64(&processing)
			fmt.Printf("    %2d%%\r", t)
			if t == 100 {
				break
			}
			time.Sleep(time.Second)
		}
	}()

	res, err := l.Decode(unpackSizeDefined, func(unpackSize, processed uint64) error {
		atomic.StoreUint64(&processing, 100*processed/unpackSize)
		return nil
	})
	if err != nil {
		return err
	}
	fmt.Println()

	switch res {
	case LzmaSpec.LzmaResError:
		return errors.New("LZMA decoding error")
	case LzmaSpec.LzmaResFinishedWithoutMarker:
		fmt.Println("Finished without end marker")
	case LzmaSpec.LzmaResFinishedWithMarker:
		if unpackSizeDefined {
			if l.GetOutStreamProcessed() != unpackSize {
				return errors.New("finished with end marker before than specified size")
			}
			fmt.Print("Warning: ")
		}
		fmt.Println("Finished with end marker")
	default:
		return errors.New("internal Error")
	}

	if l.IsCorrupted() {
		fmt.Println("\nWarning: LZMA stream is corrupted")
	}
	return nil
}
