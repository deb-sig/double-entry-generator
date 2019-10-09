package main

import (
	"log"

	"github.com/gaocegege/double-entry-generator/pkg/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	err := doc.GenMarkdownTree(cmd.RCmd, "./doc")
	if err != nil {
		log.Fatal(err)
	}
}
