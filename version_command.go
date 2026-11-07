package main

import "fmt"

type VersionCommand struct{}

func (vc VersionCommand) run() {
	fmt.Println(VERSION)
}
