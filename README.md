# addFileToGo 说明
该可执行程序用来将文件添加到go代码中,使go可执行程序能携带文件.  
go run addFileToGo.go a.txt b.txt c.txt  
会产生resources.go文件,只需要添加到项目目录按照如下用法就可以还原文件.  
go run main.go resources.go即可释放几个文件到指定目录.  
```go
package main

func main() {
  // 如下方法,当文件不存在或大小不正确,则会重新释放文件
  WriteBytesToFile("D:\\a.txt", "a.txt")
  WriteBytesToFile("D:\\b.txt", "b.txt")
  WriteBytesToFile("D:\\c.txt", "c.txt")
}
```

# 执行7za命令
```go
package main

import (
  "github.com/jan-bar/addFileToGo/SevenZip"
)

func main() {
  // windows不存在7za.exe,自动创建7za.exe到环境变量,下次启动直接使用环境变量的7za.exe
  // windows下只传一个参数,则直接使用该命令行,不会拆分成[]string类型
  // 如果命令行里面有引号括起来的空格参数是没有问题的,这种写法只支持windows
  SevenZip.Command("a xx.7z a.log b.txt").Output()

  // Linux下7za命令不存在直接报panic错误
  // 下面这种则是默认exec.Command那种方式,windows和linux下均相同
  SevenZip.Command("a", "xx.7z", "a.log", "b.txt").Output()

  // 之所以windows和Linux有所不同,看源码吧,我只是比较懒,不想处理Linux下的逻辑.
}
```

# 两个结合起来用
1. 使用7za命令将多个文件压缩成一个文件,使用addFileToGo将文件写入go代码
2. 使用go代码释放文件,使用7za命令行解压即可

# 使用lzma库
1. 提供了接口通过windows的dll调用lzmalib库
2. 提供了接口通过windows的cgo调用lzmalib库
3. 修改LzmaUtil.c文件提供通用接口,windows和Linux都可以支持

# 总结
1. 之所以7za没有做成复杂的封装好的方法,是因为提供基础方法使用灵活,还有懒.
2. 写这个也主要是平时我写的程序需要携带文件,避免使用者那边没有必要文件.
