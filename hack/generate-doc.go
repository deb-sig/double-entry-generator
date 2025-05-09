package main

import (
	"log"

	"github.com/deb-sig/double-entry-generator/v2/pkg/cmd"
	"github.com/spf13/cobra/doc"
)

func main() {
	err := doc.GenMarkdownTree(cmd.RCmd, "./doc")
	if err != nil {
		log.Fatal(err)
	}
}
