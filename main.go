/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package main

import (
	"log"

	"github.com/myml/linglong-tools/cmd"
)

func main() {
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	cmd.Execute()
}
