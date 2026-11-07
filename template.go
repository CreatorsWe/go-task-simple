package main

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/fatih/color"
)

const TEMPLATE = `[task.build]
cmds = [
    ""
]
desc = ""
silent = false

[task.run]
cmds = [
    ""
]
desc = ""
silent = false`

const TEMPLATE_NAME = "Taskfile.toml"

func init_template() error {
	wd, err := os.Getwd()
	if err != nil {
		msg := color.RedString("[init-tempalte failed] %s", err)
		return errors.New(msg)
	}

	template_path := filepath.Join(wd, TEMPLATE_NAME)

	if _, err := os.Stat(template_path); !errors.Is(err, fs.ErrNotExist) {
		scanner := bufio.NewScanner(os.Stdin)
		fmt.Printf("Taskfile.toml already exists, do you want to replace it? (y/n): ")
		if scanner.Scan() {
			judgement := scanner.Text()
			if judgement == "n" {
				return nil
			}
			if judgement != "n" && judgement != "y" {
				msg := color.RedString("[init-template failed] unexpected input: %s", judgement)
				return errors.New(msg)
			}
		}
	}

	if err := os.WriteFile(template_path, []byte(TEMPLATE), 0666); err != nil {
		msg := color.RedString("[init-tempalte failed] %s", err)
		return errors.New(msg)
	}

	return nil
}
