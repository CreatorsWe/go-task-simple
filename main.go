package main

func main() {
	subcommand := parse_flag()

	if subcommand == nil {
		panic("unexpected return value nil from parse_flag")
	}

	subcommand.run()
}
