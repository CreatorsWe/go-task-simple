package main

import "testing"

func TestErrMsgWithoutRun(t *testing.T) {
	wrap_msg := error_msg_without_run("parse error.")
	t.Log(wrap_msg)
}

func TestMsgTaskSuccess(t *testing.T) {
	output := `
		line 1.
		line 2.

	`
	wrap_msg := msg_task_success(output, "test", 1)
	t.Log(wrap_msg)
}

func TestMsgTaskFail(t *testing.T) {
	output := `
		line 1.
		line 2.

	`
	wrap_msg := msg_task_fail(output, "test", 1, "", -1, false)
	t.Log("[no error message, no silent]")
	t.Log(wrap_msg)

	wrap_msg = msg_task_fail(output, "test", 2, "cann't open file.", -1, false)
	t.Log("[no silent]")
	t.Log(wrap_msg)

	wrap_msg = msg_task_fail(output, "test", 2, "cann't open file.", -1, true)
	t.Log("[silent]")
	t.Log(wrap_msg)
}
