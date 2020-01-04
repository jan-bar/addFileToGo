package LzmaSpec

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
)

// from 7zTypes.h
const (
	SzOk               = 0
	SzErrorData        = 1
	SzErrorMem         = 2
	SzErrorCrc         = 3
	SzErrorUnsupported = 4
	SzErrorParam       = 5
	SzErrorInputEOF    = 6
	SzErrorOutputEOF   = 7
	SzErrorRead        = 8
	SzErrorWrite       = 9
	SzErrorProgress    = 10
	SzErrorFail        = 11
	SzErrorThread      = 12
	SzErrorArchive     = 16
	SzErrorNoArchive   = 17
)

const (
	LzmaPropsSize = uint64(5)

	sizeHeader = 8
)

func getLenBytes(byt []byte) uint64 {
	size := uint64(0)
	for i := 0; i < sizeHeader; i++ {
		size |= uint64(byt[i]) << (8 * i)
	}
	return size
}

func setLenBytes(size uint64) []byte {
	byt := make([]byte, sizeHeader)
	for i := 0; i < sizeHeader; i++ {
		byt[i] = byte(size >> (8 * i))
	}
	return byt
}

type UseType byte

const (
	UseDll = 1
	UseCgo = 2
)

func LzmaCompress(r io.Reader, w io.Writer, useType UseType) error {
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	var (
		srcLen    = uint64(len(src))
		propsSize = LzmaPropsSize
		dstLen    = srcLen + srcLen/3 + 128
		dst       = make([]byte, sizeHeader+LzmaPropsSize+dstLen)
		ret       = -1
	)
	copy(dst, setLenBytes(srcLen)) // 保存源文件长度

	switch useType { // outProps也需要保存
	case UseDll:
		ret, err = lzmaCompressDll(dst[sizeHeader+LzmaPropsSize:], &dstLen, src, srcLen,
			dst[sizeHeader:sizeHeader+LzmaPropsSize], &propsSize, 9, 1<<24, 3, 0, 2, 32, 2)
		if err != nil {
			return err
		}
	case UseCgo:
		ret = lzmaCompressCgo(dst[sizeHeader+LzmaPropsSize:], &dstLen, src, srcLen,
			dst[sizeHeader:sizeHeader+LzmaPropsSize], &propsSize, 9, 1<<24, 3, 0, 2, 32, 2)
	default:
		return fmt.Errorf("useType error: %d", useType)
	}
	if ret != SzOk {
		return fmt.Errorf("lzmaCompress ret: %d", ret)
	}
	if propsSize != LzmaPropsSize {
		return errors.New("propsSize error")
	}
	_, err = w.Write(dst[:sizeHeader+LzmaPropsSize+dstLen])
	return err
}

func LzmaUnCompress(r io.Reader, w io.Writer, useType UseType) error {
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	var (
		srcLen = uint64(len(src)) - sizeHeader - LzmaPropsSize // 去掉头部
		dstLen = getLenBytes(src[:sizeHeader])                 // 读取源文件大小
		dst    = make([]byte, dstLen)                          // 申请资源
		ret    = -1
	)
	switch useType { // 使用r中读到的props传递参数
	case UseDll:
		ret, err = lzmaUncompressDll(dst, &dstLen, src[sizeHeader+LzmaPropsSize:],
			&srcLen, src[sizeHeader:sizeHeader+LzmaPropsSize], LzmaPropsSize)
		if err != nil {
			return err
		}
	case UseCgo:
		ret = lzmaUncompressCgo(dst, &dstLen, src[sizeHeader+LzmaPropsSize:],
			&srcLen, src[sizeHeader:sizeHeader+LzmaPropsSize], LzmaPropsSize)
	}
	if ret != SzOk {
		return fmt.Errorf("lzmaCompress ret: %d", ret)
	}
	_, err = w.Write(dst[:dstLen])
	return err
}
