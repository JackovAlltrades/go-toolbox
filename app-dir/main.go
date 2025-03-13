package main

import (
	"github.com/JackovAlltrades/go-toolbox"
)
	
func main() {
	var tools toolbox.Tools

	tools.CreateDirIfNotExist("./test-dir")
}