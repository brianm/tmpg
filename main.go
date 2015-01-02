package main

import (
	"gilog"
	"io/ioutil"
	"os"
	//"os/exec"
)

func main() {
	data_dir, err := ioutil.TempDir(os.TempDir(), "tmpg.")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(data_dir)

	println(data_dir)
}
