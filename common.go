package main

type SubCommand interface {
	run()
}

const VERSION = "v0.0.5"
const TEMPLATE_NAME = "Taskfile.yml"
