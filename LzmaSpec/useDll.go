/**
[https://www.7-zip.org/a/lzma1900.7z]解压后lzma1900\C\Util\LzmaLib
使用vs2010打开LzmaLib.dsw,在配置管理器里面新建一个x64的release版本(产生dll比较小)
然后生成解决方案,会在C:\Util产生LZMA.dll文件

安装vs2010,可以用下面命令查看dll对外提供接口,只能查看函数名
dumpbin.exe -exports c:\Util\LZMA.dll
然后在这个\lzma1900\C\LzmaLib.h里面可以看到 LzmaCompress和 LzmaUncompress的定义
以及各种注意事项,下面用法就是从上面得到的
**/
package LzmaSpec

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"syscall"
	"unsafe"
)

var (
	compress   *syscall.Proc
	unCompress *syscall.Proc
)

func LoadLzmaDll(file string) {
	dll := syscall.MustLoadDLL(file)
	compress = dll.MustFindProc("LzmaCompress")
	unCompress = dll.MustFindProc("LzmaUncompress")
}

/*
RAM requirements for LZMA:
  for compression:   (dictSize * 11.5 + 6 MB) + state_size
  for decompression: dictSize + state_size
    state_size = (4 + (1.5 << (lc + lp))) KB
    by default (lc=3, lp=0), state_size = 16 KB.

LZMA properties (5 bytes) format
    Offset Size  Description
      0     1    lc, lp and pb in encoded form.
      1     4    dictSize (little endian).

LzmaCompress
------------

outPropsSize -
     In:  the pointer to the size of outProps buffer; *outPropsSize = LZMA_PROPS_SIZE = 5.
     Out: the pointer to the size of written properties in outProps buffer; *outPropsSize = LZMA_PROPS_SIZE = 5.

  LZMA Encoder will use defult values for any parameter, if it is
  -1  for any from: level, loc, lp, pb, fb, numThreads
   0  for dictSize

level - compression level: 0 <= level <= 9;

  level dictSize algo  fb
    0:    16 KB   0    32
    1:    64 KB   0    32
    2:   256 KB   0    32
    3:     1 MB   0    32
    4:     4 MB   0    32
    5:    16 MB   1    32
    6:    32 MB   1    32
    7+:   64 MB   1    64

  The default value for "level" is 5.

  algo = 0 means fast method
  algo = 1 means normal method

dictSize - The dictionary size in bytes. The maximum value is
        128 MB = (1 << 27) bytes for 32-bit version
          1 GB = (1 << 30) bytes for 64-bit version
     The default value is 16 MB = (1 << 24) bytes.
     It's recommended to use the dictionary that is larger than 4 KB and
     that can be calculated as (1 << N) or (3 << N) sizes.

lc - The number of literal context bits (high bits of previous literal).
     It can be in the range from 0 to 8. The default value is 3.
     Sometimes lc=4 gives the gain for big files.

lp - The number of literal pos bits (low bits of current position for literals).
     It can be in the range from 0 to 4. The default value is 0.
     The lp switch is intended for periodical data when the period is equal to 2^lp.
     For example, for 32-bit (4 bytes) periodical data you can use lp=2. Often it's
     better to set lc=0, if you change lp switch.

pb - The number of pos bits (low bits of current position).
     It can be in the range from 0 to 4. The default value is 2.
     The pb switch is intended for periodical data when the period is equal 2^pb.

fb - Word size (the number of fast bytes).
     It can be in the range from 5 to 273. The default value is 32.
     Usually, a big number gives a little bit better compression ratio and
     slower compression process.

numThreads - The number of thereads. 1 or 2. The default value is 2.
     Fast mode (algo = 0) can use only 1 thread.

Out:
  destLen  - processed output size
Returns:
  SZ_OK               - OK
  SZ_ERROR_MEM        - Memory allocation error
  SZ_ERROR_PARAM      - Incorrect paramater
  SZ_ERROR_OUTPUT_EOF - output buffer overflow
  SZ_ERROR_THREAD     - errors in multithreading functions (only for Mt version)
*/
// MY_STDAPI LzmaCompress(unsigned char *dest, size_t *destLen, const unsigned char *src, size_t srcLen,
//   unsigned char *outProps, size_t *outPropsSize, /* *outPropsSize must be = 5 */
// int level,      /* 0 <= level <= 9, default = 5 */
// unsigned dictSize,  /* default = (1 << 24) */
// int lc,        /* 0 <= lc <= 8, default = 3  */
// int lp,        /* 0 <= lp <= 4, default = 0  */
// int pb,        /* 0 <= pb <= 4, default = 2  */
// int fb,        /* 5 <= fb <= 273, default = 32 */
// int numThreads /* 1 or 2, default = 2 */
// );
func lzmaCompress(dst []byte, dstLen *uint64, src []byte, srcLen uint64,
	outProps []byte, outPropsSize *uint64,
	level, dictSize, lc, lp, pb, fb, numThreads uint32) (int, error) {
	r1, _, err := compress.Call(
		uintptr(unsafe.Pointer(&dst[0])), uintptr(unsafe.Pointer(dstLen)),
		uintptr(unsafe.Pointer(&src[0])), uintptr(srcLen),
		uintptr(unsafe.Pointer(&outProps[0])),
		uintptr(unsafe.Pointer(outPropsSize)),
		uintptr(level), uintptr(dictSize), uintptr(lc), uintptr(lp),
		uintptr(pb), uintptr(fb), uintptr(numThreads))
	if sysErr, ok := err.(syscall.Errno); !ok || sysErr != 0 {
		return 0, err
	}
	return int(r1), nil
}

/*
LzmaUncompress
--------------
In:
  dest     - output data
  destLen  - output data size
  src      - input data
  srcLen   - input data size
Out:
  destLen  - processed output size
  srcLen   - processed input size
Returns:
  SZ_OK                - OK
  SZ_ERROR_DATA        - Data error
  SZ_ERROR_MEM         - Memory allocation arror
  SZ_ERROR_UNSUPPORTED - Unsupported properties
  SZ_ERROR_INPUT_EOF   - it needs more bytes in input buffer (src)

MY_STDAPI LzmaUncompress(unsigned char *dest, size_t *destLen, const unsigned char *src, SizeT *srcLen,
  const unsigned char *props, size_t propsSize);
*/
func lzmaUncompress(dst []byte, dstLen *uint64, src []byte, srcLen *uint64,
	props []byte, propsSize uint64) (int, error) {
	r1, _, err := unCompress.Call(
		uintptr(unsafe.Pointer(&dst[0])), uintptr(unsafe.Pointer(dstLen)),
		uintptr(unsafe.Pointer(&src[0])), uintptr(unsafe.Pointer(srcLen)),
		uintptr(unsafe.Pointer(&props[0])), uintptr(propsSize))
	if sysErr, ok := err.(syscall.Errno); !ok || sysErr != 0 {
		return 0, err
	}
	return int(r1), nil
}

func LzmaCompress(r io.Reader, w io.Writer) error {
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	var (
		srcLen    = uint64(len(src))
		propsSize = LzmaPropsSize
		dstLen    = srcLen + srcLen/3 + 128
		dst       = make([]byte, sizeHeader+LzmaPropsSize+dstLen)
	)
	copy(dst, setLenBytes(srcLen)) // 保存源文件长度
	// 同时也保存outProps
	ret, err := lzmaCompress(dst[sizeHeader+LzmaPropsSize:], &dstLen, src, srcLen,
		dst[sizeHeader:sizeHeader+LzmaPropsSize], &propsSize, 9, 1<<24, 3, 0, 2, 32, 2)
	if err != nil {
		return err
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

func LzmaUnCompress(r io.Reader, w io.Writer) error {
	src, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}
	var (
		srcLen = uint64(len(src)) - sizeHeader - LzmaPropsSize // 去掉头部
		dstLen = getLenBytes(src[:sizeHeader])                 // 读取源文件大小
		dst    = make([]byte, dstLen)                          // 申请资源
	)
	// 使用r中读到的props传递参数
	ret, err := lzmaUncompress(dst, &dstLen, src[sizeHeader+LzmaPropsSize:],
		&srcLen, src[sizeHeader:sizeHeader+LzmaPropsSize], LzmaPropsSize)
	if err != nil {
		return err
	}
	if ret != SzOk {
		return fmt.Errorf("lzmaCompress ret: %d", ret)
	}
	_, err = w.Write(dst[:dstLen])
	return err
}

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
