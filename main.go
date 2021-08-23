package main

import "github.com/go-git/go-git/v5" // with go modules enabled (GO111MODULE=on or outside GOPATH)

func main() {
	r, err := git.PlainOpen(".")
}
