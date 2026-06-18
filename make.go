package xcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// Check if the given path is a folder and exists
func (x *XCli) isFolderExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

// Check if the given path is a file and exists
func (x *XCli) isFileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return !info.IsDir(), nil
}

// Create a file from a stub
func (x *XCli) CreateFileFromStub(stubPath, destPath string, replacements map[string]string) error {
	return x.CreateFileFromStubExtended(stubPath, destPath, replacements, false)
}

// Create a file from a stub with extended option to process modules
func (x *XCli) CreateFileFromStubExtended(stubPath, destPath string, replacements map[string]string, isModule bool) error {
	dir := filepath.Dir(destPath)

	// Create directories if needed
	if exists, err := x.isFolderExists(dir); err != nil {
		return fmt.Errorf("failed to check folder existence: %w", err)
	} else if !exists {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create folder: %w", err)
		}
	}

	if exists, err := x.isFileExists(destPath); err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	} else if exists {
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

	// Post-process Go files
	if strings.HasSuffix(destPath, ".go") {
		text = x.postProcessGoContent(text, isModule)
	}

	if err := os.WriteFile(destPath, []byte(text), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

func (x *XCli) getProjectModuleName() string {
	content, err := os.ReadFile("go.mod")
	if err != nil {
		return "github.com/imohamedsheta/xapp" // fallback
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module"))
		}
	}
	return "github.com/imohamedsheta/xapp"
}

func (x *XCli) postProcessGoContent(content string, isModule bool) string {
	projModule := x.getProjectModuleName()

	// 1. Replace imohamedsheta/xframe imports with target paths
	content = strings.ReplaceAll(content, "{{HANDLER_PATH}}", getEnvPath("XCLI_IMPORT_HANDLER", projModule+"/app/http/handler"))
	content = strings.ReplaceAll(content, "{{MODELS_PATH}}", getEnvPath("XCLI_IMPORT_MODELS", projModule+"/app/models"))
	content = strings.ReplaceAll(content, "{{ENUMS_PATH}}", getEnvPath("XCLI_IMPORT_ENUMS", projModule+"/app/domain/enums"))
	content = strings.ReplaceAll(content, "{{UTILS_PATH}}", getEnvPath("XCLI_IMPORT_UTILS", projModule+"/app/domain/utils"))
	content = strings.ReplaceAll(content, "{{X_APP_PATH}}", getEnvPath("XCLI_IMPORT_X_APP", projModule+"/app/x"))
	content = strings.ReplaceAll(content, "{{PKG_PATH}}", getEnvPath("XCLI_IMPORT_PKG", projModule+"/pkg"))
	content = strings.ReplaceAll(content, "{{INERTIA_PATH}}", getEnvPath("XCLI_IMPORT_INERTIA", projModule+"/pkg/inertia"))
	content = strings.ReplaceAll(content, "{{REQUESTS_PATH}}", getEnvPath("XCLI_IMPORT_REQUESTS", projModule+"/app/http/requests"))

	return content
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

func (x *XCli) generateGoFileModule(stubPath, moduleName, suffix, entityName string) (string, string, error) {
	entityNameLower := strings.ToLower(entityName)
	moduleNameLower := strings.ToLower(moduleName)
	isMain := entityNameLower == moduleNameLower || pluralize(toSnakeCase(entityNameLower)) == moduleNameLower || toSnakeCase(entityNameLower) == moduleNameLower

	var fileName string
	var structName string

	switch suffix {
	case "Handler":
		if isMain {
			fileName = "handler.go"
			structName = "Handler"
		} else {
			fileName = toSnakeCase(entityName) + "_handler.go"
			structName = toCamel(entityName) + "Handler"
		}
	case "Repository":
		if isMain {
			fileName = "repository.go"
		} else {
			fileName = toSnakeCase(entityName) + "_repository.go"
		}
		structName = toCamel(entityName) + "Repository"
	case "Request":
		if isMain {
			fileName = "requests.go"
		} else {
			fileName = toSnakeCase(entityName) + "_requests.go"
		}
		structName = toCamel(entityName) + "Request"
	default:
		fileName = toSnakeCase(entityName) + "_" + strings.ToLower(suffix) + ".go"
		structName = toCamel(entityName) + suffix
	}

	dir := filepath.Join(getEnvPath("XCLI_PATH_MODULES", "app/modules"), moduleName)
	filePath := filepath.Join(dir, fileName)

	replacements := map[string]string{
		"package": moduleName,
		"struct":  structName,
	}

	if err := x.CreateFileFromStubExtended(stubPath, filePath, replacements, true); err != nil {
		return "", "", err
	}

	// Check if routes.go exists, if not generate it
	routesPath := filepath.Join(dir, "routes.go")
	if exists, err := x.isFileExists(routesPath); err != nil {
		return "", "", fmt.Errorf("failed to check file existence: %w", err)
	} else if !exists {
		routesReplacements := map[string]string{
			"package": moduleName,
		}
		if err := x.CreateFileFromStubExtended("stubs/routes.stub", routesPath, routesReplacements, true); err != nil {
			return "", "", err
		}
	}

	return filePath, moduleName, nil
}

func (x *XCli) makeModuleCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "make:module [name]",
		Short: "Create a new application module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				return fmt.Errorf("name is required")
			}

			moduleName := pluralize(toSnakeCase(name))
			fmt.Printf("📦 Scaffolding new module: %s\n", moduleName)

			// Generate the routes.go file via create from stub
			dir := filepath.Join(getEnvPath("XCLI_PATH_MODULES", "app/modules"), moduleName)
			routesPath := filepath.Join(dir, "routes.go")
			replacements := map[string]string{
				"package": moduleName,
			}
			if err := x.CreateFileFromStubExtended("stubs/routes.stub", routesPath, replacements, true); err != nil {
				return err
			}
			fmt.Printf("✅ Created routes: %s\n", routesPath)

			// Generate basic handler
			_, _, err := x.generateGoFileModule("stubs/handler.stub", moduleName, "Handler", name)
			if err != nil {
				return fmt.Errorf("failed to generate handler: %w", err)
			}

			// Generate basic repository
			_, _, err = x.generateGoFileModule("stubs/repository.stub", moduleName, "Repository", name)
			if err != nil {
				return fmt.Errorf("failed to generate repository: %w", err)
			}

			// Generate basic requests
			_, _, err = x.generateGoFileModule("stubs/request.stub", moduleName, "Request", name)
			if err != nil {
				return fmt.Errorf("failed to generate request: %w", err)
			}

			fmt.Printf("\n🎉 Module %s scaffolded successfully!\n", moduleName)
			return nil
		},
	}

	return cmd
}
