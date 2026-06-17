package xcli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// readStub reads a stub file, checking the local directory first, then falling back to embedded stubs
func (x *XCli) readStub(stubPath string) ([]byte, error) {
	if _, err := os.Stat(stubPath); err == nil {
		return os.ReadFile(stubPath)
	}
	return stubsFS.ReadFile(stubPath)
}

func (x *XCli) stubPublishCommand() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "stub:publish",
		Short: "Publish all embedded stubs to the local 'stubs' directory for customization",
		RunE: func(cmd *cobra.Command, args []string) error {
			localDir := "stubs"

			// Create local stubs folder if it doesn't exist
			if err := os.MkdirAll(localDir, 0755); err != nil {
				return fmt.Errorf("failed to create 'stubs' directory: %w", err)
			}

			// Read all files from embedded stubs directory
			entries, err := stubsFS.ReadDir("stubs")
			if err != nil {
				return fmt.Errorf("failed to read embedded stubs directory: %w", err)
			}

			for _, entry := range entries {
				if entry.IsDir() {
					continue
				}

				stubName := entry.Name()
				embedPath := "stubs/" + stubName
				destPath := filepath.Join(localDir, stubName)

				// Check if file already exists locally
				if _, err := os.Stat(destPath); err == nil && !force {
					fmt.Printf("⏭ Skipped (already exists): %s\n", destPath)
					continue
				}

				// Read embedded stub file content
				content, err := stubsFS.ReadFile(embedPath)
				if err != nil {
					return fmt.Errorf("failed to read embedded stub %s: %w", stubName, err)
				}

				// Write to local stubs directory
				if err := os.WriteFile(destPath, content, 0644); err != nil {
					return fmt.Errorf("failed to write stub %s to %s: %w", stubName, destPath, err)
				}

				fmt.Printf("✅ Published: %s\n", destPath)
			}

			fmt.Println("\n🎉 All stubs published successfully!")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing local stubs")

	return cmd
}
