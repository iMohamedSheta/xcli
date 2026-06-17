package xcli

import (
	"github.com/spf13/cobra"
)

func (x *XCli) makeNotificationCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "make:notification [name]",
		Short: "Create a new notification",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return x.generateNotification(args[0])
		},
	}
}
