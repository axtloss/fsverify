package core

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/axtloss/fsverify/verify/config"
)

func WarnUser() {
	fmt.Println(config.FbWarnLoc)
	fmt.Println(config.BVGLoc)

	sizeCMD := exec.Command("./getscreensize")
	out, err := sizeCMD.Output()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(out))
	height, err := strconv.Atoi(string(out))
	var scaleFactor float64
	if err != nil {
		scaleFactor = 1
	}

	scaleFactor = 600.0 / float64(height)
	fmt.Println(scaleFactor)

	warnCMD := exec.Command(config.FbWarnLoc, config.BVGLoc, fmt.Sprintf("%f", scaleFactor))
	_, err = warnCMD.Output()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	os.Exit(1)
}
