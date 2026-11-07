package main

import (
	"flag"
	"fmt"

	"github.com/fatih/color"
)

func main() {
	is_init := flag.Bool("init", false, "create a Taskfile.toml template in current directory.")

	flag.Parse()

	flag.Usage = func() {
		flag.CommandLine.PrintDefaults()
		fmt.Printf("task --init or task [task_name]\n")
	}

	pos_args := flag.Args()

	if *is_init && len(pos_args) == 0 {
		err := init_template()
		if err != nil {
			fmt.Println(err)
			return
		}
	} else if !*is_init && len(pos_args) == 1 {
		// 解析任务名称并执行
		config, err := parse()
		if err != nil {
			fmt.Println(err)
			return
		}

		err = run_task(pos_args[0], config)
		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		msg := color.RedString("error command. please use `task --help`")
		fmt.Println(msg)
		return
	}

}
