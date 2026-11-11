package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"
)

type ExecutionStatus int

const (
	Success ExecutionStatus = iota
	Fail
	NoExecution
)

type TaskCommand struct {
	task_name string
	flag_var  map[string]string
}

func (tc TaskCommand) run() {
	config := parse_config(tc.flag_var)

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

	if task.Timeout == 0 {
		if task.Parallel {
			_run_with_parallel(tc.task_name, &task)
		} else {
			_run_in_order(tc.task_name, &task)
		}
		return
	}

	return_chan := make(chan struct{})

	go func() {
		if task.Parallel {
			_run_with_parallel(tc.task_name, &task)
		} else {
			_run_in_order(tc.task_name, &task)
		}
		if return_chan != nil {
			close(return_chan)
		}
	}()

	select {
	case <-return_chan:
		os.Exit(1)
	case <-time.After(time.Duration(time.Second * time.Duration(task.Timeout))):
		wrap_msg := msg_timeout(tc.task_name, task.Timeout)
		fmt.Fprintln(os.Stderr, wrap_msg)
		close(return_chan)
		os.Exit(1)
	}
}

type Exiterror struct {
	err_msg  string
	exitcode int
}

// command 执行失败， 程序错误码也返回 0
func _run_with_parallel(task_name string, task *Task) {
	var wg sync.WaitGroup

	for step, cmd := range task.Cmds {
		if cmd == "" {
			continue
		}
		wg.Go(func() {
			exiterror, status := _run_single_command(cmd, task.Silent, task.Shell)
			// 执行失败，但是 continue, 错误信息必须打印
			switch status {
			case Success:
				return
			case Fail:
				fail_msg := msg_task_fail(task_name, step, exiterror.err_msg, exiterror.exitcode)
				fmt.Fprintln(os.Stderr, fail_msg)
			case NoExecution:
				wrap_msg := msg_command_no_execution(task_name, step)
				fmt.Fprintln(os.Stderr, wrap_msg)
			}

			if !task.Continue {
				os.Exit(1)
			}
		})
	}

	wg.Wait()
}

func _run_in_order(task_name string, task *Task) {
	for step, cmd := range task.Cmds {
		if cmd == "" {
			continue
		}
		exiterror, status := _run_single_command(cmd, task.Silent, task.Shell)
		// 执行失败，但是 continue, 错误信息必须打印
		switch status {
		case Success:
			continue
		case Fail:
			fail_msg := msg_task_fail(task_name, step, exiterror.err_msg, exiterror.exitcode)
			fmt.Fprintln(os.Stderr, fail_msg)
		case NoExecution:
			wrap_msg := msg_command_no_execution(task_name, step)
			fmt.Fprintln(os.Stderr, wrap_msg)
		}
		if !task.Continue {
			os.Exit(1)
		}
	}
}

// 执行一条命令原语，返回拼接后的输出信息和错误
func _run_single_command(command string, silent bool, shell bool) (*Exiterror, ExecutionStatus) {
	// output 是 stdout ; err 是 stderr
	// Command 执行命令返回的 err 接口底层类型就是 ExitCode 类型
	// CommandContext 带超时机制，但是 CommandContext 仅取消子进程，会导致孤儿进程，需要手动设置子进程为进程组 leader，才能取消进程组
	// timeout 会自动终止整个进程组，被 timeout 终止返回 124 exitcode
	// sh -c 只有第一个非选项参数是“命令字符串”, 它之后的所有参数作为位置参数（$0、$1、$2...）传给这个脚本
	var cmd *exec.Cmd

	if shell {
		cmd = exec.Command("sh", "-c", command)
	} else {
		cmd = exec.Command(command)
	}

	// 直接将命令的标准输出重定向到终端
	// 重定向到给定的错误缓冲区

	if silent {
		cmd.Stdout = nil
	} else {
		cmd.Stdout = os.Stdout
	}
	var err_buffer bytes.Buffer
	cmd.Stderr = &err_buffer

	err := cmd.Run()

	// err == nil 执行成功
	if err == nil {
		return nil, Success
	}
	// err != nil, 看命令是否执行(exec.ExitErr 是否是退出错误)，执行了看错误码
	if exiterr, ok := errors.AsType[*exec.ExitError](err); ok {
		exitcode := exiterr.ExitCode()
		switch exitcode {
		case 0: // 执行成功返回
			return nil, Success
		default:
			exiterror := Exiterror{
				err_msg:  err_buffer.String(),
				exitcode: exitcode,
			}
			return &exiterror, Fail
		}

	} else { // 没有执行
		return nil, NoExecution
	}
}
