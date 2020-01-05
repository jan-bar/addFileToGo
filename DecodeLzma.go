package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jan-bar/addFileToGo/LzmaSpec"
)

func main() {
	// LzmaSpec\lzma.exe e test.txt LZMA.lzma
	// java -jar LzmaSpec\lzma.jar e test.txt LZMA.lzma
	err := decoder("LZMA.lzma", "test.txt")
	if err != nil { // 加压标准lzma文件
		fmt.Println("decoder", err)
		return
	}
	err = janbar("test.txt", "LZMA_my.lzma")
	if err != nil { // 使用库进行自定义压缩解压
		fmt.Println("janbar", err)
		return
	}
	err = lzmaUtil("test.txt", "LZMA_util.lzma")
	if err != nil { // 使用库进行自定义压缩解压
		fmt.Println("lzmaUtil", err)
		return
	}
}

func lzmaUtil(input, output string) error {
	err := LzmaSpec.LzmaCompressUtil(input, output, LzmaSpec.UtilEnc)
	if err != nil {
		return err
	}
	return LzmaSpec.LzmaCompressUtil(output, output+".d", LzmaSpec.UtilDec)
}

// 下面时使用7z提供的接口压缩和解压缩
// diff  LZMA_my.lzma.1 LZMA_my.lzma.2 // 这两个也是压缩文件可以用7z打开查看详情
// diff3 LZMA_my.lzma.1.d LZMA_my.lzma.2.d test.txt
func janbar(input, output string) error {
	LzmaSpec.LoadLzmaDll("LzmaSpec/LZMA.dll")
	err := encJanbar(input, output+".1", LzmaSpec.UseDll)
	if err != nil {
		return err
	}
	err = decJanbar(output+".1", output+".1.d", LzmaSpec.UseDll)
	if err != nil {
		return err
	}

	err = encJanbar(input, output+".2", LzmaSpec.UseCgo)
	if err != nil {
		return err
	}
	err = decJanbar(output+".2", output+".2.d", LzmaSpec.UseCgo)
	if err != nil {
		return err
	}
	return nil
}

func encJanbar(input, output string, useType LzmaSpec.UseType) error {
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
	return LzmaSpec.LzmaCompress(fr, fw, useType)
}

func decJanbar(input, output string, useType LzmaSpec.UseType) error {
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
	return LzmaSpec.LzmaUnCompress(fr, fw, useType)
}

// 使用lzma.exe,lzma.jar等标准压缩文件,下面时解压逻辑
func decoder(input, output string) error {
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
