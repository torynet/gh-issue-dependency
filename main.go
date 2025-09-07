package main

import (
	"os"

	"github.com/torynet/gh-issue-dependency/cmd"
)

func main() {
	code := cmd.Execute()
	os.Exit(code)
}
