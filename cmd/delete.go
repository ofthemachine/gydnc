package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"gydnc/config"
	"gydnc/core/content"
	"gydnc/model"
	"gydnc/storage/localfs"

	"github.com/spf13/cobra"
)

var forceDelete bool

var deleteCmd = &cobra.Command{
	Use:   "delete <alias1> [alias2 ...]",
	Short: "Delete one or more guidance entities by alias (from all backends)",
	Long: `Deletes one or more guidance entities by alias. Searches all configured backends for each alias.
Requires confirmation unless --force is specified.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		aliases := args
		cfg := config.Get()
		var toDelete []model.Entity
		var notFound []string

		// Discover all matching entities across all backends
		for _, alias := range aliases {
			found := false
			for backendName, backendConfig := range cfg.StorageBackends {
				if backendConfig.Type != "localfs" {
					continue
				}
				if backendConfig.LocalFS == nil || backendConfig.LocalFS.Path == "" {
					continue
				}
				store, err := localfs.NewStore(*backendConfig.LocalFS)
				if err != nil {
					continue
				}
				if err := store.Init(nil); err != nil {
					continue
				}
				contentBytes, meta, err := store.Read(alias)
				if err == nil && meta != nil {
					parsed, _ := content.ParseG6E(contentBytes)
					entity := model.Entity{
						Alias:          alias,
						SourceBackend:  backendName,
						Title:          parsed.Title,
						Description:    parsed.Description,
						Tags:           parsed.Tags,
						CustomMetadata: meta,
						Body:           parsed.Body,
					}
					cid, _ := parsed.GetContentID()
					entity.CID = cid
					toDelete = append(toDelete, entity)
					found = true
				}
			}
			if !found {
				notFound = append(notFound, alias)
			}
		}

		if len(toDelete) == 0 {
			fmt.Println("No matching entities found to delete.")
			if len(notFound) > 0 {
				fmt.Printf("Not found: %s\n", strings.Join(notFound, ", "))
			}
			return nil
		}

		// Print summary and confirm
		if !forceDelete {
			fmt.Println("The following entities will be deleted:")
			for _, e := range toDelete {
				path := ""
				if p, ok := e.CustomMetadata["path"].(string); ok {
					path = p
				}
				title := e.Title
				if title == "" {
					title = "(no title)"
				}
				fmt.Printf("- %s (backend: %s, path: %s, title: %s)\n", e.Alias, e.SourceBackend, path, title)
			}
			fmt.Print("Proceed with deletion? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			resp, _ := reader.ReadString('\n')
			resp = strings.TrimSpace(strings.ToLower(resp))
			if resp != "y" && resp != "yes" {
				fmt.Println("Aborted.")
				return nil
			}
		}

		// Perform deletions
		var deleted, failed []string
		for _, e := range toDelete {
			beConfig := cfg.StorageBackends[e.SourceBackend]
			store, err := localfs.NewStore(*beConfig.LocalFS)
			if err != nil {
				failed = append(failed, fmt.Sprintf("%s (backend: %s): %v", e.Alias, e.SourceBackend, err))
				continue
			}
			if err := store.Init(nil); err != nil {
				failed = append(failed, fmt.Sprintf("%s (backend: %s): %v", e.Alias, e.SourceBackend, err))
				continue
			}
			path := ""
			if p, ok := e.CustomMetadata["path"].(string); ok {
				path = p
			} else {
				path = e.Alias
			}
			if err := store.Delete(path); err != nil {
				failed = append(failed, fmt.Sprintf("%s (backend: %s): %v", e.Alias, e.SourceBackend, err))
			} else {
				deleted = append(deleted, fmt.Sprintf("%s (backend: %s)", e.Alias, e.SourceBackend))
			}
		}

		if len(deleted) > 0 {
			fmt.Println("Deleted:")
			for _, d := range deleted {
				fmt.Printf("- %s\n", d)
			}
		}
		if len(failed) > 0 {
			fmt.Println("Failed to delete:")
			for _, f := range failed {
				fmt.Printf("- %s\n", f)
			}
		}
		if len(notFound) > 0 {
			fmt.Printf("Not found: %s\n", strings.Join(notFound, ", "))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)
	deleteCmd.Flags().BoolVarP(&forceDelete, "force", "f", false, "Delete without confirmation")
}
