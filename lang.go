package xcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func (x *XCli) makeLangFilesCommand() *cobra.Command {
	stub := "stubs/lang.stub"
	prefixDir := "resources/js/lang"

	cmd := &cobra.Command{
		Use:   "make:lang [name]",
		Short: "Create new lang files in all lang folders",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				return fmt.Errorf("name is required")
			}

			// List all subfolders in resources/js/lang
			entries, err := os.ReadDir(prefixDir)
			if err != nil {
				return err
			}

			for _, e := range entries {
				if !e.IsDir() {
					continue
				}
				langDir := filepath.Join(prefixDir, e.Name())
				fileName := filepath.Join(langDir, name+".ts")

				// Create file only if not exists
				if _, err := os.Stat(fileName); os.IsNotExist(err) {
					err := x.CreateFileFromStub(stub, fileName, nil)
					if err != nil {
						return err
					}
					fmt.Printf("✅ Created %s\n", fileName)
				} else {
					fmt.Printf("⏭ Skipped (already exists): %s\n", fileName)
				}

				// Update index.ts for this lang folder
				indexFile := filepath.Join(langDir, "index.ts")
				if _, err := os.Stat(indexFile); err == nil {
					err = updateLangIndex(indexFile, name)
					if err != nil {
						return err
					}
					fmt.Printf("🔄 Updated %s\n", indexFile)
				} else {
					fmt.Printf("⚠ Skipped updating, no index.ts in %s\n", langDir)
				}
			}
			return nil
		},
	}

	return cmd
}

// updateLangIndex adds import + export to index.ts if not exists
func updateLangIndex(indexFile, name string) error {
	content, _ := os.ReadFile(indexFile)
	lines := strings.Split(string(content), "\n")

	importLine := fmt.Sprintf("import %s from './%s';", name, name)

	// Check if already exists
	for _, l := range lines {
		if strings.Contains(l, importLine) {
			return nil
		}
	}

	// Insert import before export default
	newLines := []string{}
	added := false
	for _, l := range lines {
		if strings.HasPrefix(strings.TrimSpace(l), "export default") && !added {
			newLines = append(newLines, importLine)
			added = true
		}
		newLines = append(newLines, l)
	}

	// Fix export object
	for i, l := range newLines {
		if strings.HasPrefix(strings.TrimSpace(l), "export default") {
			if !strings.Contains(l, name) {
				newLines[i] = strings.Replace(l, "{", "{\n  "+name+",", 1)
			}
		}
	}

	return os.WriteFile(indexFile, []byte(strings.Join(newLines, "\n")), 0644)
}
