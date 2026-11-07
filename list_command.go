package main

import (
	"fmt"
	"os"
)

type ListCommand struct{}

func (lc ListCommand) run() {
	config := parse_config()

	if config == nil {
		msg := error_msg_without_run("empty configure")
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	for task_name, info := range config.Tasks {
		item := _format_task_info(task_name, info.Desc)
		fmt.Println(item)
	}
}

func _format_task_info(task_name, desc string) string {
	return fmt.Sprintf("\x1b[32m%-1s\x1b[0m     %s", task_name, desc)
}
