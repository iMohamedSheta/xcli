package xcli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func (x *XCli) makeCrudCommand() *cobra.Command {
	var handlerPackage string
	var skipVue bool
	var skipRepo bool

	cmd := &cobra.Command{
		Use:   "make:crud [name]",
		Short: "Create full CRUD operations (request, handler, action, repository, vue index)",
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

			return x.generateCrud(name, handlerPackage, skipVue, skipRepo)
		},
	}

	cmd.Flags().StringVarP(&handlerPackage, "package", "p", "admin", "Handler package name (admin, manager, etc.)")
	cmd.Flags().BoolVar(&skipVue, "skip-vue", false, "Skip Vue Index page generation")
	cmd.Flags().BoolVar(&skipRepo, "skip-repo", false, "Skip Repository generation")

	return cmd
}

func (x *XCli) generateCrud(name, handlerPackage string, skipVue, skipRepo bool) error {
	entityName := toCamel(name)
	entityLower := strings.ToLower(name)
	entityPlural := pluralize(entityLower)
	tableName := tableNameFromModel(entityName)
	routeName := entityPlural

	// Generate replacements map
	replacements := map[string]string{
		"ENTITY":           entityName,
		"entity":           entityLower,
		"entities":         entityPlural,
		"table_name":       tableName,
		"route_name":       routeName,
		"entity_ar":        "العنصر",  // Default Arabic, can be customized
		"entity_ar_plural": "العناصر", // Default Arabic plural
		"package":          handlerPackage,
	}

	// 1. Generate Request file
	fmt.Println("📝 Generating Request file...")
	requestFile, _, err := x.generateCrudFile(
		"stubs/crud_request.stub",
		"app/api/requests",
		name+"_request.go",
		replacements,
		entityLower,
	)
	if err != nil {
		return fmt.Errorf("failed to generate request: %w", err)
	}
	fmt.Printf("✅ Created request: %s\n", requestFile)

	// 2. Generate Action file
	fmt.Println("⚡ Generating Action file...")
	actionFile, _, err := x.generateCrudFile(
		"stubs/crud_action.stub",
		"app/api/actions",
		name+"_action.go",
		replacements,
		entityLower,
	)
	if err != nil {
		return fmt.Errorf("failed to generate action: %w", err)
	}
	fmt.Printf("✅ Created action: %s\n", actionFile)

	// 3. Generate Repository file (if not skipped)
	if !skipRepo {
		fmt.Println("🗄️  Generating Repository file...")
		repoFile, _, err := x.generateCrudFile(
			"stubs/crud_repository.stub",
			"app/api/repository",
			name+"_repository.go",
			replacements,
			entityLower,
		)
		if err != nil {
			return fmt.Errorf("failed to generate repository: %w", err)
		}
		fmt.Printf("✅ Created repository: %s\n", repoFile)
	}

	// 4. Generate Handler file
	fmt.Println("🎯 Generating Handler file...")
	handlerDir := filepath.Join("app/api/handlers", handlerPackage)
	handlerFile, _, err := x.generateCrudFile(
		"stubs/crud_handler.stub",
		handlerDir,
		name+"_handler.go",
		replacements,
		entityLower,
	)
	if err != nil {
		return fmt.Errorf("failed to generate handler: %w", err)
	}
	fmt.Printf("✅ Created handler: %s\n", handlerFile)

	// 5. Generate Vue Index page (if not skipped)
	if !skipVue {
		fmt.Println("🎨 Generating Vue Index page...")
		vueDir := filepath.Join("resources/js/Pages", entityName)
		vueFile, _, err := x.generateCrudFile(
			"stubs/crud_vue_index.stub",
			vueDir,
			"Index.vue",
			replacements,
			entityLower,
		)
		if err != nil {
			return fmt.Errorf("failed to generate vue index: %w", err)
		}
		fmt.Printf("✅ Created vue index: %s\n", vueFile)
	}

	fmt.Println("\n🎉 CRUD generation complete!")
	fmt.Println("\n📋 Next steps:")
	fmt.Println("1. Update the model file if needed")
	fmt.Println("2. Register handler in app/providers/handlers.go")
	fmt.Println("3. Add routes in app/api/routes/" + handlerPackage + ".go")
	if !skipVue {
		fmt.Println("4. Create Add" + entityName + " component in resources/js/components/app/" + entityName + "/")
		fmt.Println("5. Add " + entityName + " type to resources/js/types/index.ts")
	}

	return nil
}

func (x *XCli) generateCrudFile(stubPath, prefixDir, fileName string, replacements map[string]string, entityLower string) (string, string, error) {
	// Create directory if needed
	dir := prefixDir
	if !x.IsFolderExists(dir) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return "", "", fmt.Errorf("failed to create folder: %w", err)
		}
	}

	filePath := filepath.Join(dir, fileName)

	if x.IsFileExists(filePath) {
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
	// Replace entity_name_placeholder with actual Vue interpolation
	if strings.Contains(text, "entity_name_placeholder") {
		text = strings.ReplaceAll(text, "entity_name_placeholder", fmt.Sprintf("{{ %s.name }}", entityLower))
	}
	if strings.Contains(text, "entity_created_at_placeholder") {
		text = strings.ReplaceAll(text, "entity_created_at_placeholder", fmt.Sprintf("{{ formatDate(%s.created_at!) }}", entityLower))
	}

	if err := os.WriteFile(filePath, []byte(text), 0644); err != nil {
		return "", "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, "", nil
}
