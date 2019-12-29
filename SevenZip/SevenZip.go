package SevenZip

import (
	"os/exec"
)

var path7za = ""

const app7za = "7za"

func Command(arg ...string) *exec.Cmd {
	return command(arg...)
}

func Get7zaPath() string {
	return path7za
}
