package commands

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "ostrakon",
	Short: "Ostrakon is a secure vault for storing secrets in a private GitHub repository",
	Long: `Ostrakon is a secure vault for storing secrets in a private GitHub repository.

In ancient Athens, an ostrakon was a piece of pottery used as a scrap for
everyday writing, tax receipts, and secret voting. It was the ancient world's
equivalent of a Gist or a pastebin.

Ostrakon provides client-side encryption, ensuring your secrets are encrypted
before they leave your computer.`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(initCmd)
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(getCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(shredCmd)
}
