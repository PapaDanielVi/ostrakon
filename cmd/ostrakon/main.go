// Package main is the entry point for the Ostrakon CLI application. It initializes and executes the command-line interface defined in the commands package.
package main

import (
	"github.com/PapaDanielVi/ostrakon/cmd/ostrakon/commands"
)

func main() {
	commands.Execute()
}
