package main

import (
	"fmt"
	"os"
	"strings"
)

func parse_flag() SubCommand {
	// 手动解析参数
	args := os.Args[1:]
	if len(args) != 1 {
		err_msg := error_msg_without_run("unknown flag")
		fmt.Fprintln(os.Stderr, err_msg)
		os.Exit(1)
	}

	subcommand := args[0]

	if !strings.HasPrefix(subcommand, "-") {
		return TaskCommand{task_name: subcommand}
	}

	switch subcommand {
	case "--help", "-h":
		return HelpCommand{}
	case "--list", "-l":
		return ListCommand{}
	case "--version", "-v":
		return VersionCommand{}
	case "--init", "-i":
		return InitCommand{}
	default:
		err_msg := error_msg_without_run("unknown flag")
		fmt.Fprintln(os.Stderr, err_msg)
		os.Exit(1)
	}
	return nil
}
