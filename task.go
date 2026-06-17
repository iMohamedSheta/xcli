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
	file, pkg, err := x.generateGoFile(
		"stubs/task.stub",
		"app/shared/tasks", // output folder
		"Task",             // struct suffix (empty)
		name,
		"tasks", // default package name
	)

	if err != nil {
		return err
	}

	fmt.Printf("🟩 Created task: %s (package %s)\n", file, pkg)
	return nil
}
