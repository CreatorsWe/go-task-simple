package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"sync"
)

type ExecutionStatus int

const (
	Success ExecutionStatus = iota
	Fail
	Timeout
)

type TaskCommand struct {
	task_name string
}

func (tc TaskCommand) run() {
	config := parse_config()

	if config == nil {
		msg := error_msg_without_run("empty configure")
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	task, ok := config.Tasks[tc.task_name]
	if !ok {
		msg := error_msg_without_run("undefined task %s", tc.task_name)
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	// 并行执行
	if task.Parallel {
		_run_with_parallel(tc.task_name, &task)
	} else {
		_run_in_order(tc.task_name, &task)
	}

}

// command 执行失败， 程序错误码也返回 0
func _run_with_parallel(task_name string, task *Task) {
	var print_mutex sync.Mutex
	var wg sync.WaitGroup
	for step, cmd := range task.Cmds {
		wg.Go(func() {
			wrap_output, status := _run_single_command(cmd, task_name, task.Silent, step, task.Timeout)
			// 执行失败，但是 continue, 错误信息必须打印
			switch status {
			case Success:
				if !task.Silent {
					_mutex_print(wrap_output, &print_mutex)
				}
			case Timeout:
				_mutex_print(wrap_output, &print_mutex)
			case Fail:
				if task.Continue {
					_mutex_print(wrap_output, &print_mutex)
				} else {
					_mutex_print(wrap_output, &print_mutex)
					os.Exit(0)
				}
			}
		})
	}

	wg.Wait()
}

func _mutex_print(msg string, mt *sync.Mutex) {
	mt.Lock()
	fmt.Println(msg)
	mt.Unlock()
}

func _run_in_order(task_name string, task *Task) {
	for step, cmd := range task.Cmds {
		wrap_output, status := _run_single_command(cmd, task_name, task.Silent, step, task.Timeout)
		// 执行失败，但是 continue, 错误信息必须打印
		switch status {
		case Success:
			if !task.Silent {
				fmt.Println(wrap_output)
			}
		case Timeout:
			fmt.Println(wrap_output)
		case Fail:
			if task.Continue {
				fmt.Println(wrap_output)
				continue
			} else {
				fmt.Println(wrap_output)
				os.Exit(0)
			}
		}
	}
}

// 执行一条命令原语，返回拼接后的输出信息和错误
func _run_single_command(command string, task_name string, silent bool, step int, timeout uint64) (string, ExecutionStatus) {
	// output 是 stdout ; err 是 stderr
	// Command 执行命令返回的 err 接口底层类型就是 ExitCode 类型
	// CommandContext 带超时机制，但是 CommandContext 仅取消子进程，会导致孤儿进程，需要手动设置子进程为进程组 leader，才能取消进程组
	// timeout 会自动终止整个进程组，被 timeout 终止返回 124 exitcode
	// sh -c 只有第一个非选项参数是“命令字符串”, 它之后的所有参数作为位置参数（$0、$1、$2...）传给这个脚本
	var cmd *exec.Cmd
	if timeout != 0 {
		cmd = exec.Command("timeout", strconv.FormatUint(timeout, 10), "sh", "-c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	// err interface 是 ExitErr 类型
	output, err := cmd.Output()

	output_s := string(output)

	// err == nil 执行成功
	if err == nil {
		msg := msg_task_success(output_s, task_name, step)
		return msg, Success
	}
	// err != nil, 看命令是否执行(exec.ExitErr 是否是退出错误)，执行了看错误码
	if exiterr, ok := errors.AsType[*exec.ExitError](err); ok {
		exitcode := exiterr.ExitCode()
		switch exitcode {
		case 0: // 执行成功返回
			msg := msg_task_success(output_s, task_name, step)
			return msg, Success
		case 124: // timeout
			msg := msg_timeout(output_s, task_name, step, timeout, silent)
			return msg, Timeout
		default:
			msg := msg_task_fail(output_s, task_name, step, string(exiterr.Stderr), exitcode, silent)
			return msg, Fail
		}

	} else { // 没有执行
		msg := error_msg_without_run("task don't execute: %s", err)
		return msg, Fail
	}
}
