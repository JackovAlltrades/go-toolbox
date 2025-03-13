package main

import (
	"fmt"

	"github.com/JackovAlltrades/toolbox"
)

func main() {
	var tools toolbox.Tools

	s := tools.RandomString(7)
	fmt.Println("Random string:", s)
}
