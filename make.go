package xcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// Check if the given path is a folder and exists
func (x *XCli) IsFolderExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}

		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	return info.IsDir()
}

// Check if the given path is a file and exists
func (x *XCli) IsFileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}

		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	return !info.IsDir()
}

// Create a file from a stub
func (x *XCli) CreateFileFromStub(stubPath, destPath string, replacements map[string]string) error {
	dir := filepath.Dir(destPath)

	// Create directories if needed
	if !x.IsFolderExists(dir) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create folder: %w", err)
		}
	}

	if x.IsFileExists(destPath) {
		return fmt.Errorf("file %s already exists", destPath)
	}

	content, err := x.readStub(stubPath)
	if err != nil {
		return fmt.Errorf("failed to read stub: %w", err)
	}

	text := string(content)
	for key, val := range replacements {
		placeholder := fmt.Sprintf("{{%s}}", key)
		text = strings.ReplaceAll(text, placeholder, val)
	}

	if err := os.WriteFile(destPath, []byte(text), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (x *XCli) makeVueComponentCommand() *cobra.Command {
	stub := "stubs/vue_component.stub"
	prefixDir := "resources/js/"
	var isView bool

	cmd := &cobra.Command{
		Use:   "make:vue [name]",
		Short: "Create a new vue component",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				return fmt.Errorf("name is required")
			}

			targetDir := "components"
			if isView {
				targetDir = "pages"
			}

			fileName := filepath.Join(prefixDir, targetDir, name+".vue")

			err := x.CreateFileFromStub(stub, fileName, nil)
			if err != nil {
				return err
			}

			fmt.Printf("Success: created %s\n", fileName)
			return nil
		},
	}

	cmd.Flags().BoolVarP(&isView, "view", "v", false, "Create a view component in 'pages' folder")
	return cmd
}

func (x *XCli) generateGoFile(stubPath, prefixDir, suffix, name, defaultPkg string) (string, string, error) {
	parts := strings.Split(name, "/")
	structName := toCamel(parts[len(parts)-1]) + suffix

	// package rule
	packageName := defaultPkg
	if len(parts) > 1 {
		packageName = parts[0]
	}

	// target path
	dir := filepath.Join(append([]string{prefixDir}, parts[:len(parts)-1]...)...)
	filePath := filepath.Join(dir, parts[len(parts)-1]+"_"+strings.ToLower(suffix)+".go")

	// replacements
	replacements := map[string]string{
		"package": packageName,
		"struct":  structName,
	}

	if err := x.CreateFileFromStub(stubPath, filePath, replacements); err != nil {
		return "", "", err
	}

	return filePath, packageName, nil
}
