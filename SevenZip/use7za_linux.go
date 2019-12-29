package SevenZip

import (
	"os/exec"
)

func init() {
	if lp, err := exec.LookPath(app7za); err == nil {
		path7za = lp
	} else {
		panic("7za can not find")
	}
}

func command(arg ...string) *exec.Cmd {
	return &exec.Cmd{
		Path: path7za,
		Args: append([]string{app7za}, arg...),
	}
}
