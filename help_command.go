package main

import (
	"fmt"
)

type HelpCommand struct{}

func (hc HelpCommand) run() {
	fmt.Println("[Usage] task <task-name>  // registered task in Taskfile.yml")
	fmt.Println("        task <option>     // options: --list, --help, --version, --init")
}
