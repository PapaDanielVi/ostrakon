package commands

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/github"
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
	ctx := context.Background()

	// Get PAT from keychain
	pat, err := config.GetPAT()
	if err != nil {
		return fmt.Errorf("not initialized: %w", err)
	}

	// Create GitHub client
	client, err := github.NewClient(pat)
	if err != nil {
		return err
	}

	// Check if vault exists
	exists, err := client.VaultExists(ctx)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("vault does not exist. Run 'ostrakon init' first")
	}

	// List files
	files, err := client.ListFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	// Filter by profile if specified
	if listProfile != "" {
		prefix := fmt.Sprintf("profiles/%s/", listProfile)
		var filtered []github.FileInfo
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
	writer.Flush()

	fmt.Printf("\n%d secret(s) in vault\n", len(files))
	return nil
}
