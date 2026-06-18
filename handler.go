package xcli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (x *XCli) makeHandlerCommand() *cobra.Command {
	var withAction, withRepo, withRequest bool
	var module string

	cmd := &cobra.Command{
		Use:   "make:handler [name]",
		Short: "Create a new handler in a module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// always create handler
			if err := x.generateHandler(name, module); err != nil {
				return err
			}

			// cascade to extras
			if withAction {
				if err := x.generateAction(name, module); err != nil {
					return err
				}
			}
			if withRepo {
				if err := x.generateRepo(name, module); err != nil {
					return err
				}
			}
			if withRequest {
				if err := x.generateRequest(name, module); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&withAction, "action", "a", false, "Also create action")
	cmd.Flags().BoolVarP(&withRepo, "repo", "r", false, "Also create repo")
	cmd.Flags().BoolVarP(&withRequest, "request", "R", false, "Also create request")
	cmd.Flags().StringVarP(&module, "module", "m", "", "Specify the target module name (e.g. users)")

	return cmd
}

func (x *XCli) makeActionCommand() *cobra.Command {
	var module string
	cmd := &cobra.Command{
		Use:   "make:action [name]",
		Short: "Create a new action in a module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return x.generateAction(args[0], module)
		},
	}
	cmd.Flags().StringVarP(&module, "module", "m", "", "Specify the target module name (e.g. users)")
	return cmd
}

func (x *XCli) makeRepoCommand() *cobra.Command {
	var module string
	cmd := &cobra.Command{
		Use:   "make:repo [name]",
		Short: "Create a new repo in a module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return x.generateRepo(args[0], module)
		},
	}
	cmd.Flags().StringVarP(&module, "module", "m", "", "Specify the target module name (e.g. users)")
	return cmd
}

func (x *XCli) makeRequestCommand() *cobra.Command {
	var module string
	cmd := &cobra.Command{
		Use:   "make:req [name]",
		Short: "Create a new request in a module",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return x.generateRequest(args[0], module)
		},
	}
	cmd.Flags().StringVarP(&module, "module", "m", "", "Specify the target module name (e.g. users)")
	return cmd
}

func (x *XCli) generateHandler(name string, module string) error {
	if module == "" {
		module = pluralize(toSnakeCase(name))
	}
	file, pkg, err := x.generateGoFileModule("stubs/handler.stub", module, "Handler", name)
	if err != nil {
		return err
	}
	fmt.Printf("✅ Created handler: %s (package %s)\n", file, pkg)
	return nil
}

func (x *XCli) generateAction(name string, module string) error {
	if module == "" {
		module = pluralize(toSnakeCase(name))
	}
	file, pkg, err := x.generateGoFileModule("stubs/action.stub", module, "Action", name)
	if err != nil {
		return err
	}
	fmt.Printf("✅ Created action: %s (package %s)\n", file, pkg)
	return nil
}

func (x *XCli) generateRepo(name string, module string) error {
	if module == "" {
		module = pluralize(toSnakeCase(name))
	}
	file, pkg, err := x.generateGoFileModule("stubs/repository.stub", module, "Repository", name)
	if err != nil {
		return err
	}
	fmt.Printf("✅ Created repo: %s (package %s)\n", file, pkg)
	return nil
}

func (x *XCli) generateRequest(name string, module string) error {
	if module == "" {
		module = pluralize(toSnakeCase(name))
	}
	file, pkg, err := x.generateGoFileModule("stubs/request.stub", module, "Request", name)
	if err != nil {
		return err
	}
	fmt.Printf("✅ Created request: %s (package %s)\n", file, pkg)
	return nil
}

func (x *XCli) generateNotification(name string) error {
	dir := getEnvPath("XCLI_PATH_NOTIFICATIONS", "app/domain/notifications")
	file, pkg, err := x.generateGoFile("stubs/notification.stub", dir, "Notification", name, "notifications")
	if err != nil {
		return err
	}
	fmt.Printf("✅ Created notification: %s (package %s)\n", file, pkg)
	return nil
}
