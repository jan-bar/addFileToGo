package SevenZip

import (
	"os/exec"
)

var Path7za = ""

const app7za = "7za"

func Command(arg ...string) *exec.Cmd {
	return command(arg...)
}
