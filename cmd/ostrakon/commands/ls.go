package commands

import (
	"fmt"
	"os"
	"slices"
	"strings"
	"text/tabwriter"

	"github.com/PapaDanielVi/ostrakon/pkg/config"
	"github.com/PapaDanielVi/ostrakon/pkg/github"
	"github.com/PapaDanielVi/ostrakon/pkg/vault"
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls [<path>] [--tree]",
	Short: "List all secrets in the vault",
	Long: `List all secrets stored in the vault using the Git Trees API for efficiency.
Supports path filtering and tree view for hierarchical organization.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runLs,
}

var (
	listProfile string
	treeView    bool
	searchTerm  string
)

func init() {
	lsCmd.Flags().StringVarP(&listProfile, "profile", "p", "", "Filter by profile/namespace (deprecated, use path instead)")
	lsCmd.Flags().BoolVarP(&treeView, "tree", "t", false, "Show secrets as a tree")
	lsCmd.Flags().StringVarP(&searchTerm, "search", "s", "", "Search for secrets by name pattern")
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

	// Determine path filter
	pathFilter := ""
	if len(args) > 0 {
		pathFilter = args[0]
	}

	// List files
	files, err := client.ListFiles(ctx)
	if err != nil {
		return fmt.Errorf("failed to list files: %w", err)
	}

	// Normalize path filter (ensure it has trailing slash for directory matching)
	normalizedPath := pathFilter
	if normalizedPath != "" && normalizedPath[len(normalizedPath)-1] != '/' {
		normalizedPath += "/"
	}

	// Filter by path or profile
	if pathFilter != "" {
		var filtered []vault.FileInfo
		for _, f := range files {
			// Match exact path or path prefix
			if f.Path == pathFilter || (len(f.Path) > len(normalizedPath) && f.Path[:len(normalizedPath)] == normalizedPath) {
				filtered = append(filtered, f)
			}
		}
		files = filtered
	} else if listProfile != "" {
		prefix := fmt.Sprintf("profiles/%s/", listProfile)
		var filtered []vault.FileInfo
		for _, f := range files {
			if len(f.Path) >= len(prefix) && f.Path[:len(prefix)] == prefix {
				filtered = append(filtered, f)
			}
		}
		files = filtered
	}

	// Filter by search term (fuzzy substring match)
	if searchTerm != "" {
		var filtered []vault.FileInfo
		for _, f := range files {
			if containsSubstring(f.Path, searchTerm) {
				filtered = append(filtered, f)
			}
		}
		files = filtered
	}

	// Print results
	if treeView {
		printTree(files)
	} else {
		writer := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
		fmt.Fprintln(writer, "NAME\tSIZE")
		for _, f := range files {
			fmt.Fprintf(writer, "%s\t%d\n", f.Path, f.Size)
		}
		if err := writer.Flush(); err != nil {
			return fmt.Errorf("failed to flush output: %w", err)
		}
	}

	fmt.Printf("\n%d secret(s) in vault\n", len(files))
	return nil
}

// printTree prints files in a tree structure.
func printTree(files []vault.FileInfo) {
	// Build tree structure
	tree := make(map[string][]string)
	roots := []string{}

	for _, f := range files {
		parts := splitPath(f.Path)
		if len(parts) == 1 {
			roots = append(roots, f.Path)
		} else {
			parent := parts[0]
			child := f.Path[len(parent)+1:]
			tree[parent] = append(tree[parent], child)
			if !contains(roots, parent) {
				roots = append(roots, parent)
			}
		}
	}

	// Print tree
	for _, root := range roots {
		if children, ok := tree[root]; ok {
			fmt.Printf("%s/\n", root)
			for _, child := range children {
				fmt.Printf("  └── %s\n", child)
			}
		} else {
			fmt.Printf("%s\n", root)
		}
	}
}

// splitPath splits a path into components.
func splitPath(path string) []string {
	var parts []string
	start := 0
	for i := range len(path) {
		if path[i] == '/' {
			parts = append(parts, path[start:i])
			start = i + 1
		}
	}
	if start <= len(path) {
		parts = append(parts, path[start:])
	}
	return parts
}

// contains checks if a string is in a slice.
func contains(slice []string, item string) bool {
	return slices.Contains(slice, item)
}

// containsSubstring checks if a string contains a substring (case-insensitive).
func containsSubstring(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
