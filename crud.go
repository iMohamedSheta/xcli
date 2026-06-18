package xcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func (x *XCli) makeCrudCommand() *cobra.Command {
	var skipVue bool
	var skipRepo bool
	var module string

	cmd := &cobra.Command{
		Use:   "make:crud [name]",
		Short: "Create full CRUD operations in a module (request, handler, action, repository, vue index)",
		Long: `Create complete CRUD operations for an entity including:
- Request file (Filters, Create, Update)
- Handler file (Index, Create, Update, Delete)
- Action file (Create, Update, Delete)
- Repository file (Paginate, Create, Update, FindById, Delete)
- Vue Index page (Table with filters, pagination, search)`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				return fmt.Errorf("name is required")
			}

			return x.generateCrud(name, skipVue, skipRepo, module)
		},
	}

	cmd.Flags().BoolVar(&skipVue, "skip-vue", false, "Skip Vue Index page generation")
	cmd.Flags().BoolVar(&skipRepo, "skip-repo", false, "Skip Repository generation")
	cmd.Flags().StringVarP(&module, "module", "m", "", "Specify target module name (e.g. users)")

	return cmd
}

func (x *XCli) generateCrud(name string, skipVue, skipRepo bool, module string) error {
	entityName := toCamel(name)
	entityLower := strings.ToLower(name)
	entityPlural := pluralize(entityLower)
	tableName := tableNameFromModel(entityName)
	routeName := entityPlural

	if module == "" {
		module = pluralize(toSnakeCase(name))
	}

	// Generate replacements map
	replacements := map[string]string{
		"ENTITY":           entityName,
		"entity":           entityLower,
		"entities":         entityPlural,
		"table_name":       tableName,
		"route_name":       routeName,
		"entity_ar":        "العنصر",  // Default Arabic, can be customized
		"entity_ar_plural": "العناصر", // Default Arabic plural
		"package":          module,
	}

	// Determine directories and file names (Module-only mode)
	moduleLower := strings.ToLower(module)
	isMain := entityLower == moduleLower || pluralize(toSnakeCase(entityLower)) == moduleLower || toSnakeCase(entityLower) == moduleLower

	modulesDir := getEnvPath("XCLI_PATH_MODULES", "app/modules")
	reqDir := filepath.Join(modulesDir, module)
	actDir := filepath.Join(modulesDir, module)
	repoDir := filepath.Join(modulesDir, module)
	handDir := filepath.Join(modulesDir, module)

	var reqFile, actFile, repoFile, handFile string
	if isMain {
		reqFile = "requests.go"
		actFile = "actions.go"
		repoFile = "repository.go"
		handFile = "handler.go"
	} else {
		reqFile = toSnakeCase(name) + "_requests.go"
		actFile = toSnakeCase(name) + "_actions.go"
		repoFile = toSnakeCase(name) + "_repository.go"
		handFile = toSnakeCase(name) + "_handler.go"
	}

	vueDir := filepath.Join(getEnvPath("XCLI_PATH_VUE_PAGES", "resources/js/pages"), module)
	vueFile := "Index.vue"

	// 1. Generate Request file
	fmt.Println("📝 Generating Request file...")
	var actualReqFile string
	var err error
	actualReqFile, _, err = x.generateCrudFile(
		"stubs/crud_request.stub",
		reqDir,
		reqFile,
		replacements,
		entityLower,
		true,
	)
	if err != nil {
		return fmt.Errorf("failed to generate request: %w", err)
	}
	fmt.Printf("✅ Created request: %s\n", actualReqFile)

	// 2. Generate Action file
	fmt.Println("⚡ Generating Action file...")
	var actualActFile string
	actualActFile, _, err = x.generateCrudFile(
		"stubs/crud_action.stub",
		actDir,
		actFile,
		replacements,
		entityLower,
		true,
	)
	if err != nil {
		return fmt.Errorf("failed to generate action: %w", err)
	}
	fmt.Printf("✅ Created action: %s\n", actualActFile)

	// 3. Generate Repository file (if not skipped)
	if !skipRepo {
		fmt.Println("🗄️  Generating Repository file...")
		var actualRepoFile string
		actualRepoFile, _, err = x.generateCrudFile(
			"stubs/crud_repository.stub",
			repoDir,
			repoFile,
			replacements,
			entityLower,
			true,
		)
		if err != nil {
			return fmt.Errorf("failed to generate repository: %w", err)
		}
		fmt.Printf("✅ Created repository: %s\n", actualRepoFile)
	}

	// 4. Generate Handler file
	fmt.Println("🎯 Generating Handler file...")
	var actualHandFile string
	actualHandFile, _, err = x.generateCrudFile(
		"stubs/crud_handler.stub",
		handDir,
		handFile,
		replacements,
		entityLower,
		true,
	)
	if err != nil {
		return fmt.Errorf("failed to generate handler: %w", err)
	}
	fmt.Printf("✅ Created handler: %s\n", actualHandFile)

	// 5. Generate Vue Index page (if not skipped)
	if !skipVue {
		fmt.Println("🎨 Generating Vue Index page...")
		var actualVueFile string
		actualVueFile, _, err = x.generateCrudFile(
			"stubs/crud_vue_index.stub",
			vueDir,
			vueFile,
			replacements,
			entityLower,
			true,
		)
		if err != nil {
			return fmt.Errorf("failed to generate vue index: %w", err)
		}
		fmt.Printf("✅ Created vue index: %s\n", actualVueFile)
	}

	// check and generate routes.go
	routesPath := filepath.Join(handDir, "routes.go")
	if exists, err := x.isFileExists(routesPath); err != nil {
		return fmt.Errorf("failed to check file existence: %w", err)
	} else if !exists {
		routesReplacements := map[string]string{
			"package": module,
		}
		if err := x.CreateFileFromStubExtended("stubs/routes.stub", routesPath, routesReplacements, true); err != nil {
			return fmt.Errorf("failed to generate routes: %w", err)
		}
	}

	fmt.Println("\n🎉 CRUD generation complete!")
	fmt.Println("\n📋 Next steps:")
	fmt.Println("1. Update the model file if needed")
	fmt.Println("2. Register module routes in app/providers/routes.go or register the module in registers.go")
	if !skipVue {
		fmt.Println("3. Add Vue component / components in resources/js/components/app/" + module + "/")
		fmt.Println("4. Add " + entityName + " type to resources/js/types/index.ts")
	}

	return nil
}

func (x *XCli) generateCrudFile(stubPath, prefixDir, fileName string, replacements map[string]string, entityLower string, isModule bool) (string, string, error) {
	// Create directory if needed
	dir := prefixDir
	if exists, err := x.isFolderExists(dir); err != nil {
		return "", "", fmt.Errorf("failed to check folder existence: %w", err)
	} else if !exists {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", "", fmt.Errorf("failed to create folder: %w", err)
		}
	}

	filePath := filepath.Join(dir, fileName)

	if exists, err := x.isFileExists(filePath); err != nil {
		return "", "", fmt.Errorf("failed to check file existence: %w", err)
	} else if exists {
		return "", "", fmt.Errorf("file %s already exists", filePath)
	}

	content, err := x.readStub(stubPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read stub: %w", err)
	}

	text := string(content)
	// First pass: replace all placeholders
	for key, val := range replacements {
		placeholder := fmt.Sprintf("{{%s}}", key)
		text = strings.ReplaceAll(text, placeholder, val)
	}
	// Second pass: handle Vue template interpolation placeholders
	if strings.Contains(text, "entity_name_placeholder") {
		text = strings.ReplaceAll(text, "entity_name_placeholder", fmt.Sprintf("{{ %s.name }}", entityLower))
	}
	if strings.Contains(text, "entity_created_at_placeholder") {
		text = strings.ReplaceAll(text, "entity_created_at_placeholder", fmt.Sprintf("{{ formatDate(%s.created_at!) }}", entityLower))
	}

	// Post-process Go files
	if strings.HasSuffix(filePath, ".go") {
		text = x.postProcessGoContent(text, isModule)
	}

	if err := os.WriteFile(filePath, []byte(text), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, "", nil
}
