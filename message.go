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

// 任务成功执行返回的信息
func msg_task_success(output, task_name string, step int) string {
	success_prompt := fmt.Sprintf("\x1b[32m[%s]step %d: execute success\x1b[0m", task_name, step)
	output_trim := strings.TrimSpace(output)
	if output_trim == "" {
		return fmt.Sprint(success_prompt)
	}
	return fmt.Sprintf("%s\n%s", output_trim, success_prompt)
}

// 任务执行失败返回的信息
func msg_task_fail(output, task_name string, step int, err_msg string, exitcode int, silent bool) string {
	fail_prompt := fmt.Sprintf("\x1b[31m[%s]step %d: execute fial. exitcode status: %d\n%s\x1b[0m", task_name, step, exitcode, err_msg)
	var wrap_msg string
	output_trim := strings.TrimSpace(output)
	if silent || output_trim == "" {
		wrap_msg = fail_prompt
	} else {
		wrap_msg = fmt.Sprintf("%s\n%s", output_trim, fail_prompt)
	}
	return wrap_msg
}

// timeout 包装信息
func msg_timeout(output, task_name string, step int, timeout uint64, silent bool) string {
	timeout_prompt := fmt.Sprintf("\x1b[33m[%s]step %d: timetout in %d seconds\x1b[0m", task_name, step, timeout)
	var wrap_msg string
	output_trim := strings.TrimSpace(output)
	if silent || output_trim == "" {
		wrap_msg = timeout_prompt
	} else {
		wrap_msg = fmt.Sprintf("%s\n%s", output_trim, timeout_prompt)
	}
	return wrap_msg
}
