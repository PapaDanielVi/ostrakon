package commands

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/github"
	"github.com/PapaDanielVi/ostrakon/pkg/vault"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls [--profile profile]",
	Short: "List all secrets in the vault",
	Long:  `List all secrets stored in the vault using the Git Trees API for efficiency.`,
	RunE:  runLs,
}

var listProfile string

func init() {
	lsCmd.Flags().StringVarP(&listProfile, "profile", "p", "", "Filter by profile/namespace")
}

func runLs(cmd *cobra.Command, args []string) error {
	ctx := cmd.Context()

	// Get token and repo info from keychain
	token, err := config.GetToken()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	owner, err := config.GetRepoOwner()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	repoName, err := config.GetRepoName()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	// Create GitHub client
	client, err := github.NewClient(token, owner, repoName)
	if err != nil {
		return err
	}

	// List files
	files, err := client.ListFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	// Filter by profile if specified
	if listProfile != "" {
		prefix := fmt.Sprintf("profiles/%s/", listProfile)
		var filtered []vault.FileInfo
		for _, f := range files {
			if len(f.Path) > len(prefix) && f.Path[:len(prefix)] == prefix {
				f.Path = f.Path[len(prefix):]
				filtered = append(filtered, f)
			}
		}
		files = filtered
	}

	// Print table
	writer := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
	fmt.Fprintln(writer, "NAME\tSIZE")
	for _, f := range files {
		fmt.Fprintf(writer, "%s\t%d\n", f.Path, f.Size)
	}
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush output: %w", err)
	}

	fmt.Printf("\n%d secret(s) in vault\n", len(files))
	return nil
}
