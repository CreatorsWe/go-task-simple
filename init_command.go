package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

type InitCommand struct{}

func (ic InitCommand) run() {
	const TEMPLATE = `# yaml-language-server: $schema=none
tasks:
  build:
    desc:
    cmds:
      -
  run:
    desc:
    cmds:
      -
  clean:
    desc:
    cmds:
      -
`

	wd, err := os.Getwd()
	if err != nil {
		msg := error_msg_without_run(err.Error())
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}

	template_path := filepath.Join(wd, TEMPLATE_NAME)

	if _, err := os.Stat(template_path); !errors.Is(err, fs.ErrNotExist) {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Printf("Taskfile.toml already exists, do you want to replace it? (y/n): ")
		if scanner.Scan() {
			judgement := scanner.Text()
			if judgement == "n" {
				os.Exit(0)
			}
			if judgement != "n" && judgement != "y" {
				msg := error_msg_without_run("unexpected input: %s", judgement)
				fmt.Println(msg)
				os.Exit(0)
			}
		}
	}

	if err := os.WriteFile(template_path, []byte(TEMPLATE), 0666); err != nil {
		msg := error_msg_without_run(err.Error())
		fmt.Fprintln(os.Stderr, msg)
		os.Exit(1)
	}
}
