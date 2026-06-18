package xcli

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/cobra"
)

func (x *XCli) makeModelCommand() *cobra.Command {
	stub := "stubs/model.stub"
	prefixDir := getEnvPath("XCLI_PATH_MODELS", "app/models")

	cmd := &cobra.Command{
		Use:   "make:model [name]",
		Short: "Create a new model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			if name == "" {
				return fmt.Errorf("name is required")
			}

			fileName := filepath.Join(prefixDir, toSnakeCase(name)+".go")
			tableName := tableNameFromModel(name)
			data := map[string]string{
				"MODEL_NAME": name,
				"TABLE_NAME": tableName,
			}

			err := x.CreateFileFromStub(stub, fileName, data)
			if err != nil {
				return err
			}

			fmt.Printf("Success: created %s\n", fileName)
			return nil
		},
	}

	return cmd
}

func toSnakeCase(str string) string {
	re := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := re.ReplaceAllString(str, "${1}_${2}")
	return strings.ToLower(snake)
}

func pluralize(word string) string {
	if strings.HasSuffix(word, "y") && !strings.HasSuffix(word, "ay") && !strings.HasSuffix(word, "ey") && !strings.HasSuffix(word, "iy") && !strings.HasSuffix(word, "oy") && !strings.HasSuffix(word, "uy") {
		return word[:len(word)-1] + "ies"
	}
	if strings.HasSuffix(word, "s") || strings.HasSuffix(word, "sh") || strings.HasSuffix(word, "ch") || strings.HasSuffix(word, "x") || strings.HasSuffix(word, "z") {
		return word + "es"
	}
	return word + "s"
}

func tableNameFromModel(modelName string) string {
	return pluralize(toSnakeCase(modelName))
}
