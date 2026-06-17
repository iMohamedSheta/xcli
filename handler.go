package xcli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (x *XCli) makeHandlerCommand() *cobra.Command {
	var withAction, withRepo, withRequest bool

	cmd := &cobra.Command{
		Use:   "make:handler [name]",
		Short: "Create a new handler",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			// always create handler
			if err := x.generateHandler(name); err != nil {
				return err
			}

			// cascade to extras
			if withAction {
				if err := x.generateAction(name); err != nil {
					return err
				}
			}
			if withRepo {
				if err := x.generateRepo(name); err != nil {
					return err
				}
			}
			if withRequest {
				if err := x.generateRequest(name); err != nil {
					return err
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&withAction, "action", "a", false, "Also create action")
	cmd.Flags().BoolVarP(&withRepo, "repo", "r", false, "Also create repo")
	cmd.Flags().BoolVarP(&withRequest, "request", "R", false, "Also create request")

	return cmd
}

func (x *XCli) makeActionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "make:action [name]",
		Short: "Create a new action",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return x.generateAction(args[0])
		},
	}
}

func (x *XCli) makeRepoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "make:repo [name]",
		Short: "Create a new repo",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return x.generateRepo(args[0])
		},
	}
}

func (x *XCli) makeRequestCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "make:req [name]",
		Short: "Create a new request",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return x.generateRequest(args[0])
		},
	}
}

func (x *XCli) generateHandler(name string) error {
	file, pkg, err := x.generateGoFile("stubs/handler.stub", "app/api/handlers", "Handler", name, "handlers")
	if err != nil {
		return err
	}
	fmt.Printf("✅ Created handler: %s (package %s)\n", file, pkg)
	return nil
}

func (x *XCli) generateAction(name string) error {
	file, pkg, err := x.generateGoFile("stubs/action.stub", "app/api/actions", "Action", name, "actions")
	if err != nil {
		return err
	}
	fmt.Printf("✅ Created action: %s (package %s)\n", file, pkg)
	return nil
}

func (x *XCli) generateRepo(name string) error {
	file, pkg, err := x.generateGoFile("stubs/repository.stub", "app/api/repository", "Repository", name, "repository")
	if err != nil {
		return err
	}
	fmt.Printf("✅ Created repo: %s (package %s)\n", file, pkg)
	return nil
}

func (x *XCli) generateRequest(name string) error {
	file, pkg, err := x.generateGoFile("stubs/request.stub", "app/api/requests", "Request", name, "requests")
	if err != nil {
		return err
	}
	fmt.Printf("✅ Created request: %s (package %s)\n", file, pkg)
	return nil
}

func (x *XCli) generateNotification(name string) error {
	file, pkg, err := x.generateGoFile("stubs/notification.stub", "app/shared/notifications", "Notification", name, "notifications")
	if err != nil {
		return err
	}
	fmt.Printf("✅ Created repo: %s (package %s)\n", file, pkg)
	return nil
}
