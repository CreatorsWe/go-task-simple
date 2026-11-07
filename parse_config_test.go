package main

import "testing"

func TestParseConfig(t *testing.T) {
	config := parse_config()

	_debug(*config)
}
