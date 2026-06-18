package xcli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (x *XCli) makeTaskCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "make:task [name]",
		Short: "Create a new Asynq task + handler stub",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return x.generateTask(args[0])
		},
	}
}

func (x *XCli) generateTask(name string) error {
	dir := getEnvPath("XCLI_PATH_TASKS", "app/domain/tasks")
	file, pkg, err := x.generateGoFile(
		"stubs/task.stub",
		dir,
		"Task",
		name,
		"tasks",
	)

	if err != nil {
		return err
	}

	fmt.Printf("🟩 Created task: %s (package %s)\n", file, pkg)
	return nil
}
