package main

import (
	"flag"
	"testing"
)

func TestHelpCommand(t *testing.T) {
	var hc HelpCommand

	hc.run()

	flag.Usage()
}

func TestInitCommand(t *testing.T) {
	var ic InitCommand

	ic.run()
}

func TestListCommand(t *testing.T) {
	var lc ListCommand

	lc.run()
}

func TestTaskCommand(t *testing.T) {
	var tc = TaskCommand{task_name: "build"}

	tc.run()
}

func TestMain(m *testing.M) {
	m.Run()
}
