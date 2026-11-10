package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

func parse_flag() SubCommand {
	// 手动解析参数
	args := os.Args[1:]
	if len(args) < 1 {
		err_msg := error_msg_without_run("unknown flag")
		fmt.Fprintln(os.Stderr, err_msg)
		os.Exit(1)
	}

	subcommand := args[0]

	if !strings.HasPrefix(subcommand, "-") {
		var flag_vars map[string]string
		switch len(args) {
		case 1:
			return TaskCommand{task_name: subcommand, flag_var: flag_vars}
		case 3:
			if args[1] == "--var" || args[1] == "-var" {
				flag_vars = make(map[string]string)
				flag_vars = _parse_flag_variable(args[2])
				return TaskCommand{task_name: subcommand, flag_var: flag_vars}
			} else {
				err_msg := error_msg_without_run("unknown flag:%s", args[1])
				fmt.Fprintln(os.Stderr, err_msg)
				os.Exit(1)
			}
		default:
			err_msg := error_msg_without_run("unknown flag")
			fmt.Fprintln(os.Stderr, err_msg)
			os.Exit(1)
		}

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
		err_msg := error_msg_without_run("unknown flag:%s", subcommand)
		fmt.Fprintln(os.Stderr, err_msg)
		os.Exit(1)
	}
	return nil
}

// task build --var "time=4;path=./src fdsfdsf  ;"
func _parse_flag_variable(flag_vaiables string) map[string]string {
	if flag_vaiables == "" {
		return nil
	}

	result := make(map[string]string)

	vars := strings.Split(flag_vaiables, ";")

	// 正则提取
	regex := regexp.MustCompile(`(\w+)\s*=\s*(.*)`)

	for _, vari := range vars {
		if vari == "" {
			continue
		}
		match := regex.FindStringSubmatch(vari)
		if len(match) != 3 {
			msg := error_msg_without_run("flag varable must be key-value")
			fmt.Fprintln(os.Stderr, msg)
			os.Exit(1)
		}
		if match[2] == "" {
			continue
		}
		value := strings.TrimSpace(match[2])
		result[match[1]] = value
	}

	return result
}
