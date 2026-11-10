package main

type SubCommand interface {
	run()
}

const VERSION = "v0.0.6"
const TEMPLATE_NAME = "Taskfile.yml"
