package main

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

func run_task(task_name string, config *Config) error {
	err := replace_variables(config)
	if err != nil {
		return err
	}

	task, exist := config.Tasks[task_name]
	if !exist {
		msg := color.RedString("undefined task %s", task_name)
		return errors.New(msg)
	}

	// 执行单挑命令
	if len(task.Cmds) == 1 {
		output, err := exec.Command("sh", "-c", task.Cmds[0]).Output()

		err = execute_status(strings.TrimSpace(string(output)), err, &task)

		if err != nil {
			return err
		}

		msg := color.GreenString("[%s] execute success", task.Cmds[0])

		fmt.Println(msg)

		return nil
	}

	// 执行多条命令, 每条命令执行完成要提示
	for _, command := range task.Cmds {
		output, err := exec.Command("sh", "-c", command).Output()

		err = execute_status(strings.TrimSpace(string(output)), err, &task)
		if err != nil {
			return err
		}

		msg := color.GreenString("[%s] execute success", command)
		fmt.Println(msg)
	}

	return nil
}

func execute_status(out string, err error, task *Task) error {
	if err != nil { // 执行失败
		if exitErr, ok := errors.AsType[*exec.ExitError](err); ok {
			if !task.Silent && out != "" {
				// output 信息
				fmt.Println(out)
			}
			// error 信息
			msg := color.RedString("[run failed] failed to run task %s: exit status: %d", task.Cmds[0], exitErr.ExitCode())
			return errors.New(msg)
		} else if errors.Is(err, exec.ErrNotFound) { // "" 执行空命令会抛出 exec.ErrNotFound 错误
			return nil
		} else { //命令没跑
			msg := color.RedString("unexecutale for unknown reason")
			return errors.New(msg)
		}
	} else { // 执行成功
		if !task.Silent && out != "" {
			fmt.Println(out)
		}
		return nil
	}
}
