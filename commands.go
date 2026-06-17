package xcli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func (x *XCli) versionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show CLI version",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("XCli CLI version ", version)
		},
	}
}

func (x *XCli) helpCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "help",
		Short: "Show help",
		Run: func(cmd *cobra.Command, args []string) {
			x.help()
		},
	}
}
