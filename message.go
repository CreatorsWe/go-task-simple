// 提供 待打印信息 的包装函数
package main

import (
	"fmt"
	"strings"
)

// 非任务执行时返回的程序错误信息
func error_msg_without_run(content string, args ...any) string {
	content = fmt.Sprintf(content, args...)
	return fmt.Sprintf("\x1b[31m%s\x1b[0m", content)
}

// timeout 包装信息
func msg_timeout(task_name string, timeout uint64) string {
	timeout_prompt := fmt.Sprintf("\x1b[31m[%s]timetout in %d seconds\x1b[0m", task_name, timeout)
	return timeout_prompt
}

// fail prompt
func msg_task_fail(task_name string, step int, err_msg string, exitcode int) string {
	err_msg_trim := strings.TrimSpace(err_msg)
	wrap_msg := fmt.Sprintf("\x1b[31m[%s:step %d]%s exitcode status: %d\x1b[0m", task_name, step, err_msg_trim, exitcode)
	return wrap_msg
}

func msg_command_no_execution(task_name string, step int) string {
	wrap_msg := fmt.Sprintf("\x1b[31m[%s:step %d]no execution\x1b[0m", task_name, step)
	return wrap_msg
}
