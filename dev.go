package xcli

import (
	"fmt"
	"os"
	"sync"

	"github.com/spf13/cobra"
)

func (x *XCli) scanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "scan",
		Short: "Run golangci-lint",
		Run: func(cmd *cobra.Command, args []string) {
			if err := x.runInCli(cmd.Context(), "golangci-lint", []string{"run"}); err != nil {
				fmt.Println(Red+"❌ Goose make migration failed:", err, Reset)
				os.Exit(1)
			}
		},
	}
}

func (x *XCli) lintCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "lint",
		Short: "Lint and auto-fix with gofmt",
		Run: func(cmd *cobra.Command, args []string) {

			if err := x.runInCli(cmd.Context(), "golangci-lint", []string{"run"}); err != nil {
				fmt.Println(Red+"❌ Goose make migration failed:", err, Reset)
				os.Exit(1)
			}

			if err := x.runInCli(cmd.Context(), "gofmt", []string{"-w", "."}); err != nil {
				fmt.Println(Red+"❌ Goose make migration failed:", err, Reset)
				os.Exit(1)
			}

		},
	}
}

func (x *XCli) devCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "dev",
		Short: "Run development using air",
		Run: func(cmd *cobra.Command, args []string) {
			ctx := cmd.Context()

			var wg sync.WaitGroup
			wg.Add(2)

			go func() {
				defer wg.Done()
				if err := x.runInCli(ctx, "air", []string{"run"}); err != nil {
					fmt.Println(Red+"air stopped:", err, Reset)
				}
			}()

			go func() {
				defer wg.Done()
				if err := x.runInCli(ctx, "npm", []string{"run", "dev"}); err != nil {
					fmt.Println(Red+"npm stopped:", err, Reset)
				}
			}()

			wg.Wait()
		},
	}
}
