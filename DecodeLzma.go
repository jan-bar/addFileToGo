package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jan-bar/addFileToGo/LzmaSpec"
)

func main() {
	in := flag.String("i", "", "input lzma file")
	out := flag.String("o", "", "output file")
	flag.Parse()
	if *in == "" || *out == "" {
		flag.Usage()
		return
	}
	//err := decoder(*in, *out)
	//if err != nil {
	//	fmt.Println(err)
	//}

	err := decoder2(*in, *out)
	if err != nil {
		fmt.Println(err)
	}

}

func decoder2(input, output string) error {
	fr, err := os.Open(input)
	if err != nil {
		return err
	}
	fw, err := os.Create(output)
	if err != nil {
		return err
	}

	LzmaSpec.LoadLzmaDll("LzmaSpec/LZMA.dll")
	err = LzmaSpec.LzmaCompress(fr, fw)
	if err != nil {
		return err
	}
	fr.Close()
	fw.Close()

	fr, err = os.Open(output)
	if err != nil {
		return err
	}
	fw, err = os.Create(input + ".txt")
	if err != nil {
		return err
	}
	err = LzmaSpec.LzmaUnCompress(fr, fw)
	if err != nil {
		return err
	}
	fr.Close()
	fw.Close()
	return nil
}

func decoder(input, output string) error {
	if input != "" {
		return nil
	}
	fr, err := os.Open(input)
	if err != nil {
		return err
	}
	defer fr.Close()
	fw, err := os.Create(output)
	if err != nil {
		return err
	}
	defer fw.Close()

	l := LzmaSpec.NewCLzmaDecoder(fr, fw)
	val, unpackSize, unpackSizeDefined, err := l.DecodeProperties()
	if err != nil {
		return err
	}
	fmt.Printf("\nlc=%d, lp=%d, pb=%d", val[0], val[1], val[2])
	fmt.Printf("\nDictionary Size in properties = %d", val[3])
	fmt.Printf("\nDictionary Size for decoding  = %d\n", val[4])

	done := make(chan struct{})
	if unpackSizeDefined {
		fmt.Println("Uncompressed Size: ", unpackSize)

		go func() {
			for { // 这样打印不费性能
				t := l.GetOutStreamProcessed() * 100 / unpackSize
				fmt.Printf("        %2d%%\r", t)
				if t == 100 {
					done <- struct{}{}
					break
				}
				time.Sleep(time.Second)
			}
		}()
	} else {
		fmt.Println("End marker is expected")
	}

	res, err := l.Decode(unpackSizeDefined, unpackSize)
	if unpackSizeDefined {
		select {
		case <-time.After(time.Second): // 防止异常
		case <-done: // 等一下打印到100%
		}
	}
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
